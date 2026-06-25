# Sprint 4 — Task Log

> Goal: Aplikasi tetap berfungsi penuh saat offline. Data lokal SQLite + sync engine ke Supabase.
> Started: 2026-06-25 | Completed: 2026-06-25

## Sprint Backlog

| # | Task | Assignee | Status | Notes |
|---|------|----------|--------|-------|
| 1 | Wire up sync endpoints di main.go | backend-dev | ✅ Done | Sync routes registered (+ moved ke protected routes) |
| 2 | Sync Engine: push transaksi + stok ke API | backend-dev | ✅ Done | Udah di `internal/sync/sync.go` + `handler/sync.go` |
| 3 | Sync Engine: pull master data dari API | backend-dev | ✅ Done | Udah di `handler/sync.go Pull()` |
| 4 | Conflict resolution (last-write-wins + manual review) | backend-dev | ✅ Done | Udah di `handler/sync.go Resolve()` |
| 5 | Flutter: SQLite lokal (sqflite) | frontend-dev | ✅ Done | 9 DAOs + local_database + schema mirror + sync fields |
| 6 | Flutter: pending sync queue + status UI | frontend-dev | ✅ Done | sync_service + sync_provider + sync_status_widget + SyncStatusScreen |
| 7 | Flutter: sync trigger (auto on reconnect + manual) | frontend-dev | ✅ Done | connectivity_service + SyncStatusIcon di AppBar + manual sync button |
| 8 | QA test sync flow + offline mode | quality-assurance | ✅ Done | Test case by @tukangqabot — 10/13 PASS + 3 partial (dlq, conflict dialog, load test) |

## Files Produced

### Backend
- `cmd/server/main.go` — wiring sync routes (protected)
- `internal/database/sqlite.go` — SQLite schema (8 tables + sync_queue)
- `internal/handler/sync.go` — Push, Pull, Resolve handlers
- `internal/sync/sync.go` — Sync engine

### Frontend
- `lib/database/local_database.dart` — SQLite singleton + migration
- `lib/database/daos/` — 9 DAOs (base, product, branch_product, transaction, transaction_item, stock_mutation, user, branch, category, sync_queue)
- `lib/services/sync_service.dart` — Sync engine (push/pull/conflict)
- `lib/services/connectivity_service.dart` — Connectivity monitor
- `lib/providers/sync_provider.dart` — Provider untuk UI
- `lib/widgets/sync_status_widget.dart` — 4 widgets (icon, card, conflict tile, full screen)
