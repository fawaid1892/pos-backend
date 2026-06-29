# BUG_REPORT.md — QA Retest Sprint 2 (BE #2)

**Date**: 2026-06-28
**Repo**: `/home/fakhriaz/projects/pos-backend`
**Head**: `49afbd8` (`feat(backend): Sprint 6 - multi payment + low stock + PDF export`)

> GitHub Issue #2 tidak dapat diakses (404 Not Found). Analisis dilakukan dengan membandingkan kode asli Sprint 2 (commit `bf60c15`) dengan kode terkini (`HEAD`).

---

## 📋 Sprint 2 Original Bugs — Status

| # | Bug Description | Original Sprint 2 | Current Fix | Status |
|---|----------------|-------------------|-------------|--------|
| 1 | **Tax implementation missing** — No `tax_rate` / `tax_amount` fields in Branch, Transaction, or checkout logic | ❌ Not present | ✅ Added via migration `004_add_tax.sql`; `tax_rate` on Branch, `tax_rate`+`tax_amount` on Transaction | **✅ Fixed** |
| 2 | **Cash-only payment** — `CashAmount <= 0` blocked ALL payment methods; no `payment_method` field | ❌ Only cash supported | ✅ `PaymentMethod`, `PaymentReference`, `CashAmount` fields added; supports cash, qris, transfer, edc | **✅ Fixed** |
| 3 | **No branch-level stock deduction** — Only deducted from global `products.stock`, missing `branch_products` deduction | ❌ Missing | ✅ `DeductBranchProductStock()` called after global stock deduction | **✅ Fixed** |
| 4 | **writeJSON could panic** — `json.NewEncoder(w).Encode(data)` without error check | ❌ No error handling | ✅ Error result captured with `_ = err` | **✅ Fixed** |
| 5 | **Branch route mapping bug** — No branch ID validation on GET/PUT/DELETE branch endpoints | ❌ Primitive | ✅ Proper `uuid.Parse()` validation on all branch endpoints | **✅ Fixed** |
| 6 | **No low stock detection / WebSocket broadcast** — After checkout, no low-stock alerts | ❌ Missing | ✅ `checkAndBroadcastLowStock()` goroutine called after checkout + adjustment | **✅ Fixed** |

---

## 🐛 New / Remaining Bugs Found (Current Code)

### Bug A: Barcode Duplicate Returns 500 Instead of 400/409

- **File**: `internal/handler/product.go` (Create method, line ~59-75)
- **Impact**: Creating a product with an existing barcode causes a PostgreSQL unique constraint violation → handler returns HTTP 500
- **Root Cause**: No explicit barcode uniqueness check before INSERT. DB UNIQUE constraint catches it but bubbles up as generic 500.
- **Steps to Reproduce**:
  1. POST `/api/v1/products` with `{"name":"Test","barcode":"8991001001001","price":10000,"stock":5}`
  2. POST again with same barcode
  3. → HTTP 500 instead of 400/409
- **Status**: ❌ **Still broken**

### Bug B: CategoryID Not Validated on Product Create

- **File**: `internal/handler/product.go` (Create method, line ~59-75)
- **Impact**: Creating a product with a non-existent `category_id` returns HTTP 500 instead of 400
- **Root Cause**: `CategoryID` is not validated to exist before the INSERT. DB foreign key constraint returns error as 500.
- **Steps to Reproduce**:
  1. POST `/api/v1/products` with `{"category_id":"00000000-0000-0000-0000-000000000000","name":"Test","barcode":"NEW123","price":10000,"stock":5}`
  2. → HTTP 500 instead of 400
- **Status**: ❌ **Still broken**

### Bug C: P&L Report Uses Hardcoded Cost Multiplier (0.7)

- **File**: `internal/repository/repository.go` (lines 668-669 in `GetProfitLossReport`)
- **Impact**: Profit/Loss calculation uses `p.price * 0.7` as cost instead of reading `cost_price` from products table
- **Root Cause**: Despite commit `bf24a8b` claiming to fix this, `cost_price` column doesn't exist in the Product model or schema. Code still hardcodes 0.7.
- **Evidence**:
  ```sql
  COALESCE(SUM(ti.quantity * p.price * 0.7), 0) as cost
  ```
- **Status**: ❌ **Still broken** (fix attempted but incomplete)

### Bug D: No Transaction Rollback on Partial Failure

- **File**: `internal/handler/transaction.go` (Checkout method, lines 134-180)
- **Impact**: If transaction creation or item insertion fails, stock has already been deducted with no rollback
- **Root Cause**: Stock deduction happens before transaction creation, and there's no database transaction wrapping the entire checkout flow
- **Steps to Reproduce**:
  1. Fill inventory: Product X has stock=1
  2. Checkout with Product X → stock deducted to 0
  3. If `CreateTransaction` fails (e.g., DB error), stock remains at 0
  4. Product X is now stuck at stock=0 with no way to recover
- **Status**: ❌ **Still broken**

### Bug E: Soft Delete Ignores Referential Integrity

- **Files**: `internal/repository/repository.go` (`SoftDeleteBranch`, `SoftDeleteProduct`)
- **Impact**: A branch/product can be soft-deleted even if it has associated transactions, users, or inventory
- **Root Cause**: No `WHERE NOT EXISTS` check or cascade validation before soft delete
- **Steps to Reproduce**:
  1. Create Branch A
  2. Create User with `branch_id = BranchA`
  3. DELETE `/api/v1/branches/{BranchA_id}` → succeeds
  4. User now references a "deleted" branch
- **Status**: ⚠️ **Partial** — Soft-delete pattern exists but lacks referential safety checks

### Bug F: No Role-Based Access Control (RBAC)

- **Files**: `internal/middleware/auth.go`, all handlers
- **Impact**: Any authenticated user (kasir, admin, owner) can access ALL endpoints — there's no role-based restriction
- **Root Cause**: `AuthMiddleware` only validates JWT validity, not role authorization. No handler checks the user's role.
- **Steps to Reproduce**:
  1. Login as `kasir1` → get JWT
  2. POST `/api/v1/branches` → succeeds (kasir shouldn't be able to create branches)
  3. DELETE `/api/v1/products/{id}` → succeeds (kasir shouldn't be able to delete products)
- **Status**: ⚠️ **Partial** — Role exists in JWT (`role` claim) and User model, but no authorization middleware enforces it

---

## 📊 Summary

| Category | Count | Status |
|----------|-------|--------|
| Original Sprint 2 bugs — **Fixed** | 6 | ✅ All resolved |
| New bugs found (current code) — **Critical** | 2 (A, B) | ❌ Still broken |
| New bugs found — **Major** | 2 (C, D) | ❌ Still broken |
| New bugs found — **Minor** | 2 (E, F) | ⚠️ Partial |

### Recommendation
1. Add explicit barcode uniqueness check in `CreateProduct` handler → return 409 Conflict
2. Add `CategoryID` existence validation before product INSERT
3. Add `cost_price` column to Product model/migration and use it in P&L query
4. Wrap entire checkout flow in a DB transaction with proper rollback
5. Add referential checks before soft-deleting branches/products
6. Implement RBAC middleware (route-based role restrictions)

---

*Generated by QA Retest Sprint 2 (BE #2)*
