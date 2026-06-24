# Task: Frontend — POS Multi Branch Retail (Flutter)

**Agent:** @tukangfrontendbot

## Stack
- Flutter (Android + Windows)
- HTTP client (dio atau http package)
- Supabase Realtime client

## PRD
Baca dulu: `PRD.md` di folder ini.

## Yang Harus Dikerjain

### ⚡ Arsitektur Local-First
Aplikasi harus jalan FULL offline. Data transaksi disimpan dulu lokal pakai **SQLite (Floor/drift)**, baru sync ke server saat online.

### Sprint 1 — Foundation & Master Data (Online-First)
1. **Setup Flutter project**
   - Multi-platform: Android + Windows
   - State management (Provider / Riverpod / BLoC — pilih salah satu)
   - Folder structure: `screens/`, `widgets/`, `services/`, `models/`, `providers/`, `database/`, `sync/`

2. **Auth Screens**
   - Login screen (email + password)
   - Supabase Auth integration
   - Token persistence (flutter_secure_storage)
   - Logout
   - Role-based navigation (owner multi-branch, kasir single branch)

3. **Branch Management** (owner only)
   - List cabang
   - Tambah/edit cabang
   - Aktif/nonaktifkan cabang

4. **Product Management** (per cabang)
   - List produk (per cabang)
   - Tambah produk (dengan barcode, harga, kategori)
   - Edit produk
   - Hapus produk
   - Search produk by nama/barcode

### Sprint 2 — POS Core (Online-First)
1. **POS Screen** (main transaction screen)
   - Search & add produk ke keranjang
   - Barcode scanner (gunakan package kamera atau hardware scanner)
   - Keranjang: list item, qty (+/-), hapus
   - Total per item & grand total
   - Diskon

2. **Checkout / Payment**
   - Pilih metode bayar (tunai, QRIS, transfer)
   - Input nominal bayar → hitung kembalian
   - Konfirmasi transaksi
   - Cetak struk via Bluetooth thermal printer (Android) / USB printer (Windows)

3. **Transaction History**
   - List transaksi per cabang
   - Detail transaksi
   - Cetak ulang struk

### Sprint 3 — Local Storage Engine 🏠
1. **Setup SQLite Local Database**
   - Package: Floor atau drift
   - Tabel lokal mirror Supabase + tambahan:
     - `pending_sync` boolean
     - `synced_at` timestamp
     - `sync_status` ('pending', 'synced', 'conflict')
   - DAOs untuk semua tabel

2. **Local Repository Layer**
   - Interface abstract repository (online & local impl)
   - Read selalu dari lokal dulu (cache)
   - Write: kalo online → API + update lokal. Kalo offline → lokal aja + flag pending_sync

3. **Connectivity Service**
   - Monitor koneksi (connectivity_plus)
   - Stream status koneksi ke seluruh app
   - Auto-trigger sync saat online

4. **Sync Engine**
   - Queue: ambil semua data dengan pending_sync=TRUE
   - Push ke API secara berurutan
   - Update status sync (success/conflict)
   - Pull master data terbaru dari server
   - Progress indicator buat user

5. **Conflict Resolution UI**
   - Notifikasi kalo ada konflik sync
   - Tampilkan data lokal vs server
   - Pilih: pake data lokal / server / manual

### Sprint 4 — Inventory & Reports
1. **Stock Management**
   - Stok masuk form
   - Stok keluar form
   - Riwayat mutasi stok
   - Minimum stok indicator

2. **Reports**
   - Sales report daily/weekly/monthly
   - Stock report
   - Profit & loss sederhana
   - Export PDF/Excel

## API Contract
Base URL: (sesuai dari backend deployment)
Prefix: `/api/v1`

Response format:
```json
{
  "status": "success",
  "data": {},
  "pagination": { "page": 1, "limit": 20, "total": 100, "totalPages": 5 }
}
```

Error:
```json
{
  "status": "error",
  "message": "something went wrong"
}
```

Auth header: `Authorization: Bearer <supabase-jwt>`

### Endpoints
Semua endpoint ada di PRD.md. Pastikan endpoint yang dipake sesuai sama yg udah disepakati.

## Priority
Sprint 1 → Sprint 2 → Sprint 3. Jangan lompat-lompat.

## Output
- Full Flutter project di folder `pos-multi-branch/frontend/`
- README dengan cara run & build

Kalau butuh klarifikasi API, tanya ke gua dulu biar gua koordinasiin ke backend. Gas 🔥
