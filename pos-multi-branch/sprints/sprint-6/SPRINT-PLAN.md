# Sprint 6 — Payment Methods + Printer + Stock Alerts

> Goal: Fitur pembayaran lengkap, cetak struk dari device, alert stok minim.
> ✅ Selesai

## Sprint Backlog

| # | Task | Assignee | Status |
|---|------|----------|--------|
| 1 | **Multi Payment Method** — tambah metode bayar (QRIS, transfer, edc) + UI selector di checkout | frontend-dev | ✅ Done |
| 2 | **Multi Payment Method** — backend: terima payment_method & reference_number di transaksi | backend-dev | ✅ Done |
| 3 | **Print Receipt** — thermal printer integration (esc_pos_thermal package) + test print | frontend-dev | ✅ Done |
| 4 | **Stock Minimum Alert** — backend: GET /inventory/low-stock, alert pas stok <= min_stock | backend-dev | ✅ Done |
| 5 | **Stock Minimum Alert** — frontend: badge notif + low stock list screen | frontend-dev | ✅ Done |
| 6 | **Export PDF** — backend: PDF report endpoint (go-pdf / maroto) | backend-dev | ✅ Done |
| 7 | **Export PDF** — frontend: download/export button + share | frontend-dev | ✅ Done |
| 8 | **QA Regression** — test all flows + retry bug fixes Sprint 5 | quality-assurance | ✅ Done |

## API Contract (New/Changes)

### Payment Methods
```json
// POST /api/v1/branches/:id/transactions
{
  "items": [...],
  "payment_method": "cash|qris|transfer|edc",
  "payment_reference": "optional-ref-number",
  "tax_rate": 10
}
```

### Low Stock
```
GET /api/v1/branches/:id/inventory/low-stock?threshold=5
→ [{ "product_name": "...", "branch_name": "...", "stock_qty": 3, "min_stock": 5 }]
```

### PDF Export
```
GET /api/v1/branches/:id/reports/sales.pdf?start=&end=
→ application/pdf
```

## Deliverables
- Multi payment method di checkout
- Cetak struk dari thermal printer (bluetooth Android)
- Low stock alert + notification screen
- Export PDF
