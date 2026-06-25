package sync

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Sync status constants.
const (
	StatusPending  = "pending"
	StatusSynced   = "synced"
	StatusConflict = "conflict"
)

<<<<<<< HEAD
=======
// validSyncTables is a whitelist of table names allowed in dynamic SQL queries.
// Never add user-controlled values — only hardcoded table names that exist in the schema.
var validSyncTables = map[string]bool{
	"transactions":      true,
	"transaction_items": true,
	"products":          true,
	"categories":        true,
	"branches":          true,
	"branch_products":   true,
	"sync_queue":        true,
	"users":             true,
}

// safeTableName validates a table name against the whitelist.
// Returns the table name if valid, or empty string if not.
func safeTableName(name string) string {
	if validSyncTables[name] {
		return name
	}
	return ""
}

>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
// SyncRecord represents a queued sync operation.
type SyncRecord struct {
	ID        int64  `json:"id"`
	TableName string `json:"table_name"`
	RecordID  int64  `json:"record_id"`
	Action    string `json:"action"`
	Payload   string `json:"payload,omitempty"`
	CreatedAt string `json:"created_at"`
	Status    string `json:"status"`
}

// SyncConfig holds configuration for the sync engine.
type SyncConfig struct {
	APIBaseURL  string
	APIKey      string
	BranchID    int64
	SyncInterval time.Duration
	BatchSize   int
}

// Engine is the main sync engine coordinating push and pull.
type Engine struct {
	config   SyncConfig
	sqliteDB *sql.DB
	client   *http.Client
	mu       sync.Mutex
	stopCh   chan struct{}
}

// NewEngine creates a new sync engine.
func NewEngine(db *sql.DB, cfg SyncConfig) *Engine {
	return &Engine{
		config:   cfg,
		sqliteDB: db,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		stopCh: make(chan struct{}),
	}
}

// Start begins the periodic sync loop.
func (e *Engine) Start(ctx context.Context) {
	log.Printf("[sync] engine started — push/pull every %v", e.config.SyncInterval)
	ticker := time.NewTicker(e.config.SyncInterval)
	defer ticker.Stop()

	// Run once immediately
	e.syncOnce(ctx)

	for {
		select {
		case <-ticker.C:
			e.syncOnce(ctx)
		case <-e.stopCh:
			log.Printf("[sync] engine stopped")
			return
		case <-ctx.Done():
			return
		}
	}
}

// Stop signals the engine to stop.
func (e *Engine) Stop() {
	close(e.stopCh)
}

// syncOnce runs a single push + pull cycle.
func (e *Engine) syncOnce(ctx context.Context) {
	e.mu.Lock()
	defer e.mu.Unlock()

	log.Printf("[sync] cycle starting")

	// 1. Push pending local data upstream
	pushErr := e.PushEngine(ctx)
	if pushErr != nil {
		log.Printf("[sync] push error: %v", pushErr)
	}

	// 2. Pull master data from upstream
	pullErr := e.PullEngine(ctx)
	if pullErr != nil {
		log.Printf("[sync] pull error: %v", pullErr)
	}

	log.Printf("[sync] cycle finished (push err: %v, pull err: %v)", pushErr, pullErr)
}

// ============================================================================
// PushEngine — kirim pending data (transaksi, stock mutations) ke API
// ============================================================================

// PushEngine sends all pending local records to the central API.
func (e *Engine) PushEngine(ctx context.Context) error {
	// 1. Collect pending transactions
	pendingTx, err := e.getPendingTransactions(ctx)
	if err != nil {
		return fmt.Errorf("get pending transactions: %w", err)
	}

	for _, tx := range pendingTx {
		if err := e.pushTransaction(ctx, tx); err != nil {
			log.Printf("[sync] push tx %d failed: %v — marking conflict", tx.ID, err)
			e.markConflict("transactions", tx.ID)
			continue
		}
		e.markSynced("transactions", tx.ID)
		e.markSyncedItemsByTransaction(tx.ID)
	}

	// 2. Collect stock mutation records from sync_queue
	queueItems, err := e.getQueueItems(ctx, "stock_mutation")
	if err != nil {
		return fmt.Errorf("get stock mutation queue: %w", err)
	}

	for _, item := range queueItems {
		if err := e.pushQueueItem(ctx, item); err != nil {
			log.Printf("[sync] push queue item %d failed: %v", item.ID, err)
			e.markQueueFailed(item.ID)
			continue
		}
		e.markQueueDone(item.ID)
	}

	return nil
}

type pendingTransactionRow struct {
	ID            int64
	BranchID      int64
	TotalAmount   float64
	Discount      float64
	PaymentMethod string
	PaymentAmount float64
	ChangeAmount  float64
	Status        string
	CreatedBy     sql.NullInt64
	CreatedAt     string
	UpdatedAt     string
}

func (e *Engine) getPendingTransactions(ctx context.Context) ([]pendingTransactionRow, error) {
	rows, err := e.sqliteDB.QueryContext(ctx, `
		SELECT id, branch_id, total_amount, discount, payment_method,
		       payment_amount, change_amount, status, created_by,
		       created_at, updated_at
		FROM transactions
		WHERE sync_status = 'pending'
		ORDER BY id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []pendingTransactionRow
	for rows.Next() {
		var r pendingTransactionRow
		if err := rows.Scan(&r.ID, &r.BranchID, &r.TotalAmount, &r.Discount,
			&r.PaymentMethod, &r.PaymentAmount, &r.ChangeAmount,
			&r.Status, &r.CreatedBy, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}

func (e *Engine) pushTransaction(ctx context.Context, tx pendingTransactionRow) error {
	payload := map[string]interface{}{
		"id":             tx.ID,
		"branch_id":      tx.BranchID,
		"total_amount":   tx.TotalAmount,
		"discount":       tx.Discount,
		"payment_method": tx.PaymentMethod,
		"payment_amount": tx.PaymentAmount,
		"change_amount":  tx.ChangeAmount,
		"status":         tx.Status,
		"created_by":     tx.CreatedBy.Int64,
		"created_at":     tx.CreatedAt,
		"updated_at":     tx.UpdatedAt,
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/api/v1/sync/push", e.config.APIBaseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", e.config.APIKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("upstream returned %d", resp.StatusCode)
	}
	return nil
}

func (e *Engine) markSynced(table string, recordID int64) {
<<<<<<< HEAD
	now := time.Now().UTC().Format(time.RFC3339)
	_, _ = e.sqliteDB.Exec(
		fmt.Sprintf(`UPDATE %s SET pending_sync=0, synced_at=?, sync_status='synced' WHERE id=?`, table),
=======
	t := safeTableName(table)
	if t == "" {
		log.Printf("[sync] markSynced: invalid table name %q", table)
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, _ = e.sqliteDB.Exec(
		fmt.Sprintf(`UPDATE %s SET pending_sync=0, synced_at=?, sync_status='synced' WHERE id=?`, t),
>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
		now, recordID,
	)
}

func (e *Engine) markConflict(table string, recordID int64) {
<<<<<<< HEAD
	_, _ = e.sqliteDB.Exec(
		fmt.Sprintf(`UPDATE %s SET sync_status='conflict' WHERE id=?`, table),
=======
	t := safeTableName(table)
	if t == "" {
		log.Printf("[sync] markConflict: invalid table name %q", table)
		return
	}
	_, _ = e.sqliteDB.Exec(
		fmt.Sprintf(`UPDATE %s SET sync_status='conflict' WHERE id=?`, t),
>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
		recordID,
	)
}

func (e *Engine) markSyncedItemsByTransaction(txID int64) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, _ = e.sqliteDB.Exec(
		`UPDATE transaction_items SET pending_sync=0, synced_at=?, sync_status='synced' WHERE transaction_id=?`,
		now, txID,
	)
}

func (e *Engine) getQueueItems(ctx context.Context, kind string) ([]SyncRecord, error) {
	rows, err := e.sqliteDB.QueryContext(ctx, `
		SELECT id, table_name, record_id, action, payload, created_at, status
		FROM sync_queue
		WHERE status = 'pending' AND ($1 = '' OR table_name = $1)
		ORDER BY id ASC
	`, kind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []SyncRecord
	for rows.Next() {
		var item SyncRecord
		if err := rows.Scan(&item.ID, &item.TableName, &item.RecordID, &item.Action,
			&item.Payload, &item.CreatedAt, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (e *Engine) pushQueueItem(ctx context.Context, item SyncRecord) error {
	url := fmt.Sprintf("%s/api/v1/sync/push", e.config.APIBaseURL)
	body := bytes.NewReader([]byte(item.Payload))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", e.config.APIKey)
	req.Header.Set("X-Sync-Table", item.TableName)
	req.Header.Set("X-Sync-Action", item.Action)

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("upstream returned %d", resp.StatusCode)
	}
	return nil
}

func (e *Engine) markQueueDone(id int64) {
	_, _ = e.sqliteDB.Exec(`UPDATE sync_queue SET status='done' WHERE id=?`, id)
}

func (e *Engine) markQueueFailed(id int64) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, _ = e.sqliteDB.Exec(`UPDATE sync_queue SET status='failed', payload=json_set(COALESCE(payload,'{}'),'$.retry_at',?) WHERE id=?`, now, id)
}

// ============================================================================
// PullEngine — tarik master data (products, categories) dari API
// ============================================================================

// PullEngine fetches updated master data from the central API and upserts into local SQLite.
func (e *Engine) PullEngine(ctx context.Context) error {
	since := e.getLastPullTimestamp()

	// Pull categories
	if err := e.pullCategories(ctx, since); err != nil {
		log.Printf("[sync] pull categories error: %v", err)
	}

	// Pull products
	if err := e.pullProducts(ctx, since); err != nil {
		log.Printf("[sync] pull products error: %v", err)
	}

	// Pull branches
	if err := e.pullBranches(ctx, since); err != nil {
		log.Printf("[sync] pull branches error: %v", err)
	}

	e.setLastPullTimestamp(time.Now().UTC().Format(time.RFC3339))
	return nil
}

func (e *Engine) getLastPullTimestamp() string {
	var ts string
	err := e.sqliteDB.QueryRow(`SELECT COALESCE(MAX(created_at),'1970-01-01T00:00:00Z') FROM sync_queue WHERE status='done' AND table_name='_meta_pull'`).Scan(&ts)
	if err != nil {
		return "1970-01-01T00:00:00Z"
	}
	return ts
}

func (e *Engine) setLastPullTimestamp(ts string) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, _ = e.sqliteDB.Exec(
		`INSERT INTO sync_queue (table_name, record_id, action, payload, created_at, status) VALUES ('_meta_pull', 0, 'pull', ?, ?, 'done')`,
		ts, now,
	)
}

func (e *Engine) pullCategories(ctx context.Context, since string) error {
	url := fmt.Sprintf("%s/api/v1/sync/pull?table=categories&since=%s", e.config.APIBaseURL, since)
	return e.pullAndUpsert(ctx, url, "categories", func(item map[string]interface{}) error {
		_, err := e.sqliteDB.Exec(`
			INSERT INTO categories (id, name, created_at, updated_at, pending_sync, sync_status)
			VALUES (?, ?, ?, ?, 0, 'synced')
			ON CONFLICT(id) DO UPDATE SET
				name=excluded.name,
				updated_at=excluded.updated_at,
				sync_status='synced'
		`, item["id"], item["name"], item["created_at"], item["updated_at"])
		return err
	})
}

func (e *Engine) pullProducts(ctx context.Context, since string) error {
	url := fmt.Sprintf("%s/api/v1/sync/pull?table=products&since=%s", e.config.APIBaseURL, since)
	return e.pullAndUpsert(ctx, url, "products", func(item map[string]interface{}) error {
		_, err := e.sqliteDB.Exec(`
			INSERT INTO products (id, category_id, name, sku, description, price, cost, unit, is_active, created_at, updated_at, deleted_at, pending_sync, sync_status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 'synced')
			ON CONFLICT(id) DO UPDATE SET
				category_id=excluded.category_id,
				name=excluded.name,
				sku=excluded.sku,
				description=excluded.description,
				price=excluded.price,
				cost=excluded.cost,
				unit=excluded.unit,
				is_active=excluded.is_active,
				updated_at=excluded.updated_at,
				deleted_at=excluded.deleted_at,
				sync_status='synced'
		`,
			item["id"], item["category_id"], item["name"], item["sku"],
			item["description"], item["price"], item["cost"],
			item["unit"], item["is_active"],
			item["created_at"], item["updated_at"], item["deleted_at"],
		)
		return err
	})
}

func (e *Engine) pullBranches(ctx context.Context, since string) error {
	url := fmt.Sprintf("%s/api/v1/sync/pull?table=branches&since=%s", e.config.APIBaseURL, since)
	return e.pullAndUpsert(ctx, url, "branches", func(item map[string]interface{}) error {
		_, err := e.sqliteDB.Exec(`
			INSERT INTO branches (id, name, code, address, city, phone, is_active, created_at, updated_at, deleted_at, pending_sync, sync_status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 'synced')
			ON CONFLICT(id) DO UPDATE SET
				name=excluded.name,
				code=excluded.code,
				address=excluded.address,
				city=excluded.city,
				phone=excluded.phone,
				is_active=excluded.is_active,
				updated_at=excluded.updated_at,
				deleted_at=excluded.deleted_at,
				sync_status='synced'
		`,
			item["id"], item["name"], item["code"],
			item["address"], item["city"], item["phone"],
			item["is_active"],
			item["created_at"], item["updated_at"], item["deleted_at"],
		)
		return err
	})
}

func (e *Engine) pullAndUpsert(ctx context.Context, url string, table string, upsertFn func(map[string]interface{}) error) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", e.config.APIKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pull %s returned %d", table, resp.StatusCode)
	}

	var result struct {
		Success bool                      `json:"success"`
		Data    []map[string]interface{}  `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode %s response: %w", table, err)
	}

	if !result.Success {
		return fmt.Errorf("pull %s: upstream returned failure", table)
	}

	for _, item := range result.Data {
		if err := upsertFn(item); err != nil {
			return fmt.Errorf("upsert %s item: %w", table, err)
		}
	}

	log.Printf("[sync] pulled %d %s records", len(result.Data), table)
	return nil
}

// ============================================================================
// SyncQueue — antrian pending data
// ============================================================================

// Enqueue adds a record to the sync queue for later processing.
func (e *Engine) Enqueue(tableName string, recordID int64, action string, payload interface{}) error {
	var payloadStr string
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal payload: %w", err)
		}
		payloadStr = string(b)
	}

	_, err := e.sqliteDB.Exec(
		`INSERT INTO sync_queue (table_name, record_id, action, payload, status) VALUES (?, ?, ?, ?, 'pending')`,
		tableName, recordID, action, payloadStr,
	)
	if err != nil {
		return fmt.Errorf("enqueue: %w", err)
	}
	return nil
}

// GetPendingCount returns the number of pending items in the queue.
func (e *Engine) GetPendingCount(ctx context.Context) (int, error) {
	var count int
	err := e.sqliteDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM sync_queue WHERE status = 'pending'`).Scan(&count)
	return count, err
}

// GetPendingItems returns all pending sync queue items.
func (e *Engine) GetPendingItems(ctx context.Context) ([]SyncRecord, error) {
	return e.getQueueItems(ctx, "")
}

// ResolveConflict marks a conflicted record and enqueues a resolution.
func (e *Engine) ResolveConflict(tableName string, recordID int64, resolution map[string]interface{}) error {
<<<<<<< HEAD
=======
	t := safeTableName(tableName)
	if t == "" {
		return fmt.Errorf("resolve conflict: invalid table name %q", tableName)
	}

>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
	// Mark the local record as synced with the resolved data
	now := time.Now().UTC().Format(time.RFC3339)

	// Update sync_status
	_, err := e.sqliteDB.Exec(
<<<<<<< HEAD
		fmt.Sprintf(`UPDATE %s SET sync_status='synced', pending_sync=0, synced_at=? WHERE id=?`, tableName),
=======
		fmt.Sprintf(`UPDATE %s SET sync_status='synced', pending_sync=0, synced_at=? WHERE id=?`, t),
>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
		now, recordID,
	)
	if err != nil {
		return fmt.Errorf("resolve conflict update: %w", err)
	}

	// Enqueue the resolution payload
<<<<<<< HEAD
	return e.Enqueue(tableName, recordID, "resolve", resolution)
=======
	return e.Enqueue(t, recordID, "resolve", resolution)
>>>>>>> 90c46f770f2582ca6c2d103b433a1a70dc1620f9
}
