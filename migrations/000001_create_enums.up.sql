BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE SCHEMA IF NOT EXISTS ums;

-- =========================
-- USER
-- =========================

CREATE TYPE ums.user_status AS ENUM (
  'active',
  'pending',
  'suspended',
  'deleted'
);

-- =========================
-- MFA
-- =========================

CREATE TYPE ums.mfa_method_type AS ENUM (
  'totp',
  'sms',
  'email'
);

-- =========================
-- ONE TIME TOKEN
-- =========================

CREATE TYPE ums.one_time_token_purpose AS ENUM (
  'account_verify',
  'password_reset'
);

-- =========================
-- RBAC
-- =========================

CREATE TYPE ums.role_scope AS ENUM (
  'global',
  'tenant'
);

COMMIT;
