-- 007_user_admin_cabang.sql
-- POS Multi Branch — Update roles: admin -> admin_cabang, add soft delete fields

-- Update CHECK constraint to use admin_cabang instead of admin
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check
  CHECK (role IN ('admin_cabang', 'kasir', 'owner'));

-- Add soft delete fields
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_deleted ON users(deleted_at);

-- Update existing admin users to admin_cabang
UPDATE users SET role = 'admin_cabang' WHERE role = 'admin';
