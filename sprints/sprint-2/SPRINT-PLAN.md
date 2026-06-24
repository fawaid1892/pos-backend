# Sprint 2 Plan — POS Core (Online-First)

> **Periode:** [Start Date] — [End Date]
> **Goal:** Transaksi kasir, scan barcode, keranjang belanja, checkout, cetak struk

---

## Sprint Backlog

| # | Task | Assignee | Status |
|---|------|----------|--------|
| 1 | Transaksi API + endpoint checkout | backend-dev | Not Started |
| 2 | POS Screen + keranjang belanja (mock) | frontend-dev | Not Started |
| 3 | Scanning barcode | frontend-dev | Not Started |
| 4 | Checkout flow + metode bayar | frontend-dev | Not Started |
| 5 | Cetak struk (thermal) | frontend-dev | Not Started |
| 6 | QA test transaksi flow | quality-assurance | Not Started |

## Phase 1 (Hybrid)
Jalan barengan setelah API contract deal:
- Backend: transaksi API
- Frontend: POS screen pake mock data

## Phase 2 (Sequential)
Setelah backend & frontend selesai:
- QA test semua flow
- Bug fix loop kalo ada

---

## API Contract

### Transactions
- `POST /api/v1/branches/:branchId/transactions` — create transaction
  - Body: { items: [{product_id, qty, price}], discount, payment_method, payment_amount }
  - Response: { id, total_amount, change_amount, ... }
- `GET /api/v1/branches/:branchId/transactions` — list
- `GET /api/v1/branches/:branchId/transactions/:id` — detail

### Products (existing)
- `GET /api/v1/branches/:branchId/products/search?q=` — search by barcode/name

### Stock
- `GET /api/v1/branches/:branchId/products` — include stock_qty
