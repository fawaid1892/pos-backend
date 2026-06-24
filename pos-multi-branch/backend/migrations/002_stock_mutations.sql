-- 002_stock_mutations.sql
-- POS Multi Branch — Stock & Reports Schema

-- ─── Branch Products (per-branch stock tracking) ───
CREATE TABLE IF NOT EXISTS branch_products (
    branch_id  UUID NOT NULL REFERENCES branches(id),
    product_id UUID NOT NULL REFERENCES products(id),
    stock_qty  NUMERIC(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (branch_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_branch_products_branch ON branch_products (branch_id);
CREATE INDEX IF NOT EXISTS idx_branch_products_product ON branch_products (product_id);

-- ─── Stock Mutations ───
CREATE TABLE IF NOT EXISTS stock_mutations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id   UUID NOT NULL REFERENCES branches(id),
    product_id  UUID NOT NULL REFERENCES products(id),
    type        VARCHAR(20) NOT NULL CHECK (type IN ('in', 'out', 'transfer_in', 'transfer_out')),
    qty         NUMERIC(15,2) NOT NULL,
    reference_id UUID,
    notes       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_stock_mutations_branch ON stock_mutations (branch_id);
CREATE INDEX IF NOT EXISTS idx_stock_mutations_product ON stock_mutations (product_id);
CREATE INDEX IF NOT EXISTS idx_stock_mutations_created ON stock_mutations (created_at DESC);

-- ─── Copy initial product stock to branch_products for existing branches ───
INSERT INTO branch_products (branch_id, product_id, stock_qty)
SELECT b.id, p.id, p.stock::NUMERIC
FROM branches b, products p
WHERE b.deleted_at IS NULL AND p.deleted_at IS NULL
ON CONFLICT (branch_id, product_id) DO NOTHING;
