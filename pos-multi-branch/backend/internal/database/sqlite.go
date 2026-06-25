package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDB wraps the local SQLite database used for offline sync mirror.
type SQLiteDB struct {
	DB *sql.DB
}

// NewSQLiteDB opens or creates the local SQLite sync mirror.
func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create sqlite dir: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	s := &SQLiteDB{DB: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("sqlite migrate: %w", err)
	}

	log.Printf("[sqlite] sync mirror ready at %s", dbPath)
	return s, nil
}

// migrate creates the mirror schema mirroring Supabase tables plus sync fields.
func (s *SQLiteDB) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		email TEXT NOT NULL,
		name TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'cashier',
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		pending_sync INTEGER NOT NULL DEFAULT 0,
		synced_at TEXT,
		sync_status TEXT NOT NULL DEFAULT 'synced'
	);

	CREATE TABLE IF NOT EXISTS branches (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		code TEXT NOT NULL,
		address TEXT NOT NULL DEFAULT '',
		city TEXT NOT NULL DEFAULT '',
		phone TEXT NOT NULL DEFAULT '',
		is_active INTEGER NOT NULL DEFAULT 1,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		deleted_at TEXT,
		pending_sync INTEGER NOT NULL DEFAULT 0,
		synced_at TEXT,
		sync_status TEXT NOT NULL DEFAULT 'synced'
	);

	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		pending_sync INTEGER NOT NULL DEFAULT 0,
		synced_at TEXT,
		sync_status TEXT NOT NULL DEFAULT 'synced'
	);

	CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY,
		category_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		sku TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		price REAL NOT NULL DEFAULT 0,
		cost REAL NOT NULL DEFAULT 0,
		unit TEXT NOT NULL DEFAULT 'pcs',
		is_active INTEGER NOT NULL DEFAULT 1,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		deleted_at TEXT,
		pending_sync INTEGER NOT NULL DEFAULT 0,
		synced_at TEXT,
		sync_status TEXT NOT NULL DEFAULT 'synced'
	);

	CREATE TABLE IF NOT EXISTS branch_products (
		id INTEGER PRIMARY KEY,
		branch_id INTEGER NOT NULL,
		product_id INTEGER NOT NULL,
		stock INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		pending_sync INTEGER NOT NULL DEFAULT 0,
		synced_at TEXT,
		sync_status TEXT NOT NULL DEFAULT 'synced'
	);

	CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY,
		branch_id INTEGER NOT NULL,
		total_amount REAL NOT NULL DEFAULT 0,
		discount REAL NOT NULL DEFAULT 0,
		payment_method TEXT NOT NULL DEFAULT 'cash',
		payment_amount REAL NOT NULL DEFAULT 0,
		change_amount REAL NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'completed',
		created_by INTEGER,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		pending_sync INTEGER NOT NULL DEFAULT 1,
		synced_at TEXT,
		sync_status TEXT NOT NULL DEFAULT 'pending'
	);

	CREATE TABLE IF NOT EXISTS transaction_items (
		id INTEGER PRIMARY KEY,
		transaction_id INTEGER NOT NULL,
		product_id INTEGER NOT NULL,
		product_name TEXT NOT NULL DEFAULT '',
		qty INTEGER NOT NULL DEFAULT 0,
		price REAL NOT NULL DEFAULT 0,
		subtotal REAL NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL,
		pending_sync INTEGER NOT NULL DEFAULT 1,
		synced_at TEXT,
		sync_status TEXT NOT NULL DEFAULT 'pending'
	);

	CREATE TABLE IF NOT EXISTS sync_queue (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		table_name TEXT NOT NULL,
		record_id INTEGER NOT NULL,
		action TEXT NOT NULL,
		payload TEXT,
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		status TEXT NOT NULL DEFAULT 'pending'
	);
	`

	_, err := s.DB.Exec(schema)
	return err
}

// Close closes the SQLite database.
func (s *SQLiteDB) Close() error {
	return s.DB.Close()
}
