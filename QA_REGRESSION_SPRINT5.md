# QA Regression Report — Sprint 5 (BE #12)

**Date:** 2026-06-28  
**Tester:** Tukang QA Bot  
**Branch Backend:** `sprint-7` (based on `main`)  
**Branch Frontend:** `sprint-7`  

---

## 1. REGRESSION — WebSocket (Backend)

| Item | Status | Detail |
|------|--------|--------|
| WebSocket hub file exists (`internal/ws/hub.go`) | ✅ | 164 lines, proper `Hub` struct with `sync.RWMutex` |
| WebSocket handler file exists (`internal/ws/handler.go`) | ✅ | 200 lines, `ServeWS` with upgrade + read/write pumps |
| Broadcast notifikasi transaksi baru (`transaction.created`) | ⚠️ **Partial** | Event constant `EventTransactionCreated` defined; `BroadcastEvent`/`BroadcastEventToBranch` methods available. **BUT** no handler actually emits `transaction.created` events. Checkout handler only broadcasts low-stock. |
| Broadcast update stok (`stock.adjusted`, `stock.transferred`) | ⚠️ **Partial** | Event constants defined. **BUT** `stock.adjusted`/`stock.transferred` are never emitted by stock handlers. Only `stock.low` is broadcast. |
| Broadcast `stock.low` | ✅ | `checkAndBroadcastLowStock()` called in `Checkout()` (line 186) and `Adjustment()` (line 92) |
| Race condition check — Hub | ✅ **No race** | `Hub` uses `sync.RWMutex` — write lock on `Register`/`Unregister`, read lock on `broadcastBytes`/`ClientCount`/`RoomClientCount` |
| Race condition check — conn registry (`conns` map) | ✅ **No race** | Global `conns` map protected by `connMu sync.Mutex`; read/write/delete all serialize through the mutex |
| Race condition check — double conn.Close() | ⚠️ **Minor** | Both `readPump()` and `writePump()` defer blocks call `conn.Close()`. Currently safe because mutex ensures the second sees `nil`, but a race window exists if both exit simultaneously. Low severity. |

**Verdict: ⚠️ PASS (hub infrastructure OK, but `transaction.created`, `stock.adjusted`, `stock.transferred` event types defined yet never emitted by handlers)**

---

## 2. REGRESSION — Bug Fixes (Backend)

| Bug Fix | Sprint 5 Commit | Status | Detail |
|---------|----------------|--------|--------|
| SQL injection prevention — whitelist table/column validation | `3477991` | ✅ **Intact** | `validPullTables`, `validResolveTables`, `validResolveColumns` maps in `internal/handler/sync.go` (lines 20-47) |
| writeJSON error handling fix | `d33ee48` / `05db9d1` | ✅ **Intact** | `writeJSON()` function in `internal/handler/transaction.go` (lines 256-264) handles `json.Encode` error gracefully without panic |
| Remove mock cost calc in GetProfitLossReport | `d33ee48` / `05db9d1` | ✅ **Intact** | `GetProfitLossReport()` in `internal/repository/repository.go` (lines 663-696) uses actual `cost_price` from products table, fallback `p.price * 0.7` when cost_price is 0/NULL |
| Fix: hardcoded COGS multiplier → actual cost_price | `bf24a8b` | ✅ **Intact** | Already checked — same fix verified above |

**Git log — Sprint 5 bug-fix commits:**
```
d33ee48 feat(backend): WebSocket realtime + Docker + bug fixes
05db9d1 feat(backend): WebSocket realtime + Docker + bug fixes
3477991 fix(backend): SQL injection prevention - whitelist table/column validation
cc7bd87 fix(backend): SQL injection prevention - whitelist table/column validation
bf24a8b fix(pl): replace hardcoded cost multiplier with actual cost_price from products
```

**Verdict: ✅ PASS**

---

## 3. REGRESSION — Docker

| Item | Expected | Status | Detail |
|------|----------|--------|--------|
| Dockerfile exists | yes | ✅ | Multi-stage build (`golang:1.22-alpine` → `alpine:3.19`) |
| Dockerfile exposes port | 8080 | ✅ | `EXPOSE 8080` |
| Multi-stage build | yes | ✅ | Builder stage compiles, run stage copies binary |
| docker-compose.yml exists | yes | ✅ | Single service `backend` |
| Port mapping | 8080:8080 | ✅ | Correct |
| Environment variables | SUPABASE_URL, SUPABASE_ANON_KEY, JWT_SECRET, PORT | ✅ | All 4 env vars defined (read from host environment) |
| No deprecated `version` field | yes | ✅ | Fixed in commit `4eaa392` |
| Volumes declared but not mounted | — | ⚠️ **Minor** | Volume `backend_data` declared but not mounted to any service. No functional impact. |
| DB service in compose | optional | ℹ️ | No PostgreSQL service — assumes external/Supabase. Fine for cloud setup. |
| `.env.example` copied | yes | ✅ | `COPY .env.example .env` in Dockerfile |

**Verdict: ✅ PASS (2 minor notes)**

---

## 4. REGRESSION — UI Polish (Frontend)

| Item | Expected | Status | Detail |
|------|----------|--------|--------|
| `empty_state_widget.dart` in `lib/widgets/` | ✅ exists | ❌ **MISSING on sprint-7** | File exists on `main` but NOT in `sprint-7` branch |
| `error_state_widget.dart` in `lib/widgets/` | ✅ exists | ❌ **MISSING on sprint-7** | File exists on `main` but NOT in `sprint-7` branch |
| `shimmer_loading.dart` in `lib/widgets/` | ✅ exists | ✅ **Exists** | Full shimmer implementation (ShimmerLoading, ShimmerBox, ShimmerListTile, ShimmerPage, ShimmerCard) |
| Dark mode toggle via `theme_provider.dart` | ✅ working | ❌ **MISSING on sprint-7** | `lib/providers/theme_provider.dart` does not exist on sprint-7. Only `cart_provider.dart` present. |
| Dark mode toggle usage in screens | ✅ working | ❌ **MISSING** | No imports of `ThemeProvider` or `toggleTheme` found in any file on sprint-7 |

**Root cause:** Commit `bf21a23` (`feat(frontend): UI polish + dark mode + responsive layout`) was applied on `main` but `sprint-7` branched off from `89478cf` (Sprint 6) before that commit. The UI polish features from Sprint 5 were never merged into `sprint-7`.

**Verdict: ❌ FAIL — 3 critical components missing from sprint-7 branch**

---

## 5. REGRESSION — Responsive Layout (Frontend)

| Item | Expected | Status | Detail |
|------|----------|--------|--------|
| `lib/utils/responsive.dart` exists | ✅ exists | ❌ **MISSING on sprint-7** | `lib/utils/` directory does not exist on sprint-7 |
| Responsive breakpoints (mobile/tablet/desktop) | ✅ defined | ❌ **MISSING** | `Breakpoints` class with `mobile=600`, `tablet=900`, `desktop=1200` exists on `main` but not on sprint-7 |
| `ResponsiveLayout` widget | ✅ exists | ❌ **MISSING** | Same file missing |
| `ResponsiveValue<T>` helper | ✅ exists | ❌ **MISSING** | Same file missing |
| `ResponsivePadding` helper | ✅ exists | ❌ **MISSING** | Same file missing |

**Root cause:** Same as #4 — commit `bf21a23` not merged into `sprint-7`.

**Verdict: ❌ FAIL — responsive layout completely missing from sprint-7 branch**

---

## 6. Overall Summary

| # | Area | Verdict | Open Issues |
|---|------|---------|-------------|
| 1 | WebSocket (Backend) | ⚠️ PASS — Hub capable | Event types `transaction.created`, `stock.adjusted`, `stock.transferred` defined but never emitted by any handler. Only `stock.low` is broadcast. Minor: double `conn.Close()` race edge case. |
| 2 | Bug Fixes (Backend) | ✅ PASS | All Sprint 5 bug fixes intact |
| 3 | Docker | ✅ PASS | Minor: unused volume `backend_data` |
| 4 | UI Polish (Frontend) | ❌ FAIL | **3 missing components**: `empty_state_widget`, `error_state_widget`, `theme_provider` (dark mode) |
| 5 | Responsive Layout (Frontend) | ❌ FAIL | **Entire `lib/utils/responsive.dart` missing** — no breakpoints, no responsive helpers |
| **Overall** | | **FAIL (2/5)** | Frontend polish features not merged into sprint-7 branch; WebSocket event wiring incomplete |

---

## Open Bugs

1. **UI Polish files not merged into `sprint-7`** (Critical)
   - `lib/widgets/empty_state_widget.dart` — missing on sprint-7
   - `lib/widgets/error_state_widget.dart` — missing on sprint-7
   - `lib/providers/theme_provider.dart` — missing on sprint-7 (dark mode toggle broken)
   - All dark mode UI integrations absent
   - **Fix:** Merge commit `bf21a23` (`feat(frontend): UI polish + dark mode + responsive layout`) from `main` into `sprint-7`

2. **Responsive layout completely absent from `sprint-7`** (Critical)
   - `lib/utils/responsive.dart` — missing on sprint-7
   - No responsive breakpoints, no `ResponsiveLayout` widget, no `ResponsiveValue` helper
   - **Fix:** Same as above — merge `bf21a23` from `main` into `sprint-7`

3. **WebSocket: event types defined but never emitted** (Medium)
   - `EventTransactionCreated` (`transaction.created`) defined but no handler emits it after checkout
   - `EventStockAdjusted` (`stock.adjusted`) defined but stock Adjustment handler doesn't emit it
   - Only `EventStockLow` (`stock.low`) is actually broadcast
   - **Recommendation:** Add broadcast calls in `Checkout()` (for transaction.created) and `Adjustment()`/`Transfer()` (for stock.adjusted/stock.transferred)

4. **Minor: WebSocket double `conn.Close()` edge case** (Low)
   - Both `readPump()` and `writePump()` deferred cleanup may call `conn.Close()` simultaneously
   - Currently mitigated by mutex check, but not an ideal pattern
   - **Recommendation:** Check if `conn.Close()` is called before or after mutex-protected delete

---

## Recommendations

1. **Merge `main` → `sprint-7` for frontend** — The Sprint 5 UI Polish + Dark Mode + Responsive Layout features (`bf21a23`) need to be merged into sprint-7 to complete the regression suite.
2. **Add unit tests for WebSocket hub** — No tests exist for `internal/ws/hub.go` or `handler.go`. Add concurrent client register/unregister/broadcast tests to verify race condition safety.
3. **Mount `backend_data` volume in docker-compose** — Declared but unused; mount to `/app/data` or wherever SQLite/state persists.
4. **Add DB service to docker-compose** — For local dev without Supabase, add a PostgreSQL service.
5. **Track frontend branch divergence** — Ensure future sprint branches inherit all previous sprint features before adding new ones.
