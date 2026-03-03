BEGIN;

-- =========================
-- USERS
-- =========================

CREATE INDEX idx_users_username
  ON ums.users (username);

CREATE INDEX idx_users_email
  ON ums.users (email);

CREATE INDEX idx_users_status
  ON ums.users (status);

CREATE INDEX idx_users_user_level
  ON ums.users (user_level);

CREATE INDEX idx_profiles_user_id
  ON ums.profiles (user_id);

-- =========================
-- REFRESH TOKENS
-- =========================

CREATE INDEX idx_refresh_tokens_user_id
  ON ums.refresh_tokens (user_id);

CREATE INDEX idx_refresh_tokens_device_id
  ON ums.refresh_tokens (device_id);

CREATE INDEX idx_refresh_tokens_expires_at
  ON ums.refresh_tokens (expires_at);

-- =========================
-- USER DEVICES
-- =========================

CREATE INDEX idx_user_devices_user_id
  ON ums.user_devices (user_id);

CREATE INDEX idx_user_devices_device_id
  ON ums.user_devices (device_id);


-- =========================
-- ONE TIME TOKENS
-- =========================

CREATE INDEX idx_one_time_tokens_user_id
  ON ums.one_time_tokens (user_id);

CREATE INDEX idx_one_time_tokens_purpose
  ON ums.one_time_tokens (purpose);

CREATE INDEX idx_one_time_tokens_expires_at
  ON ums.one_time_tokens (expires_at);

-- =========================
-- MFA
-- =========================

CREATE INDEX idx_mfa_methods_user_id
  ON ums.mfa_methods (user_id);

CREATE INDEX idx_mfa_methods_method
  ON ums.mfa_methods (method);

CREATE INDEX idx_mfa_recovery_codes_user_id
  ON ums.mfa_recovery_codes (user_id);

CREATE INDEX idx_mfa_challenges_user_id
  ON ums.mfa_challenges (user_id);

CREATE INDEX idx_mfa_challenges_expires_at
  ON ums.mfa_challenges (expires_at);


-- =========================
-- RBAC INDEXES
-- =========================

CREATE INDEX idx_roles_scope
  ON ums.roles (scope);

CREATE UNIQUE INDEX uq_roles_global_name
  ON ums.roles (name)
  WHERE scope = 'global';

CREATE INDEX idx_role_permissions_role_id
  ON ums.role_permissions (role_id);

CREATE INDEX idx_role_permissions_permission_id
  ON ums.role_permissions (permission_id);

CREATE INDEX idx_user_roles_user_id
  ON ums.user_roles (user_id);

CREATE INDEX idx_user_roles_role_id
  ON ums.user_roles (role_id);

COMMIT;
