BEGIN;


-- =========================
-- MFA
-- =========================

DROP INDEX IF EXISTS ums.idx_mfa_challenges_expires_at;
DROP INDEX IF EXISTS ums.idx_mfa_challenges_user_id;

DROP INDEX IF EXISTS ums.idx_mfa_recovery_codes_user_id;

DROP INDEX IF EXISTS ums.idx_mfa_methods_method;
DROP INDEX IF EXISTS ums.idx_mfa_methods_user_id;

-- =========================
-- ONE TIME TOKENS
-- =========================

DROP INDEX IF EXISTS ums.idx_one_time_tokens_expires_at;
DROP INDEX IF EXISTS ums.idx_one_time_tokens_purpose;
DROP INDEX IF EXISTS ums.idx_one_time_tokens_user_id;

-- =========================
-- USER DEVICES
-- =========================

DROP INDEX IF EXISTS ums.idx_user_devices_device_id;
DROP INDEX IF EXISTS ums.idx_user_devices_user_id;

-- =========================
-- REFRESH TOKENS
-- =========================

DROP INDEX IF EXISTS ums.idx_refresh_tokens_expires_at;
DROP INDEX IF EXISTS ums.idx_refresh_tokens_device_id;
DROP INDEX IF EXISTS ums.idx_refresh_tokens_user_id;

-- =========================
-- USERS
-- =========================

DROP INDEX IF EXISTS ums.idx_users_user_level;
DROP INDEX IF EXISTS ums.idx_profiles_user_id;
DROP INDEX IF EXISTS ums.idx_users_status;
DROP INDEX IF EXISTS ums.idx_users_email;
DROP INDEX IF EXISTS ums.idx_users_username;


-- =========================
-- RBAC INDEXES
-- =========================

DROP INDEX IF EXISTS ums.uq_roles_tenant_name;
DROP INDEX IF EXISTS ums.uq_roles_global_name;
DROP INDEX IF EXISTS ums.idx_roles_scope;

DROP INDEX IF EXISTS ums.idx_user_roles_role_id;
DROP INDEX IF EXISTS ums.idx_user_roles_user_id;
DROP INDEX IF EXISTS ums.idx_role_permissions_permission_id;
DROP INDEX IF EXISTS ums.idx_role_permissions_role_id;

COMMIT;
