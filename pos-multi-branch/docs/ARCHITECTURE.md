# POS Multi Branch Retail — Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                     FLUTTER APP (Android / Windows)              │
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐   │
│  │   POS Module  │  │  Stock Module│  │    Report Module     │   │
│  │ ───────────── │  │ ──────────── │  │ ──────────────────── │   │
│  │ • Scan produk │  │ • Stok masuk │  │ • Penjualan harian   │   │
│  │ • Keranjang   │  │ • Stok keluar│  │ • Stok               │   │
│  │ • Checkout    │  │ • Transfer   │  │ • Laba rugi          │   │
│  │ • Cetak struk │  │ • Low stock  │  │ • Export PDF/XLSX    │   │
│  └──────┬───────┘  └──────┬───────┘  └──────────┬───────────┘   │
│         │                 │                      │               │
│  ┌──────┴─────────────────┴──────────────────────┴───────────┐  │
│  │                LOCAL SQLITE DATABASE                       │  │
│  │  ┌──────────────────────────────────────────────────────┐  │  │
│  │  │ 9 DAOs: products, transactions, stock, sync_queue   │  │  │
│  │  │ Local DB mirror + sync_status (pending/synced)       │  │  │
│  │  └──────────────────────────────────────────────────────┘  │  │
│  └──────────────────────────┬────────────────────────────────┘  │
│                             │                                    │
│  ┌──────────────────────────┴────────────────────────────────┐  │
│  │                 SYNC ENGINE                               │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌───────────────┐  │  │
│  │  │ Push Service │  │ Pull Service │  │ Conflict      │  │  │
│  │  │ (transaksi,  │  │ (master data,│  │ Resolution    │  │  │
│  │  │  stock ops)  │  │  since=)     │  │ Dialog + DLQ  │  │  │
│  │  └──────────────┘  └──────────────┘  └───────────────┘  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                             │ HTTP REST + WS                   │
└─────────────────────────────┼─────────────────────────────────┘
                              │
┌─────────────────────────────┼─────────────────────────────────┐
│                   GO API SERVER (port 8080)                    │
│                                                                │
│  ┌──────────┐  ┌──────────┐  ┌────────────┐  ┌────────────┐  │
│  │  Auth    │  │   POS    │  │ Inventory  │  │   Report   │  │
│  │ Handler  │  │ Handler  │  │ Handler    │  │  Handler   │  │
│  │ ──────── │  │ ──────── │  │ ─────────  │  │ ─────────  │  │
│  │ Login    │  │ Checkout │  │ Adjustment │  │ Sales      │  │
│  │ Logout   │  │ History  │  │ Transfer   │  │ PDF/XLSX   │  │
│  │ Profile  │  │ Detail   │  │ Low Stock  │  │ Stock/Cost │  │
│  └────┬─────┘  └────┬────┘  └──────┬─────┘  └─────┬──────┘  │
│       │             │              │               │          │
│  ┌────┴─────────────┴──────────────┴───────────────┴──────┐  │
│  │                    WS HUB                               │  │
│  │  Events: transaction.created, stock.adjusted,          │  │
│  │          stock.transferred, stock.low, sync.required    │  │
│  └─────────────────────────┬──────────────────────────────┘  │
│                            │                                   │
│  ┌─────────────────────────┴──────────────────────────────┐  │
│  │                    SUPABASE CLIENT                       │  │
│  └─────────────────────────┬──────────────────────────────┘  │
└─────────────────────────────┼────────────────────────────────┘
                              │
┌─────────────────────────────┼────────────────────────────────┐
│                      SUPABASE CLOUD                          │
│                                                              │
│  ┌──────────────────┐  ┌──────────────────┐                  │
│  │   PostgreSQL DB   │  │   Auth Service   │                  │
│  │                  │  │                  │                  │
│  │ - users          │  │ - Login/Register │                  │
│  │ - branches       │  │ - JWT Token      │                  │
│  │ - products       │  │ - Row Level Sec. │                  │
│  │ - transactions   │  └──────────────────┘                  │
│  │ - stock_mutations│                                        │
│  └──────────────────┘                                        │
└──────────────────────────────────────────────────────────────┘


┌──────────────────────────────────────────────┐
│           REPOSITORY STRUCTURE                │
├──────────────────────────────────────────────┤
│                                              │
│  pos-multi-branch/ (PRIVATE — main repo)     │
│  ├── backend/           (Go API server)      │
│  └── pos_flutter/       (Flutter app)        │
│                                              │
│  pos-backend/   (PUBLIC — backend only)      │
│  pos-frontend/  (PUBLIC — frontend only)     │
│                                              │
├──────────────────────────────────────────────┤
│           BACKEND STRUCTURE                  │
├──────────────────────────────────────────────┤
│  cmd/server/main.go         — Entry point    │
│  internal/                                   │
│  ├── config/config.go       — Config loader  │
│  ├── database/sqlite.go     — SQLite schema  │
│  ├── handler/               — HTTP handlers  │
│  │   ├── auth.go                            │
│  │   ├── product.go                         │
│  │   ├── transaction.go                     │
│  │   ├── stock.go                           │
│  │   ├── sync.go                            │
│  │   └── report.go                          │
│  ├── middleware/            — Auth middleware│
│  ├── model/models.go       — Data models    │
│  ├── repository/           — Supabase queries│
│  ├── sync/sync.go          — Sync engine    │
│  └── ws/                   — WebSocket hub  │
│      ├── hub.go                              │
│      └── handler.go                          │
│  migrations/               — SQL migrations  │
│  Dockerfile                — Multi-stage     │
│  docker-compose.yml        — Deploy config   │
├──────────────────────────────────────────────┤
│           FRONTEND STRUCTURE                 │
├──────────────────────────────────────────────┤
│  lib/                                        │
│  ├── database/                               │
│  │   ├── local_database.dart — SQLite init   │
│  │   └── daos/              — 9 DAO files    │
│  ├── models/                — Data models    │
│  ├── providers/             — State mgmt     │
│  │   ├── auth_provider.dart                  │
│  │   ├── cart_provider.dart                  │
│  │   ├── sync_provider.dart                  │
│  │   └── theme_provider.dart                 │
│  ├── screens/                                 │
│  │   ├── login_screen.dart                   │
│  │   ├── pos_screen.dart                     │
│  │   ├── checkout_screen.dart                │
│  │   ├── stock_screen.dart                   │
│  │   ├── report_screen.dart                  │
│  │   ├── low_stock_alert_screen.dart         │
│  │   └── ...                                 │
│  ├── services/                               │
│  │   ├── sync_service.dart                   │
│  │   ├── thermal_print_service.dart          │
│  │   ├── pdf_export_service.dart             │
│  │   └── ...                                 │
│  └── widgets/                                │
│      ├── sync_status_widget.dart             │
│      ├── conflict_resolution_dialog.dart     │
│      ├── dead_letter_queue_widget.dart       │
│      ├── shimmer_loading.dart                │
│      └── ...                                 │
└──────────────────────────────────────────────┘
```
