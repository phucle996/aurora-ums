BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;
-- Schema is resolved from current search_path (provided by migrator).

-- =========================
-- USER
-- =========================

CREATE TYPE user_status AS ENUM (
  'active',
  'pending',
  'suspended',
  'deleted'
);

-- =========================
-- MFA
-- =========================

CREATE TYPE mfa_method_type AS ENUM (
  'totp',
  'sms',
  'email'
);

-- =========================
-- RBAC
-- =========================

CREATE TYPE role_scope AS ENUM (
  'global',
  'tenant'
);

COMMIT;
