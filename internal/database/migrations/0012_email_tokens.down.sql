-- Rollback email verification and password reset schema

DROP TABLE IF EXISTS password_resets;
DROP INDEX IF EXISTS idx_password_resets_token_hash;

DROP TABLE IF EXISTS email_verifications;
DROP INDEX IF EXISTS idx_email_verifications_token_hash;

ALTER TABLE users DROP COLUMN IF EXISTS email_verified;
