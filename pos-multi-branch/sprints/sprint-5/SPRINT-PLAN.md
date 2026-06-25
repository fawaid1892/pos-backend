# Sprint 5 — Polish + Realtime + Deployment

> Goal: Fitur lengkap, realtime antar cabang, siap deploy.
> ✅ Selesai

## Sprint Backlog

| # | Task | Assignee | Status |
|---|------|----------|--------|
| 1 | WebSocket realtime: push notifikasi transaksi & stok antar cabang | backend-dev | ✅ Done |
| 2 | Bug fixes dari QA Sprint 4 | backend-dev | ✅ Done |
| 3 | UI polish: loading state, error handling, empty state | frontend-dev | ✅ Done |
| 4 | Bug fixes dari QA Sprint 4 | frontend-dev | ✅ Done |
| 5 | Responsive layout + dark mode | frontend-dev | ✅ Done |
| 6 | Dockerfile + docker-compose for deployment | backend-dev | ✅ Done |
| 7 | QA regression test all flows | quality-assurance | ✅ Done |
| 8 | Retest bug fixes | quality-assurance | ✅ Done |

## API Contract (WebSocket)
- `WS /api/v1/ws` — realtime event stream
- Events: `transaction.created`, `stock.adjusted`, `stock.transferred`, `sync.required`

## Deliverables
- Docker image backend
- docker-compose.yml (backend + supabase)
- Flutter APK build ready

## Critical Bugs Fixed (Sprint 4 Cleanup)
- SQL injection prevention (whitelist table/column)
- Conflict resolution dialog UI (frontend)
- Dead letter queue UI (frontend)
- Tax calculation implementation
- XLSX export (excelize/v2)
- SyncService real HTTP client
