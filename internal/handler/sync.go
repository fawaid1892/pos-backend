package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"pos-multi-branch/backend/internal/model"
<<<<<<< HEAD
	"pos-multi-branch/backend/internal/sync"

	"github.com/labstack/echo/v4"
)

type SyncHandler struct {
	syncEngine *sync.Engine
	sqliteDB   *sql.DB
}

func NewSyncHandler(engine *sync.Engine, sqliteDB *sql.DB) *SyncHandler {
	return &SyncHandler{
		syncEngine: engine,
		sqliteDB:   sqliteDB,
=======
)

// SyncHandler handles sync push/pull/resolve endpoints.
type SyncHandler struct {
	sqliteDB *sql.DB
}

// validPullTables is the whitelist of table names allowed for pull queries.
var validPullTables = map[string]bool{
	"categories": true,
	"products":   true,
	"branches":   true,
}

// validResolveTables is the whitelist of table names allowed for conflict resolution.
var validResolveTables = map[string]bool{
	"transactions":      true,
	"transaction_items": true,
	"products":          true,
	"categories":        true,
	"branches":          true,
	"branch_products":   true,
}

// validResolveColumns is the whitelist of column names allowed in resolution field updates.
var validResolveColumns = map[string]bool{
	"status":          true,
	"total_amount":    true,
	"discount":        true,
	"payment_method":  true,
	"name":            true,
	"price":           true,
	"cost":            true,
	"stock":           true,
	"is_active":       true,
}

// NewSyncHandler creates a new SyncHandler.
func NewSyncHandler(sqliteDB *sql.DB) *SyncHandler {
	return &SyncHandler{
		sqliteDB: sqliteDB,
>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
	}
}

// SyncPushRequest represents an incoming transaction payload from a branch.
type SyncPushRequest struct {
<<<<<<< HEAD
	ID            int64   `json:"id"`
	BranchID      int64   `json:"branch_id"`
	TotalAmount   float64 `json:"total_amount"`
	Discount      float64 `json:"discount"`
	PaymentMethod string  `json:"payment_method"`
	PaymentAmount float64 `json:"payment_amount"`
	ChangeAmount  float64 `json:"change_amount"`
	Status        string  `json:"status"`
	CreatedBy     int64   `json:"created_by"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
	Items         []SyncPushItem `json:"items,omitempty"`
}

=======
	ID            int64          `json:"id"`
	BranchID      int64          `json:"branch_id"`
	TotalAmount   float64        `json:"total_amount"`
	Discount      float64        `json:"discount"`
	PaymentMethod string         `json:"payment_method"`
	PaymentAmount float64        `json:"payment_amount"`
	ChangeAmount  float64        `json:"change_amount"`
	Status        string         `json:"status"`
	CreatedBy     int64          `json:"created_by"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
	Items         []SyncPushItem `json:"items,omitempty"`
}

// SyncPushItem represents a single transaction item in a sync push.
>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
type SyncPushItem struct {
	ProductID   int64   `json:"product_id"`
	ProductName string  `json:"product_name"`
	Qty         int     `json:"qty"`
	Price       float64 `json:"price"`
	Subtotal    float64 `json:"subtotal"`
}

// Push handles POST /api/v1/sync/push — receive pending transactions from branches.
<<<<<<< HEAD
// The central server stores them and acknowledges.
func (h *SyncHandler) Push(c echo.Context) error {
	var req SyncPushRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
	}

	// Validate required fields
	if req.BranchID == 0 {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "branch_id is required",
		})
	}

	// Detect sync table from header (for stock mutations etc)
	tableName := c.Request().Header.Get("X-Sync-Table")
	action := c.Request().Header.Get("X-Sync-Action")

	if tableName != "" && action != "" {
		// Handle generic queue item
		payload, _ := json.Marshal(req)
		if err := h.syncEngine.Enqueue(tableName, req.ID, action, string(payload)); err != nil {
			log.Printf("[sync-handler] enqueue error: %v", err)
			return c.JSON(http.StatusInternalServerError, model.APIResponse{
				Success: false,
				Error:   "failed to enqueue",
			})
		}

		return c.JSON(http.StatusOK, model.APIResponse{
			Success: true,
			Message: "sync item received",
		})
=======
func (h *SyncHandler) Push(w http.ResponseWriter, r *http.Request) {
	var req SyncPushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	if req.BranchID == 0 {
		writeJSON(w, http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "branch_id is required",
		})
		return
	}

	// Detect sync table from header (for stock mutations etc)
	tableName := r.Header.Get("X-Sync-Table")
	action := r.Header.Get("X-Sync-Action")

	if tableName != "" && action != "" {
		// Enqueue generic sync item directly into sync_queue
		payload, _ := json.Marshal(req)
		now := time.Now().UTC().Format(time.RFC3339)
		_, err := h.sqliteDB.Exec(
			`INSERT INTO sync_queue (table_name, record_id, action, payload, created_at, status) VALUES (?, ?, ?, ?, ?, 'pending')`,
			tableName, req.ID, action, string(payload), now,
		)
		if err != nil {
			log.Printf("[sync] enqueue error: %v", err)
			writeJSON(w, http.StatusInternalServerError, model.APIResponse{
				Success: false,
				Error:   "failed to enqueue",
			})
			return
		}

		writeJSON(w, http.StatusOK, model.APIResponse{
			Success: true,
			Message: "sync item received",
		})
		return
>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
	}

	// Default: store as a received transaction in the central DB
	now := time.Now().UTC().Format(time.RFC3339)

<<<<<<< HEAD
	// Upsert into the central server's transactions table (via sqlite mirror)
=======
>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
	_, err := h.sqliteDB.Exec(`
		INSERT INTO transactions (id, branch_id, total_amount, discount, payment_method,
			payment_amount, change_amount, status, created_by, created_at, updated_at,
			pending_sync, sync_status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 'synced')
		ON CONFLICT(id) DO UPDATE SET
			status=excluded.status,
			updated_at=excluded.updated_at,
			sync_status='synced'
	`, req.ID, req.BranchID, req.TotalAmount, req.Discount, req.PaymentMethod,
		req.PaymentAmount, req.ChangeAmount, req.Status, req.CreatedBy,
		req.CreatedAt, req.UpdatedAt,
	)
	if err != nil {
<<<<<<< HEAD
		log.Printf("[sync-handler] upsert error: %v", err)
		return c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "failed to store",
		})
=======
		log.Printf("[sync] upsert transaction error: %v", err)
		writeJSON(w, http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "failed to store",
		})
		return
>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
	}

	// Store items if present
	for _, item := range req.Items {
		_, _ = h.sqliteDB.Exec(`
			INSERT INTO transaction_items (transaction_id, product_id, product_name, qty, price, subtotal, created_at, pending_sync, sync_status)
			VALUES (?, ?, ?, ?, ?, ?, ?, 0, 'synced')
			ON CONFLICT DO NOTHING
		`, req.ID, item.ProductID, item.ProductName, item.Qty, item.Price, item.Subtotal, now)
	}

<<<<<<< HEAD
	return c.JSON(http.StatusOK, model.APIResponse{
=======
	writeJSON(w, http.StatusOK, model.APIResponse{
>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
		Success: true,
		Message: "transaction received",
	})
}

// Pull handles GET /api/v1/sync/pull?table=&since= — send master data updates to branches.
<<<<<<< HEAD
func (h *SyncHandler) Pull(c echo.Context) error {
	tableName := c.QueryParam("table")
	since := c.QueryParam("since")

	if tableName == "" {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "table query parameter is required",
		})
=======
func (h *SyncHandler) Pull(w http.ResponseWriter, r *http.Request) {
	tableName := r.URL.Query().Get("table")
	since := r.URL.Query().Get("since")

	if tableName == "" {
		writeJSON(w, http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "table query parameter is required",
		})
		return
>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
	}

	if since == "" {
		since = "1970-01-01T00:00:00Z"
	}

	// Parse the since timestamp
<<<<<<< HEAD
	sinceTime, err := time.Parse(time.RFC3339, since)
	if err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid since format, use RFC3339 (e.g. 2024-01-01T00:00:00Z)",
		})
	}

	var rowsData []map[string]interface{}
	query := fmt.Sprintf(`SELECT * FROM %s WHERE updated_at >= ? ORDER BY updated_at ASC`, tableName)

	switch tableName {
	case "categories":
		rowsData, err = h.queryRowsMap(query, sinceTime.Format(time.RFC3339))
	case "products":
		rowsData, err = h.queryRowsMap(query, sinceTime.Format(time.RFC3339))
	case "branches":
		rowsData, err = h.queryRowsMap(query, sinceTime.Format(time.RFC3339))
	default:
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "unsupported table: " + tableName,
		})
	}
	if err != nil {
		log.Printf("[sync-handler] query error: %v", err)
		return c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "query failed",
		})
	}

	return c.JSON(http.StatusOK, model.APIResponse{
=======
	_, err := time.Parse(time.RFC3339, since)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid since format, use RFC3339 (e.g. 2024-01-01T00:00:00Z)",
		})
		return
	}

	// Validate table name against whitelist
	if !validPullTables[tableName] {
		writeJSON(w, http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "unsupported table: " + tableName,
		})
		return
	}

	query := fmt.Sprintf(`SELECT * FROM %s WHERE updated_at >= ? ORDER BY updated_at ASC`, tableName)
	rowsData, err := h.queryRowsMap(query, since)
	if err != nil {
		log.Printf("[sync] pull query error: %v", err)
		writeJSON(w, http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "query failed",
		})
		return
	}

	writeJSON(w, http.StatusOK, model.APIResponse{
>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
		Success: true,
		Data:    rowsData,
	})
}

// Resolve handles POST /api/v1/sync/resolve — conflict resolution.
<<<<<<< HEAD
func (h *SyncHandler) Resolve(c echo.Context) error {
=======
func (h *SyncHandler) Resolve(w http.ResponseWriter, r *http.Request) {
>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
	var req struct {
		TableName  string                 `json:"table_name"`
		RecordID   int64                  `json:"record_id"`
		Resolution map[string]interface{} `json:"resolution"`
	}

<<<<<<< HEAD
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
	}

	if req.TableName == "" || req.RecordID == 0 {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "table_name and record_id are required",
		})
	}

	// Validate table name
	validTables := map[string]bool{
		"transactions": true, "transaction_items": true,
		"products": true, "categories": true,
		"branches": true, "branch_products": true,
	}
	if !validTables[req.TableName] {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid table_name: " + req.TableName,
		})
=======
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	if req.TableName == "" || req.RecordID == 0 {
		writeJSON(w, http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "table_name and record_id are required",
		})
		return
	}

	// Validate table name against whitelist
	if !validResolveTables[req.TableName] {
		writeJSON(w, http.StatusBadRequest, model.APIResponse{
			Success: false,
			Error:   "invalid table_name: " + req.TableName,
		})
		return
>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
	}

	// Apply resolution on the central side
	now := time.Now().UTC().Format(time.RFC3339)

	tx, err := h.sqliteDB.Begin()
	if err != nil {
<<<<<<< HEAD
		return c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "failed to begin transaction",
		})
	}
	defer tx.Rollback()

	// Mark the record as resolved
	_, err = tx.Exec(
		`UPDATE `+req.TableName+` SET sync_status='synced', pending_sync=0, synced_at=? WHERE id=? AND sync_status='conflict'`,
		now, req.RecordID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "failed to resolve conflict",
		})
	}

	// Apply resolution fields if provided
	if len(req.Resolution) > 0 {
		for key, val := range req.Resolution {
			col := key
			// Simple column update — only for known safe columns
			switch col {
			case "status", "total_amount", "discount", "payment_method",
				"name", "price", "cost", "stock", "is_active":
				_, err = tx.Exec(
					`UPDATE `+req.TableName+` SET `+col+`=? WHERE id=?`,
					val, req.RecordID,
				)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, model.APIResponse{
						Success: false,
						Error:   "failed to apply resolution field: " + key,
					})
				}
=======
		writeJSON(w, http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "failed to begin transaction",
		})
		return
	}
	defer tx.Rollback()

	// Mark the record as resolved (table name already whitelisted above)
	_, err = tx.Exec(
		fmt.Sprintf(`UPDATE %s SET sync_status='synced', pending_sync=0, synced_at=? WHERE id=? AND sync_status='conflict'`, req.TableName),
		now, req.RecordID,
	)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "failed to resolve conflict",
		})
		return
	}

	// Apply resolution fields if provided (column names validated against whitelist)
	if len(req.Resolution) > 0 {
		for key, val := range req.Resolution {
			col := key
			if !validResolveColumns[col] {
				writeJSON(w, http.StatusBadRequest, model.APIResponse{
					Success: false,
					Error:   "invalid resolution field: " + key,
				})
				return
			}
			_, err = tx.Exec(
				fmt.Sprintf(`UPDATE %s SET %s=? WHERE id=?`, req.TableName, col),
				val, req.RecordID,
			)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, model.APIResponse{
					Success: false,
					Error:   "failed to apply resolution field: " + key,
				})
				return
>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
			}
		}
	}

	if err := tx.Commit(); err != nil {
<<<<<<< HEAD
		return c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "failed to commit resolution",
		})
	}

	log.Printf("[sync-handler] resolved conflict in %s id=%d", req.TableName, req.RecordID)

	return c.JSON(http.StatusOK, model.APIResponse{
=======
		writeJSON(w, http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Error:   "failed to commit resolution",
		})
		return
	}

	log.Printf("[sync] resolved conflict in %s id=%d", req.TableName, req.RecordID)

	writeJSON(w, http.StatusOK, model.APIResponse{
>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
		Success: true,
		Message: "conflict resolved",
	})
}

// queryRowsMap runs a query and returns all rows as a slice of maps.
func (h *SyncHandler) queryRowsMap(query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := h.sqliteDB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			switch v := val.(type) {
			case []byte:
				row[col] = string(v)
			default:
				row[col] = v
			}
		}
		result = append(result, row)
	}

	if result == nil {
		result = []map[string]interface{}{}
	}
	return result, nil
}
