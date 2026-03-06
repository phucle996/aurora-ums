-- =========================
-- USERS (ROOT)
-- =========================

WITH upserted_root_user AS (
  INSERT INTO users (
    id,
    username,
    email,
    password_hash,
    status,
    user_level,
    on_boarding,
    created_at,
    updated_at
  )
  VALUES (
    gen_random_uuid(),
    'root',
    'root@local',
    '$argon2id$v=19$m=65536,t=3,p=2$i2eNFJFekwk96FF39ydKNA$IFJsqZfji46DsDV2fOrMV9m26cs37f6waCLb37vAhUs',
    'active',
    0,
    FALSE,
    now(),
    now()
  )
  ON CONFLICT (username) DO UPDATE
  SET
    email = EXCLUDED.email,
    password_hash = EXCLUDED.password_hash,
    status = EXCLUDED.status,
    user_level = EXCLUDED.user_level,
    on_boarding = EXCLUDED.on_boarding,
    updated_at = now()
  RETURNING id
)
INSERT INTO profiles (
  id,
  user_id,
  full_name,
  created_at,
  updated_at
)
SELECT
  gen_random_uuid(),
  upserted_root_user.id,
  'Root User',
  now(),
  now()
FROM upserted_root_user
ON CONFLICT (user_id) DO UPDATE
SET
  full_name = EXCLUDED.full_name,
  updated_at = now();


-- =========================
-- ROLES
-- =========================

INSERT INTO roles (id, name, scope, tenant_id, description, created_at)
VALUES
  (gen_random_uuid(), 'root',  'global', NULL, 'System root role',  now()),
  (gen_random_uuid(), 'admin', 'global', NULL, 'System admin role', now()),
  (gen_random_uuid(), 'user',  'global', NULL, 'Default user role', now())
ON CONFLICT DO NOTHING;

-- =========================
-- PERMISSIONS
-- =========================

INSERT INTO permissions (id, name, description, created_at)
VALUES
  (gen_random_uuid(), 'user.read',         'Read users',          now()),
  (gen_random_uuid(), 'user.write',        'Create/update users', now()),
  (gen_random_uuid(), 'role.read',         'Read roles',          now()),
  (gen_random_uuid(), 'role.write',        'Manage roles',        now()),
  (gen_random_uuid(), 'permission.read',   'Read permissions',    now()),
  (gen_random_uuid(), 'permission.write',  'Manage permissions',  now())
ON CONFLICT DO NOTHING;

-- =========================
-- ROLE PERMISSIONS
-- =========================

-- root gets all permissions
INSERT INTO role_permissions (id, role_id, permission_id, created_at)
SELECT
  gen_random_uuid(),
  r.id,
  p.id,
  now()
FROM roles r
JOIN permissions p ON TRUE
WHERE r.name = 'root'
ON CONFLICT DO NOTHING;

-- admin gets all except permission.write
DELETE FROM role_permissions
WHERE role_id IN (
  SELECT id FROM roles WHERE name = 'admin'
);

INSERT INTO role_permissions (id, role_id, permission_id, created_at)
SELECT
  gen_random_uuid(),
  r.id,
  p.id,
  now()
FROM roles r
JOIN permissions p ON p.name IN (
  'user.read',
  'user.write',
  'role.read',
  'role.write',
  'permission.read'
)
WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

-- user gets read-only permissions
DELETE FROM role_permissions
WHERE role_id IN (
  SELECT id FROM roles WHERE name = 'user'
);

INSERT INTO role_permissions (id, role_id, permission_id, created_at)
SELECT
  gen_random_uuid(),
  r.id,
  p.id,
  now()
FROM roles r
JOIN permissions p ON p.name IN (
  'user.read',
  'role.read',
  'permission.read'
)
WHERE r.name = 'user'
ON CONFLICT DO NOTHING;

-- =========================
-- USER ROLES (ROOT USER)
-- =========================

INSERT INTO user_roles (id, user_id, role_id, created_at)
SELECT
  gen_random_uuid(),
  u.id,
  r.id,
  now()
FROM users u
JOIN roles r ON r.name = 'root' 
WHERE u.username = 'root'
ON CONFLICT DO NOTHING;
