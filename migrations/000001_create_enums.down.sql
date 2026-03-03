BEGIN;

-- =========================
-- DROP ENUMS
-- =========================

DROP TYPE IF EXISTS ums.role_scope;
DROP TYPE IF EXISTS ums.one_time_token_purpose;
DROP TYPE IF EXISTS ums.mfa_method_type;
DROP TYPE IF EXISTS ums.user_status;

-- =========================
-- DROP EXTENSION 
-- =========================
DROP EXTENSION IF EXISTS pgcrypto;


COMMIT;
