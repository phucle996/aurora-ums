# Database Migrations Guide

This project uses plain SQL migrations stored in the `migrations/` directory.

## Structure
- Each migration has matching `*.up.sql` and `*.down.sql` files.
- Files are versioned and zero‑padded: `00000X_<description>.up.sql` / `.down.sql`.
- Existing migrations must never be edited once applied to any environment. Add a new version for changes.

## Order of Operations
1) Enums / types  
2) Tables  
3) Constraints (FK/UNIQUE/CHECK)  
4) Indexes  
5) Seed data (system-level only)

## Conventions
- Database: PostgreSQL
- Primary keys: UUID (`gen_random_uuid()` / `uuid_generate_v4()`)
- Timestamps: `TIMESTAMPTZ`
- JSON metadata: `JSONB`
- Constraint names:  
  - Primary key: `pk_<table>`  
  - Foreign key: `fk_<table>_<column>`  
  - Unique: `uq_<table>_<column>`  
- Index names: `idx_<table>_<column>` (or `_col1_col2` for composites)
- Default to `ON DELETE RESTRICT`; only use CASCADE when justified.

## Running Migrations
Use your preferred migration runner to apply SQL files in order. Ensure `pgcrypto` (for `gen_random_uuid`) is available.

## Creating a New Migration
1) Pick next sequential version number.
2) Create both `.up.sql` and `.down.sql`.
3) Follow the phase order above; avoid mixing phases in one file.
4) Keep changes minimal and reversible in the down migration.

## Safety
- Never drop production data without explicit approval.
- Use `IF NOT EXISTS` / `IF EXISTS` where safe to keep migrations idempotent.
- Foreign key columns must be indexed.

## Current Scope
- Core identity tables, MFA, RBAC, profiles are already defined up to version `000008`.
- Seed data initializes root user/role/permissions.

