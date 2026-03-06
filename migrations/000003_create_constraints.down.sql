BEGIN;

-- =========================
-- CORE IDENTITY
-- =========================

-- =========================
-- MFA
-- =========================

ALTER TABLE mfa_challenges
  DROP CONSTRAINT IF EXISTS fk_mfa_challenges_user;

ALTER TABLE mfa_recovery_codes
  DROP CONSTRAINT IF EXISTS fk_mfa_recovery_codes_user;

ALTER TABLE mfa_methods
  DROP CONSTRAINT IF EXISTS fk_mfa_methods_user;

-- =========================
-- AUTHENTICATION
-- =========================

ALTER TABLE profiles
  DROP CONSTRAINT IF EXISTS uq_profiles_user_id;

ALTER TABLE profiles
  DROP CONSTRAINT IF EXISTS fk_profiles_user;

ALTER TABLE user_devices
  DROP CONSTRAINT IF EXISTS fk_user_devices_user;

ALTER TABLE refresh_tokens
  DROP CONSTRAINT IF EXISTS fk_refresh_tokens_device;

ALTER TABLE refresh_tokens
  DROP CONSTRAINT IF EXISTS fk_refresh_tokens_user;

-- =========================
-- RBAC CONSTRAINTS
-- =========================

ALTER TABLE user_roles
  DROP CONSTRAINT IF EXISTS uq_user_roles_user_id_role_id;

ALTER TABLE user_roles
  DROP CONSTRAINT IF EXISTS fk_user_roles_role;

ALTER TABLE user_roles
  DROP CONSTRAINT IF EXISTS fk_user_roles_user;

ALTER TABLE role_permissions
  DROP CONSTRAINT IF EXISTS uq_role_permissions_role_id_permission_id;

ALTER TABLE role_permissions
  DROP CONSTRAINT IF EXISTS fk_role_permissions_permission;

ALTER TABLE role_permissions
  DROP CONSTRAINT IF EXISTS fk_role_permissions_role;

ALTER TABLE permissions
  DROP CONSTRAINT IF EXISTS uq_permissions_name;

ALTER TABLE roles
  DROP CONSTRAINT IF EXISTS ck_roles_scope_tenant;

ALTER TABLE roles
  DROP CONSTRAINT IF EXISTS uq_roles_name;
COMMIT;
