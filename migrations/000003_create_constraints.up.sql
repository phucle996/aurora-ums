BEGIN;

-- =========================
-- CORE IDENTITY
-- =========================

-- =========================
-- AUTHENTICATION
-- =========================

ALTER TABLE ums.refresh_tokens
  ADD CONSTRAINT fk_refresh_tokens_user
  FOREIGN KEY (user_id) REFERENCES ums.users(id);

ALTER TABLE ums.user_devices
  ADD CONSTRAINT fk_user_devices_user
  FOREIGN KEY (user_id) REFERENCES ums.users(id);

ALTER TABLE ums.one_time_tokens
  ADD CONSTRAINT fk_one_time_tokens_user
  FOREIGN KEY (user_id) REFERENCES ums.users(id);

ALTER TABLE ums.profiles
  ADD CONSTRAINT fk_profiles_user
  FOREIGN KEY (user_id) REFERENCES ums.users(id);

ALTER TABLE ums.profiles
  ADD CONSTRAINT uq_profiles_user_id UNIQUE (user_id);

  ALTER TABLE ums.one_time_tokens
ADD CONSTRAINT uniq_ott_user_purpose
UNIQUE (user_id, purpose);

-- =========================
-- MFA
-- =========================

ALTER TABLE ums.mfa_methods
  ADD CONSTRAINT fk_mfa_methods_user
  FOREIGN KEY (user_id) REFERENCES ums.users(id);

ALTER TABLE ums.mfa_recovery_codes
  ADD CONSTRAINT fk_mfa_recovery_codes_user
  FOREIGN KEY (user_id) REFERENCES ums.users(id);

ALTER TABLE ums.mfa_challenges
  ADD CONSTRAINT fk_mfa_challenges_user
  FOREIGN KEY (user_id) REFERENCES ums.users(id);



-- =========================
-- RBAC CONSTRAINTS
-- =========================

ALTER TABLE ums.permissions
  ADD CONSTRAINT uq_permissions_name UNIQUE (name);

ALTER TABLE ums.role_permissions
  ADD CONSTRAINT fk_role_permissions_role
  FOREIGN KEY (role_id) REFERENCES ums.roles(id);

ALTER TABLE ums.role_permissions
  ADD CONSTRAINT fk_role_permissions_permission
  FOREIGN KEY (permission_id) REFERENCES ums.permissions(id);

ALTER TABLE ums.role_permissions
  ADD CONSTRAINT uq_role_permissions_role_id_permission_id
  UNIQUE (role_id, permission_id);

ALTER TABLE ums.user_roles
  ADD CONSTRAINT fk_user_roles_user
  FOREIGN KEY (user_id) REFERENCES ums.users(id);

ALTER TABLE ums.user_roles
  ADD CONSTRAINT fk_user_roles_role
  FOREIGN KEY (role_id) REFERENCES ums.roles(id);

ALTER TABLE ums.user_roles
  ADD CONSTRAINT uq_user_roles_user_id_role_id
  UNIQUE (user_id, role_id);


COMMIT;
