package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

// DefaultHub is the global WebSocket hub instance.
// Set via SetDefaultHub during server startup.
var DefaultHub *Hub

// SetDefaultHub sets the global hub instance.
func SetDefaultHub(h *Hub) {
	DefaultHub = h
}

// Event types pushed over WebSocket.
const (
	EventTransactionCreated = "transaction.created"
	EventStockAdjusted      = "stock.adjusted"
	EventStockTransferred   = "stock.transferred"
	EventStockLow           = "stock.low"
	EventSyncRequired       = "sync.required"
)

// Event is the standard message envelope sent over WebSocket.
type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
	// BranchID can be zero to broadcast to all branches.
	BranchID int64 `json:"-"`
}

// Client represents a single WebSocket connection.
type Client struct {
	Hub      *Hub
	Send     chan []byte
	BranchID int64 // 0 means all branches
	UserID   int64
	Done     chan struct{}
}

// Hub manages all connected WebSocket clients and routes events by branch.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]bool
	// rooms maps branch_id -> set of clients in that branch.
	rooms map[int64]map[*Client]bool
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]bool),
		rooms:   make(map[int64]map[*Client]bool),
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[c] = true

	// Add to branch room
	branchID := c.BranchID
	if _, ok := h.rooms[branchID]; !ok {
		h.rooms[branchID] = make(map[*Client]bool)
	}
	h.rooms[branchID][c] = true

	log.Printf("[ws] client registered (branch_id=%d, total=%d)", branchID, len(h.clients))
}

// Unregister removes a client from the hub and closes its send channel.
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		// Remove from branch room
		branchID := c.BranchID
		if room, ok := h.rooms[branchID]; ok {
			delete(room, c)
			if len(room) == 0 {
				delete(h.rooms, branchID)
			}
		}
		close(c.Send)
		log.Printf("[ws] client unregistered (branch_id=%d, total=%d)", branchID, len(h.clients))
	}
}

// BroadcastEvent sends an event to all connected clients.
func (h *Hub) BroadcastEvent(evt Event) {
	data, err := json.Marshal(evt)
	if err != nil {
		log.Printf("[ws] marshal event error: %v", err)
		return
	}
	h.broadcastBytes(data, evt.BranchID)
}

// BroadcastEventToBranch sends an event only to clients in a specific branch.
// Pass branchID = 0 to broadcast to all branches.
func (h *Hub) BroadcastEventToBranch(branchID int64, evt Event) {
	evt.BranchID = branchID
	h.BroadcastEvent(evt)
}

// broadcastBytes sends raw JSON bytes to matching clients.
func (h *Hub) broadcastBytes(data []byte, branchID int64) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if branchID > 0 {
		// Send to specific branch room only
		if room, ok := h.rooms[branchID]; ok {
			for c := range room {
				select {
				case c.Send <- data:
				default:
					// Client send buffer full; skip (will be cleaned up on next read error)
					log.Printf("[ws] dropping message to slow client (branch_id=%d)", branchID)
				}
			}
		}
	} else {
		// Broadcast to ALL clients
		for c := range h.clients {
			select {
			case c.Send <- data:
			default:
				log.Printf("[ws] dropping message to slow client")
			}
		}
	}
}

// ClientCount returns the total number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// RoomClientCount returns the number of clients in a specific branch room.
func (h *Hub) RoomClientCount(branchID int64) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if room, ok := h.rooms[branchID]; ok {
		return len(room)
	}
	return 0
}

// ServeHTTP makes Hub implement http.Handler by upgrading to WebSocket.
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.ServeWS(w, r)
}
