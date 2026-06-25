-- 004_add_tax.sql
-- POS Multi Branch — Add tax/pajak support

-- Add tax_rate to branches (per-branch config)
ALTER TABLE branches ADD COLUMN IF NOT EXISTS tax_rate NUMERIC(5,2) NOT NULL DEFAULT 0;

-- Add tax columns to transactions
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS tax_rate NUMERIC(5,2) NOT NULL DEFAULT 0;
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS tax_amount NUMERIC(15,2) NOT NULL DEFAULT 0;
