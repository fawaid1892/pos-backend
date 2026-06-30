-- Migration: UUID → int64 (16-digit crypto-random)
-- Run this ONCE if Railway auto-migrate fails with "cannot cast type uuid to bigint"
-- WARNING: Drops all data and re-seeds!

DROP TABLE IF EXISTS
    role_permissions,
    promotion_branches,
    promotions,
    stock_mutations,
    transaction_items,
    transactions,
    refresh_tokens,
    users,
    branch_products,
    products,
    categories,
    branches,
    roles,
    permissions
CASCADE;
