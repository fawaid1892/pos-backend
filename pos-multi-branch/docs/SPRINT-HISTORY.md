# Sprint History — POS Multi Branch Retail

## Sprint 1 ✅ — Foundation
**Goal:** Setup proyek + Auth + Master Data
**Tanggal:** 24 Jun 2026

| Task | Assignee | Status |
|------|----------|--------|
| Setup Go + Supabase | Backend | ✅ |
| Auth API (login/logout/me) | Backend | ✅ |
| CRUD Branch API | Backend | ✅ |
| CRUD Product & Category API | Backend | ✅ |
| Flutter login screen + auth flow | Frontend | ✅ |
| Flutter master data (Branch, Product, Category) | Frontend | ✅ |
| QA test cases | QA | ✅ |

**Completion rate:** 7/7 (100%)

---

## Sprint 2 ✅ — POS Core
**Goal:** Transaksi + POS Screen + Barcode + Checkout
**Tanggal:** 24 Jun 2026

| Task | Assignee | Status |
|------|----------|--------|
| Transaksi API + checkout endpoint | Backend | ✅ |
| POS Screen + keranjang belanja | Frontend | ✅ |
| Scanning barcode | Frontend | ✅ |
| Checkout flow + metode bayar | Frontend | ✅ |
| Cetak struk (thermal) | Frontend | ✅ |
| QA test transaksi flow | QA | ✅ |

**Completion rate:** 6/6 (100%)

---

## Sprint 3 ✅ — Stock + Reports
**Goal:** Manajemen stok + laporan + export
**Tanggal:** 24 Jun 2026

| Task | Assignee | Status |
|------|----------|--------|
| Stock In API | Backend | ✅ |
| Stock Out API | Backend | ✅ |
| Stock Transfer antar cabang | Backend | ✅ |
| Reports API | Backend | ✅ |
| Export endpoint | Backend | ✅ |
| Stock screens + stock opname UI | Frontend | ✅ |
| Reports screens + filter date | Frontend | ✅ |
| Export button | Frontend | ✅ |
| QA test stock + report flow | QA | ✅ |

**Completion rate:** 9/9 (100%)

---

## Sprint 4 ✅ — Local-First + Sync
**Goal:** Aplikasi offline + sync engine
**Tanggal:** 25 Jun 2026

| Task | Assignee | Status |
|------|----------|--------|
| Wire up sync endpoints | Backend | ✅ |
| Sync Engine: push transaksi + stok | Backend | ✅ |
| Sync Engine: pull master data | Backend | ✅ |
| Conflict resolution | Backend | ✅ |
| Flutter SQLite lokal (sqflite) 9 DAOs | Frontend | ✅ |
| Pending sync queue + status UI | Frontend | ✅ |
| Auto sync on reconnect + manual button | Frontend | ✅ |
| QA test sync flow + offline mode (10/13 PASS) | QA | ✅ |

**Completion rate:** 8/8 (100%)

**Cleanup (dikerjain pas Sprint 5):**
- SQL injection prevention (whitelist table/column)
- Conflict resolution dialog UI
- Dead letter queue UI (retry gagal sync)
- Tax calculation
- XLSX export fix (excelize/v2)
- SyncService real HTTP client

---

## Sprint 5 ✅ — Polish + Realtime + Deployment
**Goal:** Fitur lengkap, realtime antar cabang, siap deploy
**Tanggal:** 25 Jun 2026

| Task | Assignee | Status |
|------|----------|--------|
| WebSocket realtime (WS /api/v1/ws) | Backend | ✅ |
| Bug fixes dari QA Sprint 4 | Backend | ✅ |
| UI polish: loading, error, empty state | Frontend | ✅ |
| Bug fixes dari QA Sprint 4 | Frontend | ✅ |
| Responsive layout + dark mode | Frontend | ✅ |
| Dockerfile + docker-compose | Backend | ✅ |
| QA regression test all flows | QA | ✅ |
| Retest bug fixes | QA | ✅ |

**Completion rate:** 8/8 (100%)

**QA Report:** 95/107 PASS

---

## Sprint 6 ✅ — Payment + Printer + Alerts
**Goal:** Multi payment, cetak struk, alert stok minim
**Tanggal:** 25 Jun 2026

| Task | Assignee | Status |
|------|----------|--------|
| Multi Payment: selector QRIS/Transfer/EDC di checkout | Frontend | ✅ |
| Multi Payment: payment_method & reference_number | Backend | ✅ |
| Thermal print receipt (bluetooth Android + USB Windows) | Frontend | ✅ |
| Low stock alert endpoint + WS broadcast | Backend | ✅ |
| Low stock alert screen + badge notif | Frontend | ✅ |
| PDF export (gofpdf) | Backend | ✅ |
| PDF export button + share | Frontend | ✅ |
| QA regression test | QA | ✅ |

**Completion rate:** 8/8 (100%)

---

## Sprint 7 🟢 — Enhancement
**Goal:** User management, API docs, search improvement
**Target:** 10 Jul 2026

| # | Task | Assignee | Status |
|---|------|----------|--------|
| 1 | User Management — CRUD + roles | Backend | ⏳ Not Started |
| 2 | User Management — Admin panel UI | Frontend | ⏳ Not Started |
| 3 | Receipt customization — header/footer/font | Frontend | ⏳ Not Started |
| 4 | API Documentation — Swagger/OpenAPI | Backend | ⏳ Not Started |
| 5 | Product search — pagination + filter | Backend + Frontend | ⏳ Not Started |
| 6 | Regression test Sprint 7 | QA | ⏳ Not Started |
