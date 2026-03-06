BEGIN;

-- =========================
-- MFA
-- =========================

DROP TABLE IF EXISTS mfa_challenges;
DROP TABLE IF EXISTS mfa_recovery_codes;
DROP TABLE IF EXISTS mfa_methods;

-- =========================
-- AUTHENTICATION
-- =========================

DROP TABLE IF EXISTS user_devices;
DROP TABLE IF EXISTS refresh_tokens;


-- =========================
-- RBAC
-- =========================
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;


-- =========================
-- CORE IDENTITY
-- =========================

DROP TABLE IF EXISTS profiles;
DROP TABLE IF EXISTS users;

COMMIT;
