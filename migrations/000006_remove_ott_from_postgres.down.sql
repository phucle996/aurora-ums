BEGIN;

DO $$
BEGIN
  CREATE TYPE one_time_token_purpose AS ENUM ('account_verify', 'password_reset');
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS one_time_tokens (
  id              UUID PRIMARY KEY,
  user_id         UUID NOT NULL,
  token_hash      TEXT UNIQUE NOT NULL,
  purpose         one_time_token_purpose,
  expires_at      TIMESTAMPTZ,
  created_at      TIMESTAMPTZ
);

DO $$
BEGIN
  ALTER TABLE one_time_tokens
    ADD CONSTRAINT fk_one_time_tokens_user
    FOREIGN KEY (user_id) REFERENCES users(id);
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

DO $$
BEGIN
  ALTER TABLE one_time_tokens
    ADD CONSTRAINT uniq_ott_user_purpose
    UNIQUE (user_id, purpose);
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_one_time_tokens_user_id
  ON one_time_tokens (user_id);

CREATE INDEX IF NOT EXISTS idx_one_time_tokens_purpose
  ON one_time_tokens (purpose);

CREATE INDEX IF NOT EXISTS idx_one_time_tokens_expires_at
  ON one_time_tokens (expires_at);

COMMIT;
