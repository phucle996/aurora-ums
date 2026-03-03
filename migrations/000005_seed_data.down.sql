-- =========================
-- USER ROLES
-- =========================

DELETE FROM ums.user_roles
WHERE role_id IN (
  SELECT id FROM ums.roles WHERE name IN ('root', 'admin', 'user') AND scope = 'global'
);

-- =========================
-- ROLE PERMISSIONS
-- =========================

DELETE FROM ums.role_permissions
WHERE role_id IN (
  SELECT id FROM ums.roles WHERE name IN ('root', 'admin', 'user') AND scope = 'global'
);

-- =========================
-- PERMISSIONS
-- =========================

DELETE FROM ums.permissions
WHERE name IN (
  'user.read',
  'user.write',
  'role.read',
  'role.write',
  'permission.read',
  'permission.write'
);

-- =========================
-- ROLES
-- =========================

DELETE FROM ums.roles
WHERE name IN ('root', 'admin', 'user')
  AND scope = 'global';

-- =========================
-- USERS
-- =========================

DELETE FROM ums.mfa_challenges
WHERE user_id IN (
  SELECT id FROM ums.users WHERE username = 'root'
);

DELETE FROM ums.mfa_recovery_codes
WHERE user_id IN (
  SELECT id FROM ums.users WHERE username = 'root'
);

DELETE FROM ums.mfa_methods
WHERE user_id IN (
  SELECT id FROM ums.users WHERE username = 'root'
);

DELETE FROM ums.one_time_tokens
WHERE user_id IN (
  SELECT id FROM ums.users WHERE username = 'root'
);

DELETE FROM ums.refresh_tokens
WHERE user_id IN (
  SELECT id FROM ums.users WHERE username = 'root'
);

DELETE FROM ums.user_devices
WHERE user_id IN (
  SELECT id FROM ums.users WHERE username = 'root'
);

DELETE FROM ums.user_roles
WHERE user_id IN (
  SELECT id FROM ums.users WHERE username = 'root'
);

DELETE FROM ums.profiles
WHERE user_id IN (
  SELECT id FROM ums.users WHERE username = 'root'
);

DELETE FROM ums.users
WHERE username = 'root';
