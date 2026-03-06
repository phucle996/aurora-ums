BEGIN;

-- =========================
-- CORE IDENTITY
-- =========================

-- =========================
-- AUTHENTICATION
-- =========================

ALTER TABLE refresh_tokens
  ADD CONSTRAINT fk_refresh_tokens_user
  FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE user_devices
  ADD CONSTRAINT fk_user_devices_user
  FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE profiles
  ADD CONSTRAINT fk_profiles_user
  FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE profiles
  ADD CONSTRAINT uq_profiles_user_id UNIQUE (user_id);

-- =========================
-- MFA
-- =========================

ALTER TABLE mfa_methods
  ADD CONSTRAINT fk_mfa_methods_user
  FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE mfa_recovery_codes
  ADD CONSTRAINT fk_mfa_recovery_codes_user
  FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE mfa_challenges
  ADD CONSTRAINT fk_mfa_challenges_user
  FOREIGN KEY (user_id) REFERENCES users(id);



-- =========================
-- RBAC CONSTRAINTS
-- =========================

ALTER TABLE permissions
  ADD CONSTRAINT uq_permissions_name UNIQUE (name);

ALTER TABLE role_permissions
  ADD CONSTRAINT fk_role_permissions_role
  FOREIGN KEY (role_id) REFERENCES roles(id);

ALTER TABLE role_permissions
  ADD CONSTRAINT fk_role_permissions_permission
  FOREIGN KEY (permission_id) REFERENCES permissions(id);

ALTER TABLE role_permissions
  ADD CONSTRAINT uq_role_permissions_role_id_permission_id
  UNIQUE (role_id, permission_id);

ALTER TABLE user_roles
  ADD CONSTRAINT fk_user_roles_user
  FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE user_roles
  ADD CONSTRAINT fk_user_roles_role
  FOREIGN KEY (role_id) REFERENCES roles(id);

ALTER TABLE user_roles
  ADD CONSTRAINT uq_user_roles_user_id_role_id
  UNIQUE (user_id, role_id);


COMMIT;
