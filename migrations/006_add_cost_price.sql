-- 006_add_cost_price.sql
-- POS Multi Branch — Add cost_price to products for P&L actual cost calculation

ALTER TABLE products ADD COLUMN IF NOT EXISTS cost_price NUMERIC(12,2) NOT NULL DEFAULT 0;

-- Update seed products with estimated cost_price (60% of selling price)
UPDATE products SET cost_price = ROUND(price * 0.6, 2) WHERE cost_price = 0 AND price > 0;
