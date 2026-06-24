# Sprint 1 Plan — POS Multi Branch

> **Periode:** [Start Date] — [End Date]
> **Goal:** Foundation online-first — Go project setup, Auth, Master Data (Branch, Product, Category)

---

## Sprint Backlog

| # | Task | Assignee | Estimasi | Status |
|---|------|----------|----------|--------|
| 1 | Setup Go project structure + Supabase client | @tukangbackendbot | - | ⚪ Not Started |
| 2 | Auth API (login/logout/me) + Supabase Auth | @tukangbackendbot | - | ⚪ Not Started |
| 3 | CRUD Branch API | @tukangbackendbot | - | ⚪ Not Started |
| 4 | CRUD Product & Category API | @tukangbackendbot | - | ⚪ Not Started |
| 5 | Flutter login screen + auth flow | @tukangfrontendbot | - | ⚪ Not Started |
| 6 | Flutter master data management (Branch, Product, Category) | @tukangfrontendbot | - | ⚪ Not Started |
| 7 | Test case untuk auth flow & master data | @tukangqabot | - | ⚪ Not Started |

## Breakdown per Task

### Task 1: Setup Go Project + Supabase
**Assignee:** @tukangbackendbot
**Acceptance Criteria:**
- [ ] Go module initialized dengan proper structure
- [ ] Supabase client (Go) terkonfigurasi
- [ ] Database migration untuk semua tabel awal (users, branches, categories, products, branch_products)
- [ ] README cara run project
- **Referensi:** `pos-multi-branch/PRD.md` — Section 4 & 7 (schema)

### Task 2: Auth API
**Assignee:** @tukangbackendbot
**Acceptance Criteria:**
- [ ] `POST /api/v1/auth/login` — login dengan email & password, return JWT/session
- [ ] `POST /api/v1/auth/logout` — invalidate session
- [ ] `GET /api/v1/auth/me` — return current user info + branch
- [ ] Middleware auth untuk endpoint lain
- **Referensi:** PRD Section 3.1, 6, 7

### Task 3: CRUD Branch API
**Assignee:** @tukangbackendbot
**Acceptance Criteria:**
- [ ] `GET /api/v1/branches` — list branch (owner all-access)
- [ ] `POST /api/v1/branches` — create branch
- [ ] `GET /api/v1/branches/:id` — get branch detail
- [ ] `PUT /api/v1/branches/:id` — update branch
- [ ] `DELETE /api/v1/branches/:id` — soft delete branch
- **Referensi:** PRD Section 6

### Task 4: CRUD Product & Category API
**Assignee:** @tukangbackendbot
**Acceptance Criteria:**
- [ ] `GET /api/v1/categories` — list categories
- [ ] `POST /api/v1/categories` — create category
- [ ] `GET /api/v1/branches/:id/products` — list products per branch
- [ ] `POST /api/v1/branches/:id/products` — create product per branch
- [ ] `PUT /api/v1/branches/:id/products/:productId` — update product
- [ ] `DELETE /api/v1/branches/:id/products/:productId` — soft delete
- [ ] `GET /api/v1/products/search?q=&branch_id=` — search produk
- **Referensi:** PRD Section 3.2, 6

### Task 5: Flutter Login Screen
**Assignee:** @tukangfrontendbot
**Acceptance Criteria:**
- [ ] Login screen dengan form email + password
- [ ] Hit endpoint login backend
- [ ] Simpan token/session
- [ ] Navigasi ke dashboard setelah login
- [ ] Tampilkan role & cabang user
- [ ] Logout
- **Referensi:** PRD Section 3.1, hubungi @tukangbackendbot untuk API contract

### Task 6: Flutter Master Data Management
**Assignee:** @tukangfrontendbot
**Acceptance Criteria:**
- [ ] Halaman list Branch (read-only dulu)
- [ ] Halaman list Category + CRUD (add/edit/delete)
- [ ] Halaman list Product per Branch + CRUD
- [ ] Form validasi
- [ ] Integrasi dengan API backend
- **Referensi:** PRD Section 3.2, hubungi @tukangbackendbot untuk API contract

### Task 7: QA Test Cases
**Assignee:** @tukangqabot
**Acceptance Criteria:**
- [ ] Test case untuk auth flow (login sukses, gagal, logout, session expired)
- [ ] Test case untuk CRUD Branch
- [ ] Test case untuk CRUD Product & Category
- [ ] Test case untuk Flutter screens
- **Referensi:** PRD Section 6, 7

---

## Risks / Blocker
- [ ] API contract antara frontend & backend harus disepakati dulu