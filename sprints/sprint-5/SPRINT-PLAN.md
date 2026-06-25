# Sprint 5 — Polish + Realtime + Deployment

> Goal: Fitur lengkap, realtime antar cabang, siap deploy.

## Sprint Backlog

| # | Task | Assignee | Status |
|---|------|----------|--------|
| 1 | WebSocket realtime: push notifikasi transaksi & stok antar cabang | backend-dev | Not Started |
| 2 | Bug fixes dari QA Sprint 4 | backend-dev | Not Started |
| 3 | UI polish: loading state, error handling, empty state | frontend-dev | Not Started |
| 4 | Bug fixes dari QA Sprint 4 | frontend-dev | Not Started |
| 5 | Responsive layout + dark mode | frontend-dev | Not Started |
| 6 | Dockerfile + docker-compose for deployment | backend-dev | Not Started |
| 7 | QA regression test all flows | quality-assurance | Not Started |
| 8 | Retest bug fixes | quality-assurance | Not Started |

## API Contract (WebSocket)
- `WS /api/v1/ws` — realtime event stream
- Events: `transaction.created`, `stock.adjusted`, `stock.transferred`, `sync.required`

## Deliverables
- Docker image backend
- docker-compose.yml (backend + supabase)
- Flutter APK build ready
