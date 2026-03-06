BEGIN;

DROP INDEX IF EXISTS idx_one_time_tokens_expires_at;
DROP INDEX IF EXISTS idx_one_time_tokens_purpose;
DROP INDEX IF EXISTS idx_one_time_tokens_user_id;

ALTER TABLE IF EXISTS one_time_tokens
  DROP CONSTRAINT IF EXISTS uniq_ott_user_purpose;

ALTER TABLE IF EXISTS one_time_tokens
  DROP CONSTRAINT IF EXISTS fk_one_time_tokens_user;

DROP TABLE IF EXISTS one_time_tokens;
DROP TYPE IF EXISTS one_time_token_purpose;

COMMIT;
