-- =========================
-- USER ROLES
-- =========================

DELETE FROM user_roles
WHERE role_id IN (
  SELECT id FROM roles WHERE name IN ('root', 'admin', 'user') AND scope = 'global'
);

-- =========================
-- ROLE PERMISSIONS
-- =========================

DELETE FROM role_permissions
WHERE role_id IN (
  SELECT id FROM roles WHERE name IN ('root', 'admin', 'user') AND scope = 'global'
);

-- =========================
-- PERMISSIONS
-- =========================

DELETE FROM permissions
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

DELETE FROM roles
WHERE name IN ('root', 'admin', 'user')
  AND scope = 'global';

-- =========================
-- USERS
-- =========================

DELETE FROM mfa_challenges
WHERE user_id IN (
  SELECT id FROM users WHERE username = 'root'
);

DELETE FROM mfa_recovery_codes
WHERE user_id IN (
  SELECT id FROM users WHERE username = 'root'
);

DELETE FROM mfa_methods
WHERE user_id IN (
  SELECT id FROM users WHERE username = 'root'
);

DELETE FROM refresh_tokens
WHERE user_id IN (
  SELECT id FROM users WHERE username = 'root'
);

DELETE FROM user_devices
WHERE user_id IN (
  SELECT id FROM users WHERE username = 'root'
);

DELETE FROM user_roles
WHERE user_id IN (
  SELECT id FROM users WHERE username = 'root'
);

DELETE FROM profiles
WHERE user_id IN (
  SELECT id FROM users WHERE username = 'root'
);

DELETE FROM users
WHERE username = 'root';
