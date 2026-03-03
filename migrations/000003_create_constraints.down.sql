BEGIN;

-- =========================
-- CORE IDENTITY
-- =========================

-- =========================
-- MFA
-- =========================

ALTER TABLE ums.mfa_challenges
  DROP CONSTRAINT IF EXISTS fk_mfa_challenges_user;

ALTER TABLE ums.mfa_recovery_codes
  DROP CONSTRAINT IF EXISTS fk_mfa_recovery_codes_user;

ALTER TABLE ums.mfa_methods
  DROP CONSTRAINT IF EXISTS fk_mfa_methods_user;

-- =========================
-- AUTHENTICATION
-- =========================

ALTER TABLE ums.one_time_tokens
  DROP CONSTRAINT IF EXISTS fk_one_time_tokens_user;

ALTER TABLE ums.profiles
  DROP CONSTRAINT IF EXISTS uq_profiles_user_id;

ALTER TABLE ums.profiles
  DROP CONSTRAINT IF EXISTS fk_profiles_user;

ALTER TABLE ums.user_devices
  DROP CONSTRAINT IF EXISTS fk_user_devices_user;

ALTER TABLE ums.refresh_tokens
  DROP CONSTRAINT IF EXISTS fk_refresh_tokens_device;

ALTER TABLE ums.refresh_tokens
  DROP CONSTRAINT IF EXISTS fk_refresh_tokens_user;

ALTER TABLE ums.one_time_tokens
  DROP CONSTRAINT IF EXISTS uniq_ott_user_purpose;

-- =========================
-- RBAC CONSTRAINTS
-- =========================

ALTER TABLE ums.user_roles
  DROP CONSTRAINT IF EXISTS uq_user_roles_user_id_role_id;

ALTER TABLE ums.user_roles
  DROP CONSTRAINT IF EXISTS fk_user_roles_role;

ALTER TABLE ums.user_roles
  DROP CONSTRAINT IF EXISTS fk_user_roles_user;

ALTER TABLE ums.role_permissions
  DROP CONSTRAINT IF EXISTS uq_role_permissions_role_id_permission_id;

ALTER TABLE ums.role_permissions
  DROP CONSTRAINT IF EXISTS fk_role_permissions_permission;

ALTER TABLE ums.role_permissions
  DROP CONSTRAINT IF EXISTS fk_role_permissions_role;

ALTER TABLE ums.permissions
  DROP CONSTRAINT IF EXISTS uq_permissions_name;

ALTER TABLE ums.roles
  DROP CONSTRAINT IF EXISTS ck_roles_scope_tenant;

ALTER TABLE ums.roles
  DROP CONSTRAINT IF EXISTS uq_roles_name;
COMMIT;
