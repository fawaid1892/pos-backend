# Sprint 4 — Local Storage + Sync Engine (Local-First)

> Goal: Aplikasi tetap berfungsi penuh saat offline. Data lokal SQLite + sync engine ke Supabase.

---

## Sprint Backlog

| # | Task | Assignee | Status |
|---|------|----------|--------|
| 1 | SQLite schema mirror Supabase + sync fields | backend-dev | Not Started |
| 2 | Sync Engine: push transaksi + stok ke API | backend-dev | Not Started |
| 3 | Sync Engine: pull master data dari API | backend-dev | Not Started |
| 4 | Conflict resolution (last-write-wins + manual review) | backend-dev | Not Started |
| 5 | Flutter: SQLite lokal (drift/floor) | frontend-dev | Not Started |
| 6 | Flutter: pending sync queue + status UI | frontend-dev | Not Started |
| 7 | Flutter: sync trigger (auto on reconnect + manual) | frontend-dev | Not Started |
| 8 | QA test sync flow + offline mode | quality-assurance | Not Started |

## API Contract

### Sync Endpoints
- `POST /api/v1/sync/push` — push pending transactions & stock mutations
- `GET /api/v1/sync/pull?since=` — pull master data updates since timestamp
- `POST /api/v1/sync/resolve` — manual conflict resolution

### SQLite Local Schema
Same as Supabase + additional fields:
- `pending_sync` boolean
- `synced_at` timestamp
- `sync_status` ('pending', 'synced', 'conflict')

## Flow
```
User action offline → SQLite (pending_sync=TRUE)
Koneksi pulih → Sync Engine push pending data → API
Master data di-cache lokal
Conflict → flag manual review
```
