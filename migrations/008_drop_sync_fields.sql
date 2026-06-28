-- Drop sync fields — ElectricSQL handle sync otomatis
ALTER TABLE users DROP COLUMN IF EXISTS pending_sync, DROP COLUMN IF EXISTS synced_at, DROP COLUMN IF EXISTS sync_status;
ALTER TABLE branches DROP COLUMN IF EXISTS pending_sync, DROP COLUMN IF EXISTS synced_at, DROP COLUMN IF EXISTS sync_status;
ALTER TABLE categories DROP COLUMN IF EXISTS pending_sync, DROP COLUMN IF EXISTS synced_at, DROP COLUMN IF EXISTS sync_status;
ALTER TABLE products DROP COLUMN IF EXISTS pending_sync, DROP COLUMN IF EXISTS synced_at, DROP COLUMN IF EXISTS sync_status;
ALTER TABLE transactions DROP COLUMN IF EXISTS pending_sync, DROP COLUMN IF EXISTS synced_at, DROP COLUMN IF EXISTS sync_status;
ALTER TABLE transaction_items DROP COLUMN IF EXISTS pending_sync, DROP COLUMN IF EXISTS synced_at, DROP COLUMN IF EXISTS sync_status;
ALTER TABLE branch_products DROP COLUMN IF EXISTS pending_sync, DROP COLUMN IF EXISTS synced_at, DROP COLUMN IF EXISTS sync_status;
DROP TABLE IF EXISTS sync_queue;
