package ws

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 4096

	// Send channel buffer size.
	sendBufferSize = 64
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins; tighten in production.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// connRegistry maps clients to their raw WebSocket connections.
var (
	connMu sync.Mutex
	conns  = make(map[*Client]*websocket.Conn)
)

// ServeWS handles the WebSocket upgrade at WS /api/v1/ws.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[ws] upgrade error: %v", err)
		return
	}

	// Parse optional branch_id from query param.
	branchIDStr := r.URL.Query().Get("branch_id")
	var branchID int64
	if branchIDStr != "" {
		if bid, err := strconv.ParseInt(branchIDStr, 10, 64); err == nil {
			branchID = bid
		}
	}

	// Parse optional user_id from query param.
	userIDStr := r.URL.Query().Get("user_id")
	var userID int64
	if userIDStr != "" {
		if uid, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			userID = uid
		}
	}

	client := &Client{
		Hub:      h,
		Send:     make(chan []byte, sendBufferSize),
		BranchID: branchID,
		UserID:   userID,
		Done:     make(chan struct{}),
	}

	h.Register(client)

	// Store the connection reference for the pumps.
	connMu.Lock()
	conns[client] = conn
	connMu.Unlock()

	// Start read and write pumps.
	go client.writePump()
	go client.readPump()
}

// readPump reads messages from the WebSocket connection.
// The server does not process incoming messages beyond handling
// control frames (close, pong), but reading is required to detect
// disconnection.
func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister(c)
		close(c.Done)
		// Clean up conn reference.
		connMu.Lock()
		conn := conns[c]
		delete(conns, c)
		connMu.Unlock()
		if conn != nil {
			conn.Close()
		}
	}()

	conn := c.getConn()
	if conn == nil {
		return
	}

	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("[ws] read error: %v", err)
			}
			break
		}
	}
}

// writePump pumps messages from the hub's send channel to the WebSocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn := c.getConn()
		if conn != nil {
			conn.Close()
		}
		// Clean up conn reference.
		connMu.Lock()
		delete(conns, c)
		connMu.Unlock()
	}()

	conn := c.getConn()
	if conn == nil {
		return
	}

	for {
		select {
		case message, ok := <-c.Send:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel.
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Drain any additional queued messages into the same frame.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.Done:
			return
		}
	}
}

// getConn safely retrieves the WebSocket connection for this client.
func (c *Client) getConn() *websocket.Conn {
	connMu.Lock()
	defer connMu.Unlock()
	return conns[c]
}

// ensure fmt is used (referenced for potential future use).
var _ = fmt.Sprintf
