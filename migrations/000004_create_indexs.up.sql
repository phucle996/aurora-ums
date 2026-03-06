BEGIN;

-- =========================
-- USERS
-- =========================

CREATE INDEX idx_users_username
  ON users (username);

CREATE INDEX idx_users_email
  ON users (email);

CREATE INDEX idx_users_status
  ON users (status);

CREATE INDEX idx_users_user_level
  ON users (user_level);

CREATE INDEX idx_profiles_user_id
  ON profiles (user_id);

-- =========================
-- REFRESH TOKENS
-- =========================

CREATE INDEX idx_refresh_tokens_user_id
  ON refresh_tokens (user_id);

CREATE INDEX idx_refresh_tokens_device_id
  ON refresh_tokens (device_id);

CREATE INDEX idx_refresh_tokens_expires_at
  ON refresh_tokens (expires_at);

-- =========================
-- USER DEVICES
-- =========================

CREATE INDEX idx_user_devices_user_id
  ON user_devices (user_id);

CREATE INDEX idx_user_devices_device_id
  ON user_devices (device_id);


-- =========================
-- MFA
-- =========================

CREATE INDEX idx_mfa_methods_user_id
  ON mfa_methods (user_id);

CREATE INDEX idx_mfa_methods_method
  ON mfa_methods (method);

CREATE INDEX idx_mfa_recovery_codes_user_id
  ON mfa_recovery_codes (user_id);

CREATE INDEX idx_mfa_challenges_user_id
  ON mfa_challenges (user_id);

CREATE INDEX idx_mfa_challenges_expires_at
  ON mfa_challenges (expires_at);


-- =========================
-- RBAC INDEXES
-- =========================

CREATE INDEX idx_roles_scope
  ON roles (scope);

CREATE UNIQUE INDEX uq_roles_global_name
  ON roles (name)
  WHERE scope = 'global';

CREATE INDEX idx_role_permissions_role_id
  ON role_permissions (role_id);

CREATE INDEX idx_role_permissions_permission_id
  ON role_permissions (permission_id);

CREATE INDEX idx_user_roles_user_id
  ON user_roles (user_id);

CREATE INDEX idx_user_roles_role_id
  ON user_roles (role_id);

COMMIT;
