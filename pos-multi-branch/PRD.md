# PRD — POS Multi Branch Retail

## 1. Ringkasan
Aplikasi POS (Point of Sale) untuk retail dengan dukungan multi cabang. Setiap cabang memiliki inventaris sendiri. Aplikasi berjalan di **Android** (mobile kasir) dan **Windows** (desktop kasir/admin).

## 2. Tech Stack
| Layer | Teknologi |
|-------|-----------|
| Frontend | Flutter (Android + Windows) |
| Backend | Go |
| Database & Auth | Supabase (PostgreSQL + Auth + Realtime) |

## 3. Fitur Core (Rilis 1)

### 3.1 Auth & Manajemen Cabang
- Login multi-level (owner, admin cabang, kasir)
- Setiap user terikat ke cabang tertentu atau all-access (owner)
- Switch cabang untuk role yang punya akses multi-cabang

### 3.2 Manajemen Produk
- CRUD produk per cabang (produk bisa beda tiap cabang)
- Kategori produk
- Barcode/scan untuk transaksi
- Pricing bisa diatur per cabang (walaupun produk sama)

### 3.3 Transaksi Kasir (POS Core)
- Scan barcode / cari produk
- Keranjang belanja (tambah, kurang, hapus item)
- Diskon per item atau per transaksi
- Multiple metode bayar (tunai, QRIS, transfer, dll)
- Hitung kembalian
- Cetak struk (thermal printer / bluetooth printer Android, USB printer Windows)
- Riwayat transaksi

### 3.4 Manajemen Stok
- Stok masuk (penerimaan barang)
- Stok keluar (retur, rusak, hilang)
- Mutasi stok antar cabang
- Minimum stok alert

### 3.5 Laporan
- Laporan penjualan per cabang (harian, mingguan, bulanan)
- Laporan stok
- Laporan laba rugi sederhana
- Export PDF/Excel

## 4. Arsitektur — Local-First

Aplikasi harus tetap berfungsi penuh saat offline. Data disimpan lokal di device (SQLite via Floor/drift), dan sync ke Supabase saat koneksi tersedia.

### Sync Strategy
- **Optimistic writes**: Transaksi langsung diproses lokal, masuk antrian sync
- **Conflict resolution**: Last-write-wins (timestamp based), dengan flag konflik untuk manual review
- **Sync triggers**: Otomatis saat koneksi pulih + manual refresh button
- **Data scope**: Master data (produk, cabang, user) di-cache lokal; transaksi & mutasi stok di-prioritaskan untuk sync

### Sync Flow
```
┌──────────────┐    Online?     ┌──────────────┐
│  User Aksi   │ ──────────▶   │  Push ke API  │
└──────┬───────┘    Yes        └──────────────┘
       │ No
       ▼
┌──────────────────────┐
│  Simpan ke SQLite    │
│  (pending_sync=TRUE) │
└──────────┬───────────┘
           │ Koneksi pulih
           ▼
┌──────────────────────┐
│  Sync Engine:        │
│  - Push transaksi    │
│  - Pull master data  │
│  - Resolve conflict  │
└──────────────────────┘
```

### Tabel Lokal (SQLite)
- Sama dengan struktur Supabase + tambahan:
  - `pending_sync` boolean
  - `synced_at` timestamp
  - `sync_status` ('pending', 'synced', 'conflict')

## 5. Arsitektur Awal (High Level)
```
┌─────────────────────────────────────────────────┐
│               Flutter App (Android/Win)          │
│  ┌──────────┐  ┌──────────┐  ┌───────────────┐  │
│  │ POS Screen│  │ Inventory│  │ Reports       │  │
│  └────┬─────┘  └────┬─────┘  └──────┬────────┘  │
│       └──────────────┴───────────────┘           │
│                    │ HTTP/REST                    │
└────────────────────┼─────────────────────────────┘
                     │
┌────────────────────┼─────────────────────────────┐
│           Go API Server                          │
│  ┌──────────┐ ┌──────────┐ ┌──────────────────┐  │
│  │ Auth     │ │ POS      │ │ Inventory        │  │
│  │ Handler  │ │ Handler  │ │ Handler          │  │
│  └────┬─────┘ └────┬─────┘ └───────┬──────────┘  │
│       └────────────┴───────────────┘              │
│                    │                               │
│         ┌──────────┴──────────┐                   │
│         │  Supabase Client    │                   │
│         └──────────┬──────────┘                   │
└────────────────────┼─────────────────────────────┘
                     │
┌────────────────────┼─────────────────────────────┐
│            Supabase                               │
│  ┌──────────┐ ┌──────────┐ ┌──────────────────┐  │
│  │ Auth     │ │PostgreSQL│ │ Realtime         │  │
│  │ Service  │ │          │ │ (subscriptions)  │  │
│  └──────────┘ └──────────┘ └──────────────────┘  │
└──────────────────────────────────────────────────┘
```

Untuk development Sprint 1, cukup **online-first** dulu. Local-first masuk di Sprint 2 (setelah POS core jalan online).

## 6. API Endpoint (Initial)

### Auth
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/logout`
- `GET /api/v1/auth/me`

### Products
- `GET /api/v1/branches/:id/products`
- `POST /api/v1/branches/:id/products`
- `PUT /api/v1/branches/:id/products/:productId`
- `DELETE /api/v1/branches/:id/products/:productId`
- `GET /api/v1/products/search?q=&branch_id=`

### Transactions
- `POST /api/v1/branches/:id/transactions`
- `GET /api/v1/branches/:id/transactions`
- `GET /api/v1/branches/:id/transactions/:transactionId`

### Inventory / Stock
- `GET /api/v1/branches/:id/inventory`
- `POST /api/v1/branches/:id/inventory/adjustment`
- `POST /api/v1/inventory/transfer`

### Reports
- `GET /api/v1/branches/:id/reports/sales?start=&end=`
- `GET /api/v1/branches/:id/reports/stock`
- `GET /api/v1/branches/:id/reports/profit-loss?start=&end=`

## 7. Database Schema (Initial)

### Users
- id, email, password_hash, name, role (owner/admin/kasir), branch_id (nullable), created_at

### Branches
- id, name, code, address, phone, is_active, created_at

### Categories
- id, name, created_at

### Products
- id, category_id, name, barcode, unit, price (default), image_url, is_active, created_at

### Branch Products (pivot dengan inventory)
- id, branch_id, product_id, price (override), stock_qty, min_stock, created_at

### Transactions
- id, branch_id, cashier_id, total_amount, discount, payment_method, payment_amount, change_amount, status, created_at

### Transaction Items
- id, transaction_id, product_id, qty, price, subtotal

### Stock Mutations
- id, branch_id (source), product_id, type (in/out/transfer), qty, reference_id, notes, created_at

## 8. Prioritas Development

### Sprint 0 — Setup & Arsitektur
- Finalisasi arsitektur local-first
- Setup SQLite schema (mirror Supabase + sync fields)

### Sprint 1 — Foundation (Online-First)

- Setup Go project structure + Supabase
- Auth API + Flutter login screen
- CRUD Branch + CRUD Product + Branch Products
- Flutter: master data management
- Setup Go project structure + Supabase
- Auth API + Flutter login screen
- CRUD Branch + CRUD Product + Branch Products
- Flutter: master data management

### Sprint 2 — POS Core (Online-First)
- Transaksi API + Flutter POS screen
- Scanning barcode
- Keranjang & checkout
- Cetak struk

### Sprint 3 — Local Storage + Sync Engine
- Stok masuk/keluar/mutasi
- Laporan dasar
- Export

### Sprint 4 — Inventory & Reporting
- Stok masuk/keluar/mutasi
- Laporan dasar
- Export

### Sprint 5 — Polish
- Realtime (notifikasi stok minim, transaksi baru)
- Print thermal refinements
- Bug fixes & testing
- Conflict resolution UI
- Realtime (notifikasi stok minim, transaksi baru)
- Print thermal refinements
- Bug fixes & testing
