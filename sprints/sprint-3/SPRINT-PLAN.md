# Sprint 3 Plan — Inventory + Sync Engine (Local-First Prep)

> Goal: Stok masuk/keluar/mutasi antar cabang, laporan dasar, export PDF/Excel

---

## Sprint Backlog

| # | Task | Assignee | Status |
|---|------|----------|--------|
| 1 | Stock In API (penerimaan barang) | backend-dev | Not Started |
| 2 | Stock Out API (retur, rusak, hilang) | backend-dev | Not Started |
| 3 | Stock Transfer antar cabang | backend-dev | Not Started |
| 4 | Reports API (sales, stock, profit-loss) | backend-dev | Not Started |
| 5 | Export PDF/Excel endpoint | backend-dev | Not Started |
| 6 | Stock screens + stock opname UI | frontend-dev | Not Started |
| 7 | Reports screens + filter date | frontend-dev | Not Started |
| 8 | Export button (PDF/Excel) | frontend-dev | Not Started |
| 9 | QA test stock + report flow | quality-assurance | Not Started |

## API Contract

### Stock
- `POST /api/v1/branches/:id/inventory/adjustment` — stock in/out
- `POST /api/v1/inventory/transfer` — transfer antar cabang
- `GET /api/v1/branches/:id/inventory` — list stok + alert

### Reports
- `GET /api/v1/branches/:id/reports/sales?start=&end=` — sales report
- `GET /api/v1/branches/:id/reports/stock` — stock report
- `GET /api/v1/branches/:id/reports/profit-loss?start=&end=` — profit loss

### Export
- `GET /api/v1/branches/:id/reports/sales/export?format=pdf|xlsx`
