# Task: Backend — POS Multi Branch Retail (Go + Supabase)

**Agent:** @tukangbackendbot

## Stack
- Go (REST API server)
- Supabase (PostgreSQL, Auth, Realtime)

## PRD
Baca dulu: `PRD.md` di folder ini.

## Yang Harus Dikerjain

### Sprint 1 — Foundation
1. **Setup Go project structure**
   - Clean architecture: `handlers/`, `services/`, `repositories/`, `models/`, `middleware/`, `migrations/`
   - Router (chi, gin, atau echo)
   - Environment config (.env)
   - CORS middleware

2. **Setup Supabase**
   - Project baru
   - Enable Auth (email/password)
   - Create migration files

3. **Database Schema & Migrations**

   ```sql
   -- users (managed by Supabase Auth + public.users table)
   CREATE TABLE public.users (
     id UUID PRIMARY KEY REFERENCES auth.users(id),
     email TEXT UNIQUE NOT NULL,
     name TEXT NOT NULL,
     role TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'kasir')),
     branch_id UUID REFERENCES public.branches(id) ON DELETE SET NULL,
     is_active BOOLEAN DEFAULT true,
     created_at TIMESTAMPTZ DEFAULT NOW()
   );

   -- branches
   CREATE TABLE public.branches (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     name TEXT NOT NULL,
     code TEXT UNIQUE NOT NULL,
     address TEXT,
     phone TEXT,
     is_active BOOLEAN DEFAULT true,
     created_at TIMESTAMPTZ DEFAULT NOW()
   );

   -- categories
   CREATE TABLE public.categories (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     name TEXT NOT NULL,
     created_at TIMESTAMPTZ DEFAULT NOW()
   );

   -- products
   CREATE TABLE public.products (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     category_id UUID REFERENCES public.categories(id) ON DELETE SET NULL,
     name TEXT NOT NULL,
     barcode TEXT UNIQUE,
     unit TEXT DEFAULT 'pcs',
     default_price DECIMAL(15,2) DEFAULT 0,
     image_url TEXT,
     is_active BOOLEAN DEFAULT true,
     created_at TIMESTAMPTZ DEFAULT NOW()
   );

   -- branch_products (pivot + inventory)
   CREATE TABLE public.branch_products (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     branch_id UUID REFERENCES public.branches(id) ON DELETE CASCADE,
     product_id UUID REFERENCES public.products(id) ON DELETE CASCADE,
     price DECIMAL(15,2) NOT NULL,
     stock_qty DECIMAL(15,2) DEFAULT 0,
     min_stock DECIMAL(15,2) DEFAULT 0,
     created_at TIMESTAMPTZ DEFAULT NOW(),
     UNIQUE(branch_id, product_id)
   );

   -- transactions
   CREATE TABLE public.transactions (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     branch_id UUID REFERENCES public.branches(id) ON DELETE CASCADE,
     cashier_id UUID REFERENCES public.users(id) ON DELETE SET NULL,
     total_amount DECIMAL(15,2) NOT NULL,
     discount DECIMAL(15,2) DEFAULT 0,
     payment_method TEXT NOT NULL,
     payment_amount DECIMAL(15,2) NOT NULL,
     change_amount DECIMAL(15,2) DEFAULT 0,
     status TEXT DEFAULT 'completed' CHECK (status IN ('pending', 'completed', 'cancelled')),
     created_at TIMESTAMPTZ DEFAULT NOW()
   );

   -- transaction_items
   CREATE TABLE public.transaction_items (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     transaction_id UUID REFERENCES public.transactions(id) ON DELETE CASCADE,
     product_id UUID REFERENCES public.products(id) ON DELETE SET NULL,
     qty DECIMAL(15,2) NOT NULL,
     price DECIMAL(15,2) NOT NULL,
     subtotal DECIMAL(15,2) NOT NULL
   );

   -- stock_mutations
   CREATE TABLE public.stock_mutations (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     branch_id UUID REFERENCES public.branches(id) ON DELETE CASCADE,
     product_id UUID REFERENCES public.products(id) ON DELETE SET NULL,
     type TEXT NOT NULL CHECK (type IN ('in', 'out', 'transfer_in', 'transfer_out')),
     qty DECIMAL(15,2) NOT NULL,
     reference_id UUID,
     notes TEXT,
     created_by UUID REFERENCES public.users(id),
     created_at TIMESTAMPTZ DEFAULT NOW()
   );
   ```

4. **RLS Policies**
   - Branches: owner all access, admin/kasir only assigned branch
   - Products: owner all access, admin/kasir per cabang
   - Transactions: by branch access

5. **Auth Endpoints**
   - `POST /api/v1/auth/login` — login via Supabase Auth
   - `POST /api/v1/auth/logout`
   - `GET /api/v1/auth/me` — get profile + branch info
   - JWT middleware

6. **Branch CRUD**
   - `GET /api/v1/branches` — list
   - `POST /api/v1/branches` — create
   - `PUT /api/v1/branches/:id` — update
   - `DELETE /api/v1/branches/:id` — soft delete (set is_active=false)

7. **Category CRUD**
   - `GET /api/v1/categories`
   - `POST /api/v1/categories`
   - `PUT /api/v1/categories/:id`
   - `DELETE /api/v1/categories/:id`

8. **Product + BranchProduct Endpoints**
   - `GET /api/v1/branches/:id/products` — list produk di cabang + stock
   - `POST /api/v1/branches/:id/products` — assign product to branch (dengan price override & initial stock)
   - `PUT /api/v1/branches/:id/products/:productId` — update price/stock/min_stock
   - `DELETE /api/v1/branches/:id/products/:productId` — unassign
   - `GET /api/v1/products` — master products (global)
   - `POST /api/v1/products` — create master product
   - `PUT /api/v1/products/:id`
   - `GET /api/v1/products/search?q=&branch_id=` — search by name/barcode

#### ⚡ Local-First Consideration
Backend harus siap nerima **batch sync requests** dari device offline. Setiap endpoint harus idempotent (pake `idempotency_key` biar transaksi ga dobel).

## Sprint 2 — POS Transactions
1. **Create Transaction**
   ```
   POST /api/v1/branches/:id/transactions
   Body: {
     items: [{ product_id, qty, price }],
     discount: 0,
     payment_method: "cash",
     payment_amount: 50000
   }
   ```
   - Dalam satu endpoint: create transaction + items + kurangi stock di branch_products
   - Pakai database transaction

2. **List Transactions**
   `GET /api/v1/branches/:id/transactions?page=1&limit=20&start=&end=`

3. **Transaction Detail**
   `GET /api/v1/branches/:id/transactions/:transactionId`

4. **Realtime**
   - Subscribe ke tabel transactions via Supabase Realtime
   - Notify cashier screen ada transaksi baru

### Sprint 2.5 — Sync Endpoints (Local-First) 🏠
1. **Batch Sync Entry**
   ```
   POST /api/v1/sync/push
   Body: {
     device_id: "uuid",
     transactions: [{ ... }],
     stock_mutations: [{ ... }],
     last_synced_at: "ISO8601"
   }
   Response: {
     synced: [{ id: "...", status: "success" }, ...],
     conflicts: [{ local_id: "...", server_data: {...}, reason: "..." }],
     server_time: "ISO8601"
   }
   ```
   - Idempotency via `idempotency_key` di setiap item
   - Server validasi: cek duplikat, cek stok cukup
   - Kembalikan conflict kalo stok di server beda

2. **Full Pull (Master Data)**
   ```
   GET /api/v1/sync/pull?branch_id=&since=ISO8601
   Response: {
     products: [...],
     branch_products: [...],
     branches: [...],
     categories: [...],
     server_time: "ISO8601"
   }
   ```
   - Hanya kirim data yang updated `>= since`
   - Seluruh master data per cabang

3. **Conflict Resolution**
   ```
   POST /api/v1/sync/resolve
   Body: {
     conflicts: [
       {
         local_id: "...",
         table: "transactions",
         resolution: "local" | "server",
         merged_data: {...} // optional kalo manual
       }
     ]
   }
   ```

### Sprint 3 — Inventory & Reports
1. **Stock Adjustment**
   `POST /api/v1/branches/:id/inventory/adjustment`
   Body: { product_id, type: "in"|"out", qty, notes }
   - Insert stock_mutations + update branch_products stock_qty

2. **Transfer Stock**
   `POST /api/v1/inventory/transfer`
   Body: { source_branch_id, target_branch_id, product_id, qty, notes }
   - Insert stock_mutations (transfer_out + transfer_in)
   - Update source: stock_qty - qty
   - Update target: stock_qty + qty

3. **Stock History**
   `GET /api/v1/branches/:id/inventory/history?product_id=&type=&start=&end=`

4. **Reports**
   - `GET /api/v1/branches/:id/reports/sales?start=&end=`
   - `GET /api/v1/branches/:id/reports/stock`
   - `GET /api/v1/branches/:id/reports/profit-loss?start=&end=`

## API Contract
- Base: `/api/v1`
- Response success: `{"status": "success", "data": {...}}`
- Response list: `{"status": "success", "data": [...], "pagination": { "page": 1, "limit": 20, "total": 100, "totalPages": 5 }}`
- Response error: `{"status": "error", "message": "..."}`
- IDs: UUID v4
- Timestamps: ISO 8601
- Auth: `Authorization: Bearer <supabase-jwt>`
- CORS: allow all origins (development)

## Output
- Full Go source code di `pos-multi-branch/backend/`
- Migration SQL files di `pos-multi-branch/backend/migrations/`
- README dengan cara setup & run
- .env.example

## Koordinasi
- Kalo ada perubahan API contract, kabari gua dulu biar gua koordinasi ke frontend
- Priority: Sprint 1 dulu full, baru lanjut
- Jangan merge ke main kalo belum di-review

Gas 🔥
