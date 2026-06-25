-- 003_sync_fields.sql
-- POS Multi Branch — Add sync fields to existing tables

-- Users sync fields
ALTER TABLE users ADD COLUMN IF NOT EXISTS pending_sync BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS synced_at TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN IF NOT EXISTS sync_status TEXT NOT NULL DEFAULT 'synced';

-- Branches sync fields
ALTER TABLE branches ADD COLUMN IF NOT EXISTS pending_sync BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE branches ADD COLUMN IF NOT EXISTS synced_at TIMESTAMPTZ;
ALTER TABLE branches ADD COLUMN IF NOT EXISTS sync_status TEXT NOT NULL DEFAULT 'synced';

-- Categories sync fields
ALTER TABLE categories ADD COLUMN IF NOT EXISTS pending_sync BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE categories ADD COLUMN IF NOT EXISTS synced_at TIMESTAMPTZ;
ALTER TABLE categories ADD COLUMN IF NOT EXISTS sync_status TEXT NOT NULL DEFAULT 'synced';

-- Products sync fields
ALTER TABLE products ADD COLUMN IF NOT EXISTS pending_sync BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE products ADD COLUMN IF NOT EXISTS synced_at TIMESTAMPTZ;
ALTER TABLE products ADD COLUMN IF NOT EXISTS sync_status TEXT NOT NULL DEFAULT 'synced';

-- Branch Products sync fields
ALTER TABLE branch_products ADD COLUMN IF NOT EXISTS pending_sync BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE branch_products ADD COLUMN IF NOT EXISTS synced_at TIMESTAMPTZ;
ALTER TABLE branch_products ADD COLUMN IF NOT EXISTS sync_status TEXT NOT NULL DEFAULT 'synced';

-- Transactions sync fields
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS pending_sync BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS synced_at TIMESTAMPTZ;
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS sync_status TEXT NOT NULL DEFAULT 'pending';

-- Transaction Items sync fields
ALTER TABLE transaction_items ADD COLUMN IF NOT EXISTS pending_sync BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE transaction_items ADD COLUMN IF NOT EXISTS synced_at TIMESTAMPTZ;
ALTER TABLE transaction_items ADD COLUMN IF NOT EXISTS sync_status TEXT NOT NULL DEFAULT 'pending';

-- Index for sync queries
CREATE INDEX IF NOT EXISTS idx_transactions_sync_status ON transactions(sync_status);
CREATE INDEX IF NOT EXISTS idx_transaction_items_sync ON transaction_items(sync_status);
CREATE INDEX IF NOT EXISTS idx_products_sync_status ON products(sync_status);
CREATE INDEX IF NOT EXISTS idx_categories_sync_status ON categories(sync_status);
CREATE INDEX IF NOT EXISTS idx_branches_sync_status ON branches(sync_status);
CREATE INDEX IF NOT EXISTS idx_users_sync_status ON users(sync_status);
