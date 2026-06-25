# PRD — POS Multi Branch Retail

## 1. Ringkasan
Aplikasi POS (Point of Sale) untuk retail dengan dukungan multi cabang.
Setiap cabang memiliki inventaris sendiri. Aplikasi berjalan di **Android** (mobile kasir)
dan **Windows** (desktop kasir/admin).

## 2. Tech Stack
| Layer | Teknologi |
|-------|-----------|
| Frontend | Flutter (Android + Windows) |
| Backend | Go |
| Database & Auth | Supabase (PostgreSQL + Auth + Realtime) |
| Local DB | SQLite (sqflite) |
| Realtime | WebSocket (gorilla/websocket) |

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
- Multi payment method (Tunai, QRIS, Transfer, EDC)
- Pajak per transaksi (tax rate dari cabang config)
- Hitung kembalian
- Cetak struk (thermal printer / bluetooth Android, USB Windows)
- Riwayat transaksi

### 3.4 Manajemen Stok
- Stok masuk (penerimaan barang)
- Stok keluar (retur, rusak, hilang)
- Mutasi stok antar cabang
- Minimum stok alert + notifikasi
- Low stock screen + badge notif

### 3.5 Laporan
- Laporan penjualan per cabang (harian, mingguan, bulanan)
- Laporan stok
- Laporan laba rugi sederhana
- Export PDF & Excel (XLSX)
- PDF: sales report, stock report, profit-loss report

### 3.6 Realtime
- WebSocket: notifikasi transaksi baru, stok berubah, sync required
- Event: `transaction.created`, `stock.adjusted`, `stock.transferred`, `sync.required`, `stock.low`

### 3.7 Sync & Offline
- Local-first architecture
- SQLite lokal mirror database
- Optimistic writes: transaksi langsung diproses lokal
- Sync otomatis saat koneksi pulih + manual button
- Conflict resolution: last-write-wins + manual review (dialog)
- Dead letter queue: retry item gagal sync

## 4. Arsitektur

### High Level Architecture
```
┌─────────────────────────────────────────────────┐
│               Flutter App (Android/Win)          │
│  ┌──────────┐  ┌──────────┐  ┌───────────────┐  │
│  │ POS Screen│  │ Inventory│  │ Reports       │  │
│  └────┬─────┘  └────┬─────┘  └──────┬────────┘  │
│       └──────────────┴───────────────┘           │
│                    │ HTTP/REST + WS               │
└────────────────────┼─────────────────────────────┘
                     │
┌────────────────────┼─────────────────────────────┐
│           Go API Server                          │
│  ┌──────────┐ ┌──────────┐ ┌──────────────────┐  │
│  │ Auth     │ │ POS      │ │ Inventory        │  │
│  │ Handler  │ │ Handler  │ │ Handler          │  │
│  └────┬─────┘ └────┬─────┘ └───────┬──────────┘  │
│       │            │               │              │
│  ┌────┴────┐ ┌─────┴──────┐ ┌──────┴──────────┐  │
│  │ WS Hub  │ │ Sync      │ │ Export (PDF/XLSX)│  │
│  └─────────┘ └───────────┘ └─────────────────┘  │
│                    │                              │
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

### Sync Flow (Local-First)
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

## 5. API Endpoints

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
- `GET /api/v1/branches/:id/inventory/low-stock?threshold=5`

### Sync
- `POST /api/v1/sync/push`
- `GET /api/v1/sync/pull?since=`
- `POST /api/v1/sync/resolve`

### Reports
- `GET /api/v1/branches/:id/reports/sales?start=&end=`
- `GET /api/v1/branches/:id/reports/stock`
- `GET /api/v1/branches/:id/reports/profit-loss?start=&end=`
- `GET /api/v1/branches/:id/reports/sales.pdf?start=&end=`
- `GET /api/v1/branches/:id/reports/sales.xlsx?start=&end=`

### WebSocket
- `WS /api/v1/ws`

## 6. Database Schema

### Users
- id, email, password_hash, name, role (owner/admin/kasir), branch_id (nullable), created_at

### Branches
- id, name, code, address, phone, tax_rate, is_active, created_at

### Categories
- id, name, created_at

### Products
- id, category_id, name, barcode, unit, price (default), image_url, is_active, created_at

### Branch Products (pivot with stock)
- id, branch_id, product_id, price (override), stock_qty, min_stock, created_at

### Transactions
- id, branch_id, cashier_id, subtotal, discount_percent, discount_amount, tax_rate, tax_amount, total, payment_method, payment_reference, cash_amount, change_amount, status, created_at

### Transaction Items
- id, transaction_id, product_id, qty, price_per_unit, subtotal

### Stock Mutations
- id, branch_id (source), product_id, type (in/out/transfer), qty, reference_id, notes, created_at

## 7. Sprint History

| Sprint | Fokus | Status |
|--------|-------|--------|
| 1 | Foundation — Setup, Auth, Master Data | ✅ Selesai |
| 2 | POS Core — Transaksi, Barcode, Checkout | ✅ Selesai |
| 3 | Stock + Reports | ✅ Selesai |
| 4 | Local-First + Sync Engine | ✅ Selesai |
| 5 | Polish — WebSocket, UI, Docker, Bugfixes | ✅ Selesai |
| 6 | Payment — Multi Payment, Thermal Print, Alerts, PDF | ✅ Selesai |
| 7 | Enhancement — User Management, API Docs, Search | 🟢 Akan datang |
