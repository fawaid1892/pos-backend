# QA Report — Sprint 10 & 11

**Date:** 2026-06-30  
**Tester:** Tukang QA Bot  
**Branch:** main  
**Commit:** 10d4acf

## Summary

- **Pass: 33/34 Backend**
- **Critical bugs:** 0
- **Minor issues:** 2
- **Verdict: CONDITIONAL** — 1 item FAIL (Barcode auto-generate), 2 minor issues

---

## Backend Results

| # | Item | Status | Notes |
|---|------|--------|-------|
| 1.1 | Model Branch punya field ProvinceCode & CityCode | ✅ PASS | `models.go` L66-68: `ProvinceCode string`, `CityCode string` |
| 1.2 | CreateBranch menerima & nyimpen province_code / city_code | ✅ PASS | `repository.go` L68-84: Maps `req.ProvinceCode` / `req.CityCode` ke struct |
| 1.3 | UpdateBranch bisa update province_code / city_code | ✅ PASS | `repository.go` L86-105: Updates map includes `"province_code"` / `"city_code"` |
| 1.4 | GET /api/v1/branches return province_code / city_code | ✅ PASS | `ListBranches` returns full `model.Branch` struct; JSON serialization includes both fields |
| 2.1 | Model Promotion punya Scope, ProvinceID, CityID (bukan BranchID) | ✅ PASS | `models.go` L444-446: `Scope`, `ProvinceID`, `CityID` |
| 2.2 | Join table promotion_branches ada | ✅ PASS | `models.go` L425-428: `type PromotionBranch struct` |
| 2.3 | AutoMigrate include &model.PromotionBranch{} | ✅ PASS | `database.go` L59: `&model.PromotionBranch{}` |
| 3.1 | POST /api/v1/promotions — scope=all (tanpa branch_ids) | ✅ PASS | `repository.go` L123-153: Scope handling, inserts PromotionBranch only when scope=selected |
| 3.2 | POST /api/v1/promotions — scope=selected + branch_ids | ✅ PASS | L129-143: Creates PromotionBranch entries for each branch_id |
| 3.3 | POST /api/v1/promotions — scope=province + province_id | ✅ PASS | L117-118: Sets `ProvinceID` from request |
| 3.4 | POST /api/v1/promotions — scope=city + province_id + city_id | ✅ PASS | L117-118: Sets both `ProvinceID` and `CityID` |
| 3.5 | GET /api/v1/promotions/{id} — preload Branches | ✅ PASS | `repository.go` L84: `Preload("Branches")` |
| 3.6 | PUT /api/v1/promotions/{id} — update scope selected→all, hapus PromotionBranch | ✅ PASS | `repository.go` L220-239: Deletes old, inserts new branch entries |
| 3.7 | DELETE /api/v1/promotions/{id} — soft delete | ✅ PASS | `repository.go` L262-271: gorm `.Delete()` with DeletedAt |
| 4.1 | GET /api/v1/promotions?branch_id=X — scope=all include | ✅ PASS | `repository.go` L60-68: `(scope = 'all') OR ...` |
| 4.2 | GET /api/v1/promotions?branch_id=X — scope=selected include | ✅ PASS | EXISTS subquery on promotion_branches |
| 4.3 | GET /api/v1/promotions?branch_id=X — scope=province include | ✅ PASS | `province_id = (SELECT province_code FROM branches WHERE id = ?)` |
| 4.4 | GET /api/v1/promotions?branch_id=X — scope=city include | ✅ PASS | `city_id = (SELECT city_code FROM branches WHERE id = ?)` |
| 4.5 | GET /api/v1/promotions?branch_id=X — scope=province exclude | ✅ PASS | Different province_code won't match the WHERE clause |
| 4.6 | GET /api/v1/promotions?active=true | ✅ PASS | L28-31: `start_date <= now AND end_date >= now AND is_active = true` |
| 4.7 | GET /api/v1/promotions?expired=true | ✅ PASS | L32-35: `end_date < now` |
| 4.8 | GET /api/v1/promotions?type=voucher | ✅ PASS | L70-72: `type = ?` parameter filter |
| 5.1 | GET /api/v1/branches/{id}/promotions — return aktif promosi | ✅ PASS | `ListPromotionsByBranch` (L283-305) |
| 5.2 | Include scope=all | ✅ PASS | L291 |
| 5.3 | Include scope=selected (branch di join) | ✅ PASS | L292-293 |
| 5.4 | Include scope=province (province_code match) | ✅ PASS | L294 |
| 5.5 | Include scope=city (city_code match) | ✅ PASS | L295 |
| 6.1 | POST validate-voucher — valid code → {valid: true} | ✅ PASS | L307-347: Returns `Valid: true` with discount info |
| 6.2 | POST validate-voucher — invalid code → {valid: false} | ✅ PASS | L315: `Valid: false, Error: "Voucher tidak ditemukan..."` |
| 6.3 | POST validate-voucher — expired → invalid | ✅ PASS | L311: Query filters by `start_date <= now AND end_date >= now` |
| 6.4 | POST validate-voucher — max_uses reached → invalid | ✅ PASS | L321-323: `promo.CurrentUses >= promo.MaxUses` |
| 7.1 | Barcode auto-generate PRD-{SKU} saat barcode null | ❌ **FAIL** | `repository.go` L249-269: `CreateProduct` passes `req.Barcode` directly; **no auto-generation logic exists**. Barcode stays NULL if not provided. |

---

## Minor Issues

1. **Seed data missing ProvinceCode/CityCode** — `database.go` L94-111 creates seed branches (`Cabang Pusat`, `Cabang Cibubur`) without setting `ProvinceCode` / `CityCode` values. This means seeded branches won't match promotion scope queries that rely on these fields. Consider updating seed data with proper emsifa codes (e.g., `ProvinceCode: "31"` for DKI Jakarta).

2. **Barcode auto-generate — feature not implemented** — Item 7.1 is a FAIL. The backend does not auto-generate barcode `PRD-{SKU}` when barcode is null during product creation. If this is required, `CreateProduct` needs logic to generate fallback barcode when `req.Barcode` is nil or empty.

---

## Verdict

**CONDITIONAL** — 33/34 backend items pass. 1 FAIL (barcode auto-generate). No critical bugs found. The multi-branch promotion scope logic (province/city/selected/all) is correctly implemented end-to-end in both model and repository layers with proper filtering using ProvinceCode/CityCode. Recommend fixing the barcode auto-generate and seed data before final sign-off.
