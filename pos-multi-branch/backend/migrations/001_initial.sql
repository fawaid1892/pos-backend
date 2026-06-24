-- 001_initial.sql
-- POS Multi Branch — Initial Schema

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ─── Users ───
CREATE TABLE users (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username   VARCHAR(100) NOT NULL UNIQUE,
    password   TEXT NOT NULL,
    full_name  VARCHAR(200) NOT NULL DEFAULT '',
    role       VARCHAR(20) NOT NULL DEFAULT 'kasir' CHECK (role IN ('admin','kasir','owner')),
    branch_id  UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Branches ───
CREATE TABLE branches (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       VARCHAR(200) NOT NULL,
    address    TEXT NOT NULL DEFAULT '',
    phone      VARCHAR(30) NOT NULL DEFAULT '',
    is_active  BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- ─── Categories ───
CREATE TABLE categories (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       VARCHAR(200) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Products ───
CREATE TABLE products (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category_id UUID NOT NULL REFERENCES categories(id),
    name        VARCHAR(300) NOT NULL,
    barcode     VARCHAR(100) NOT NULL UNIQUE,
    price       NUMERIC(12,2) NOT NULL DEFAULT 0,
    stock       INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_products_name ON products USING gin (name gin_trgm_ops);
CREATE INDEX idx_products_barcode ON products (barcode);

-- ─── Transactions ───
CREATE TABLE transactions (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    branch_id        UUID NOT NULL REFERENCES branches(id),
    user_id          UUID NOT NULL REFERENCES users(id),
    customer_name    VARCHAR(200) NOT NULL DEFAULT '',
    subtotal         NUMERIC(12,2) NOT NULL DEFAULT 0,
    discount_percent NUMERIC(5,2) NOT NULL DEFAULT 0,
    discount_amount  NUMERIC(12,2) NOT NULL DEFAULT 0,
    total            NUMERIC(12,2) NOT NULL DEFAULT 0,
    cash_amount      NUMERIC(12,2) NOT NULL DEFAULT 0,
    change_amount    NUMERIC(12,2) NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_branch ON transactions (branch_id);
CREATE INDEX idx_transactions_created ON transactions (created_at DESC);

-- ─── Transaction Items ───
CREATE TABLE transaction_items (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id   UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    product_id       UUID NOT NULL REFERENCES products(id),
    product_name     VARCHAR(300) NOT NULL,
    quantity         INT NOT NULL,
    price            NUMERIC(12,2) NOT NULL,
    subtotal         NUMERIC(12,2) NOT NULL
);

-- ─── Seed Data ───
-- Password untuk semua seed user: "password123" (bcrypt hash)
INSERT INTO users (username, password, full_name, role, branch_id) VALUES
('admin',    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Admin Utama', 'admin', NULL),
('kasir1',   '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Kasir Cabang 1', 'kasir', NULL),
('owner',    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Pemilik Toko', 'owner', NULL);

INSERT INTO branches (name, address, phone) VALUES
('Cabang Pusat', 'Jl. Merdeka No.1, Jakarta', '021-1234567'),
('Cabang Cibubur', 'Jl. Cibubur Raya No.5, Bekasi', '021-7654321');

INSERT INTO categories (name) VALUES
('Makanan'), ('Minuman'), ('Snack'), ('Alat Tulis');

INSERT INTO products (category_id, name, barcode, price, stock) VALUES
((SELECT id FROM categories WHERE name='Makanan'), 'Nasi Goreng', '8991001001001', 15000, 50),
((SELECT id FROM categories WHERE name='Minuman'), 'Air Mineral 600ml', '8991001001002', 5000, 100),
((SELECT id FROM categories WHERE name='Snack'), 'Keripik Singkong', '8991001001003', 8000, 75),
((SELECT id FROM categories WHERE name='Alat Tulis'), 'Pulpen Standard', '8991001001004', 3000, 200);
