BEGIN;

-- =========================
-- DROP ENUMS
-- =========================

DROP TYPE IF EXISTS role_scope;
DROP TYPE IF EXISTS mfa_method_type;
DROP TYPE IF EXISTS user_status;

-- =========================
-- DROP EXTENSION 
-- =========================
DROP EXTENSION IF EXISTS pgcrypto;


COMMIT;
