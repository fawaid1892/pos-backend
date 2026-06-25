# POS Multi Branch — Backend API

## Tech Stack

- Go 1.22+
- PostgreSQL (via pgx v5)
- JWT (golang-jwt v5)
- bcrypt (golang.org/x/crypto)
- No external framework — standard library `net/http` (Go 1.22 routing)

## Setup

1. Copy `.env.example` to `.env` and adjust values.
2. Create PostgreSQL database:
   ```sql
   CREATE DATABASE pos_multi_branch;
   ```
3. Run migration:
   ```bash
   psql -d pos_multi_branch -f migrations/001_initial.sql
   ```
4. Run server:
   ```bash
   go run cmd/server/main.go
   ```

## API Endpoints (v1)

### Public
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/login` | Login with username/password |

### Protected (Bearer token)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/auth/me` | Current user info |
| POST | `/api/v1/auth/logout` | Logout |
| GET | `/api/v1/branches` | List branches |
| GET | `/api/v1/branches/{id}` | Get branch by ID |
| POST | `/api/v1/branches` | Create branch |
| PUT | `/api/v1/branches/{id}` | Update branch |
| DELETE | `/api/v1/branches/{id}` | Soft delete branch |
| GET | `/api/v1/products` | List/search products (?q=&barcode=) |
| GET | `/api/v1/products/{id}` | Get product by ID |
| POST | `/api/v1/products` | Create product |
| PUT | `/api/v1/products/{id}` | Update product |
| DELETE | `/api/v1/products/{id}` | Soft delete product |
| GET | `/api/v1/categories` | List categories |
| POST | `/api/v1/categories` | Create category |
| POST | `/api/v1/transactions/checkout` | Create transaction |
| GET | `/api/v1/transactions` | List transactions (?branch_id=) |
| GET | `/api/v1/transactions/{id}` | Get transaction detail |

## Seed Credentials

| Username | Password | Role |
|----------|----------|------|
| admin | password123 | admin |
| kasir1 | password123 | kasir |
| owner | password123 | owner |
