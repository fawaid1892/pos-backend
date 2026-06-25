# Decision Log (ADR)

## ADR-001 — Local-First Architecture
**Date:** 2026-06-24
**Status:** Accepted
**Context:** POS harus tetap jalan saat offline (jaringan retail sering bermasalah).
**Decision:** Local-First — SQLite lokal sebagai source of truth, sync ke Supabase secara asinkron.
**Consequences:**
+ Aplikasi tetap berfungsi penuh saat offline
+ Performa lebih cepat (gak nunggu API)
- Perlu sync engine + conflict resolution
- Kompleksitas development naik

## ADR-002 — Tech Stack: Go + Supabase
**Date:** 2026-06-24
**Status:** Accepted
**Context:** Butuh backend yang ringan, cepat, dan mudah di-deploy.
**Decision:** Go (REST API + WebSocket) + Supabase (PostgreSQL + Auth + Realtime).
**Alternatives considered:**
1. Node.js/Express — lebih familiar tapi performa di bawah Go untuk concurrent requests
2. Laravel — terlalu berat untuk API POS
3. Firebase — vendor lock-in, pricing ga cocok untuk multi-branch
**Consequences:**
+ Performa tinggi
+ Supabase handle auth & database infra
- Perlu belajar Go untuk dev baru

## ADR-003 — Offline Sync Strategy: Optimistic Writes
**Date:** 2026-06-25
**Status:** Accepted
**Context:** User transaksi harus cepat, ga bisa nunggu validasi server setiap kali.
**Decision:** Optimistic writes — transaksi langsung di-proses lokal, masuk antrian sync, push ke server saat koneksi pulih.
**Conflict resolution:** Last-write-wins (by timestamp) + manual dialog untuk kasus conflict.
**Consequences:**
+ User experience tetap cepat meski offline
- Potensi conflict data perlu di-handle
- Dead letter queue untuk item gagal sync

## ADR-004 — Multi Payment + Tax
**Date:** 2026-06-25
**Status:** Accepted
**Context:** PRD nyebutin multi payment & pajak, tapi belum diimplementasi saat Sprint 1-5.
**Decision:** Implementasi saat Sprint 6 cleanup.
- Payment method: cash, qris, transfer, edc
- Tax: rate per cabang (Branch.TaxRate), dihitung pas checkout
**Consequences:**
+ Fitur sesuai PRD
+ Backward compatible (field nullable)

## ADR-005 — PDF Export: gofpdf
**Date:** 2026-06-25
**Status:** Accepted
**Context:** Butuh generate PDF untuk laporan.
**Decision:** gofpdf — lightweight, tanpa heavy dependencies. Alternatif: maroto (terlalu berat untuk kebutuhan sederhana).
**Consequences:**
+ Ukuran binary kecil
+ Cukup untuk laporan sederhana
- Kurang flexible untuk layout complex

## ADR-006 — Sprint Documentation: GitHub Issues
**Date:** 2026-06-25
**Status:** Accepted
**Context:** Dokumentasi sprint di file MD tidak terintegrasi dan perlu manual update.
**Decision:** Pindah ke GitHub Issues + Milestones + Labels.
**Consequences:**
+ Progress otomatis terlihat
+ Assign ke agent langsung dari GitHub
+ Labels untuk filtering
- Hilang dari repo lokal, tapi bisa diakses via GitHub API nanti
