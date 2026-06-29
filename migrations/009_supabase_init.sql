-- Supabase Migration: POS Multi Branch Retail
-- Run this in Supabase SQL Editor

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users
CREATE TABLE IF NOT EXISTS users (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username   VARCHAR(100) NOT NULL UNIQUE,
    password   TEXT NOT NULL,
    full_name  VARCHAR(200) NOT NULL DEFAULT '',
    role       VARCHAR(20) NOT NULL DEFAULT 'kasir' CHECK (role IN ('admin_cabang','kasir','owner')),
    branch_id  UUID,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Branches
CREATE TABLE IF NOT EXISTS branches (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       VARCHAR(200) NOT NULL,
    address    TEXT NOT NULL DEFAULT '',
    phone      VARCHAR(30) NOT NULL DEFAULT '',
    tax_rate   NUMERIC(5,2) NOT NULL DEFAULT 0,
    is_active  BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Categories
CREATE TABLE IF NOT EXISTS categories (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       VARCHAR(200) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Products
CREATE TABLE IF NOT EXISTS products (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category_id UUID REFERENCES categories(id),
    name        VARCHAR(300) NOT NULL,
    barcode     VARCHAR(100) NOT NULL UNIQUE,
    price       NUMERIC(12,2) NOT NULL DEFAULT 0,
    cost_price  NUMERIC(12,2) NOT NULL DEFAULT 0,
    stock       INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- Transactions
CREATE TABLE IF NOT EXISTS transactions (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    branch_id        UUID NOT NULL REFERENCES branches(id),
    user_id          UUID NOT NULL REFERENCES users(id),
    customer_name    VARCHAR(200) NOT NULL DEFAULT '',
    subtotal         NUMERIC(12,2) NOT NULL DEFAULT 0,
    discount_percent NUMERIC(5,2) NOT NULL DEFAULT 0,
    discount_amount  NUMERIC(12,2) NOT NULL DEFAULT 0,
    tax_rate         NUMERIC(5,2) NOT NULL DEFAULT 0,
    tax_amount       NUMERIC(15,2) NOT NULL DEFAULT 0,
    total            NUMERIC(12,2) NOT NULL DEFAULT 0,
    cash_amount      NUMERIC(12,2) NOT NULL DEFAULT 0,
    change_amount    NUMERIC(12,2) NOT NULL DEFAULT 0,
    payment_method   VARCHAR(20) NOT NULL DEFAULT 'cash' CHECK (payment_method IN ('cash','qris','transfer','edc')),
    payment_reference VARCHAR(200),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Transaction Items
CREATE TABLE IF NOT EXISTS transaction_items (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id   UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    product_id       UUID NOT NULL REFERENCES products(id),
    product_name     VARCHAR(300) NOT NULL,
    quantity         INT NOT NULL,
    price            NUMERIC(12,2) NOT NULL,
    subtotal         NUMERIC(12,2) NOT NULL
);

-- Branch Products
CREATE TABLE IF NOT EXISTS branch_products (
    branch_id  UUID NOT NULL REFERENCES branches(id),
    product_id UUID NOT NULL REFERENCES products(id),
    stock_qty  NUMERIC(15,2) NOT NULL DEFAULT 0,
    min_stock  NUMERIC(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (branch_id, product_id)
);

-- Stock Mutations
CREATE TABLE IF NOT EXISTS stock_mutations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id   UUID NOT NULL REFERENCES branches(id),
    product_id  UUID NOT NULL REFERENCES products(id),
    type        VARCHAR(20) NOT NULL CHECK (type IN ('in','out','transfer_in','transfer_out')),
    qty         NUMERIC(15,2) NOT NULL,
    reference_id UUID,
    notes       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_products_name ON products USING gin (name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_products_barcode ON products (barcode);
CREATE INDEX IF NOT EXISTS idx_transactions_branch ON transactions (branch_id);
CREATE INDEX IF NOT EXISTS idx_transactions_created ON transactions (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_branch_products_branch ON branch_products (branch_id);
CREATE INDEX IF NOT EXISTS idx_branch_products_product ON branch_products (product_id);

-- Seed Data
INSERT INTO users (username, password, full_name, role) VALUES
('admin', '$2a$10$NrTQTAYYBaeBAT/lrTzA1el4nTXlTUFktRnIXcqVHA6sRNMZOiima', 'Admin Utama', 'admin_cabang'),
('kasir1', '$2a$10$NrTQTAYYBaeBAT/lrTzA1el4nTXlTUFktRnIXcqVHA6sRNMZOiima', 'Kasir Cabang 1', 'kasir'),
('owner', '$2a$10$NrTQTAYYBaeBAT/lrTzA1el4nTXlTUFktRnIXcqVHA6sRNMZOiima', 'Pemilik Toko', 'owner')
ON CONFLICT (username) DO NOTHING;

INSERT INTO branches (name, address, phone) VALUES
('Cabang Pusat', 'Jl. Merdeka No.1, Jakarta', '021-1234567'),
('Cabang Cibubur', 'Jl. Cibubur Raya No.5, Bekasi', '021-7654321')
ON CONFLICT DO NOTHING;

INSERT INTO categories (name) VALUES
('Makanan'), ('Minuman'), ('Snack'), ('Alat Tulis')
ON CONFLICT (name) DO NOTHING;
