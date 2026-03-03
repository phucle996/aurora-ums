BEGIN;

-- =========================
-- MFA
-- =========================

DROP TABLE IF EXISTS ums.mfa_challenges;
DROP TABLE IF EXISTS ums.mfa_recovery_codes;
DROP TABLE IF EXISTS ums.mfa_methods;

-- =========================
-- AUTHENTICATION
-- =========================

DROP TABLE IF EXISTS ums.one_time_tokens;
DROP TABLE IF EXISTS ums.user_devices;
DROP TABLE IF EXISTS ums.refresh_tokens;


-- =========================
-- RBAC
-- =========================
DROP TABLE IF EXISTS ums.user_roles;
DROP TABLE IF EXISTS ums.role_permissions;
DROP TABLE IF EXISTS ums.permissions;
DROP TABLE IF EXISTS ums.roles;


-- =========================
-- CORE IDENTITY
-- =========================

DROP TABLE IF EXISTS ums.profiles;
DROP TABLE IF EXISTS ums.users;

COMMIT;
