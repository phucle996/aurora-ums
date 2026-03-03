BEGIN;

-- =========================
-- CORE IDENTITY
-- =========================

CREATE TABLE ums.users (
  id              UUID PRIMARY KEY,
  username        VARCHAR(100) UNIQUE NOT NULL,
  email           VARCHAR(100) UNIQUE NOT NULL,
  password_hash   TEXT,
  status          ums.user_status,
  user_level      INTEGER,
  on_boarding     BOOLEAN,
  created_at      TIMESTAMPTZ,
  updated_at      TIMESTAMPTZ
);

CREATE TABLE ums.profiles (
  id              UUID PRIMARY KEY,
  user_id         UUID NOT NULL,
  full_name       TEXT,
  company         TEXT,
  referral_source TEXT,
  phone           VARCHAR(30),
  job_function    TEXT,
  country         TEXT,
  avatar_url      TEXT,
  bio             TEXT,
  created_at      TIMESTAMPTZ,
  updated_at      TIMESTAMPTZ
);

-- =========================
-- AUTHENTICATION
-- =========================

CREATE TABLE ums.refresh_tokens (
  id              UUID PRIMARY KEY,
  user_id         UUID NOT NULL,
  device_id       UUID,
  token_hash      TEXT UNIQUE NOT NULL,
  expires_at      TIMESTAMPTZ,
  created_at      TIMESTAMPTZ
);

CREATE TABLE ums.user_devices (
  id                  UUID PRIMARY KEY,
  user_id             UUID NOT NULL,
  device_id           UUID NOT NULL,
  device_secret_hash  TEXT NOT NULL,
  user_agent          TEXT,
  ip_first            TEXT,
  ip_last             TEXT,
  revoked             BOOLEAN DEFAULT FALSE,
  created_at          TIMESTAMPTZ,
  last_seen           TIMESTAMPTZ,
  revoked_at          TIMESTAMPTZ,
  UNIQUE (device_id)
);

CREATE TABLE ums.one_time_tokens (
  id              UUID PRIMARY KEY,
  user_id         UUID NOT NULL,
  token_hash      TEXT UNIQUE NOT NULL,
  purpose         ums.one_time_token_purpose,
  expires_at      TIMESTAMPTZ,
  created_at      TIMESTAMPTZ
);

-- =========================
-- RBAC
-- =========================

CREATE TABLE ums.roles (
  id            UUID PRIMARY KEY,
  name          VARCHAR(100) NOT NULL,
  scope         ums.role_scope NOT NULL DEFAULT 'global',
  tenant_id     UUID,
  description   TEXT,
  created_at    TIMESTAMPTZ
);

CREATE TABLE ums.permissions (
  id            UUID PRIMARY KEY,
  name          VARCHAR(150) NOT NULL,
  description   TEXT,
  created_at    TIMESTAMPTZ
);

CREATE TABLE ums.role_permissions (
  id              UUID PRIMARY KEY,
  role_id         UUID NOT NULL,
  permission_id   UUID NOT NULL,
  created_at      TIMESTAMPTZ
);

CREATE TABLE ums.user_roles (
  id          UUID PRIMARY KEY,
  user_id     UUID NOT NULL,
  role_id     UUID NOT NULL,
  created_at  TIMESTAMPTZ
);


-- =========================
-- MFA
-- =========================

CREATE TABLE ums.mfa_methods (
  id              UUID PRIMARY KEY,
  user_id         UUID NOT NULL,
  method          ums.mfa_method_type,
  secret          TEXT,
  target          TEXT,
  verified_at     TIMESTAMPTZ,
  created_at      TIMESTAMPTZ
);

CREATE TABLE ums.mfa_recovery_codes (
  id              UUID PRIMARY KEY,
  user_id         UUID NOT NULL,
  code_hash       TEXT UNIQUE NOT NULL,
  created_at      TIMESTAMPTZ
);

CREATE TABLE ums.mfa_challenges (
  id              UUID PRIMARY KEY,
  user_id         UUID NOT NULL,
  method          ums.mfa_method_type,
  challenge_hash  TEXT,
  expires_at      TIMESTAMPTZ,
  created_at      TIMESTAMPTZ
);

COMMIT;
