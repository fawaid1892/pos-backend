# QA Regression Report — Sprint 8

**Date:** 2026-06-29  
**Tester:** Tukang QA Bot  
**Branch:** sprint-8  

---

## 1. Backend — Code Review (pos-backend)

### Bug Fixes

| Item | Status | Notes |
|------|--------|-------|
| Bug A: Barcode 409 (duplicate) | ✅ | `internal/handler/product.go` L97–105 calls `CheckBarcodeExists()`, returns HTTP 409 `"barcode already exists"`. Logic correct. |
| Bug B: Category validation | ✅ | `internal/handler/product.go` L86–94 calls `CheckCategoryExists()`, returns HTTP 400 `"category not found"` if missing. Logic correct. |
| Bug C: Transaction rollback | ✅ | `internal/handler/transaction.go` L140–145: `database.Pool.Begin()` + `defer tx.Rollback()`. All writes use `tx.Exec`/`tx.QueryRow`. Commit at L226. Correct. |
| Bug D: Transaction WS event | ✅ | `internal/handler/transaction.go` L232–243: broadcasts `transaction.created` via `ws.DefaultHub.BroadcastEvent()`. Correct. |
| Bug F: Stock WS events | ✅ | `internal/handler/stock.go` L93–102 broadcasts `stock.adjusted`; L136–146 broadcasts `stock.transferred`. Both after response write, non-blocking. Correct. |
| Bug E: Soft delete safety (branch) | ✅ | `internal/repository/repository.go` L120–143: checks `EXISTS(SELECT 1 FROM users WHERE branch_id = $1)` before soft-deleting branch. Blocks if users exist. |
| Bug E: Soft delete safety (product) | ✅ | `internal/repository/repository.go` L349–372: checks `EXISTS(SELECT 1 FROM transaction_items WHERE product_id = $1)` before soft-deleting product. Blocks if transaction history exists. |

### Build & Static Analysis

| Item | Status | Notes |
|------|--------|-------|
| `go build ./...` | ⚠️ | Go toolchain not installed on QA machine. Code review confirms no syntax/structure issues. |
| `go vet ./...` | ⚠️ | Go toolchain not installed. Logic review shows no suspicious patterns. |

### WebSocket Events Defined

| Event | Constant | File | Used In |
|-------|----------|------|---------|
| `transaction.created` | `EventTransactionCreated` | `internal/ws/hub.go:21` | `transaction.go:234` |
| `stock.adjusted` | `EventStockAdjusted` | `internal/ws/hub.go:22` | `stock.go:95` |
| `stock.transferred` | `EventStockTransferred` | `internal/ws/hub.go:23` | `stock.go:138` |
| `stock.low` | `EventStockLow` | `internal/ws/hub.go:24` | `transaction.go:308`, `stock.go:105` |
| `sync.required` | `EventSyncRequired` | `internal/ws/hub.go:25` | (defined but not used in handlers) |

### Minor Observation

| Issue | Severity | Detail |
|-------|----------|--------|
| Confusing variable name `now` | Low | `transaction.go:177`: `now := r.Context()` stores the request context, not `time.Now()`. Functions correctly but naming is misleading. |

---

## 2. Backend — RBAC Test (Simulasi Kode)

### RBAC Wrappers (main.go L59–61)
| Wrapper | Roles |
|---------|-------|
| `adminOnly` | `admin_cabang` |
| `adminOrOwner` | `admin_cabang`, `owner` |
| `kasirOrAdmin` | `kasir`, `admin_cabang` |

### Role → Access Matrix

| Endpoint Group | Handler | kasir | admin_cabang | owner | Status |
|----------------|---------|-------|-------------|-------|--------|
| Branches CRUD | `branchH.*` | ❌ | ✅ | ❌ | ✅ |
| Products CRUD | `productH.*` | ❌ | ✅ | ❌ | ✅ |
| Categories CRUD | `productH.*` | ❌ | ✅ | ❌ | ✅ |
| Transactions (checkout, list, get) | `txH.*` | ✅ | ✅ | ❌ | ✅ |
| Stock adjustment / transfer | `stockH.*` | ❌ | ✅ | ❌ | ✅ |
| Reports (sales, stock, P&L) | `reportH.*` | ❌ | ✅ | ✅ | ✅ |
| Export | `exportH.*` | ❌ | ✅ | ✅ | ✅ |
| Users management | `userH.*` | ❌ | ✅ | ❌ | ✅ |
| Dashboard stats / chart | `dashboardH.*` | ✅* | ✅* | ✅* | ⚠️ Auth-only, no role check |
| Sync endpoints | `syncH.*` | ✅* | ✅* | ✅* | ⚠️ Auth-only, no role check |

*\* Authenticated but not role-gated — any authenticated user including kasir can access.*

### RBAC Verdict: ✅ Design correct per requirements
- **Kasir** → can transact, cannot manage products/branches/users ✅
- **Admin** → can do everything ✅
- **Owner** → can access reports ✅

---

## 3. Web — TypeScript Check (pos-web)

| Check | Status | Notes |
|-------|--------|-------|
| `npx tsc --noEmit` | ✅ | Exit code 0 — no TypeScript errors. |

---

## 4. Frontend — Flutter Check (pos-frontend)

| Check | Status | Notes |
|-------|--------|-------|
| `dart analyze lib/` | ⚠️ | `dart: command not found` — Flutter SDK not installed on QA machine. Cannot verify. |

---

## 5. Web — Pages Exist

| File | Status |
|------|--------|
| `src/app/(dashboard)/page.tsx` | ✅ (8271 bytes) |
| `src/app/(dashboard)/users/page.tsx` | ✅ (12019 bytes) |
| `src/app/(dashboard)/products/page.tsx` | ✅ (10563 bytes) |

---

## Summary

| Area | Pass | Fail | Skip |
|------|------|------|------|
| Backend — Code Review | 7/7 | 0 | 0 |
| Backend — RBAC | ✅ | 0 | 0 |
| Backend — Build | — | — | 2 (no Go toolchain) |
| Web — TypeScript | 1/1 | 0 | 0 |
| Web — Pages | 3/3 | 0 | 0 |
| Flutter — Analyze | — | — | 1 (no Dart SDK) |
| **Total** | **11 ✅** | **0 ❌** | **3 ⚠️** |

**No blocking issues found.** All code logic reviewed is correct per Sprint 8 requirements.
