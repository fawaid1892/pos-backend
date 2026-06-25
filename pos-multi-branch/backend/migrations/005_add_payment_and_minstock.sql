-- 005_add_payment_and_minstock.sql
-- POS Multi Branch — Multi Payment Method & Low Stock Alert

-- Add payment method and reference to transactions
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS payment_method VARCHAR(20) NOT NULL DEFAULT 'cash'
  CHECK (payment_method IN ('cash','qris','transfer','edc'));
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS payment_reference VARCHAR(200);

-- Add min_stock threshold to branch_products for low-stock alerts
ALTER TABLE branch_products ADD COLUMN IF NOT EXISTS min_stock NUMERIC(15,2) NOT NULL DEFAULT 0;
