-- Rollback email logs table

DROP TABLE IF EXISTS email_logs;
DROP INDEX IF EXISTS idx_email_logs_user_id;
DROP INDEX IF EXISTS idx_email_logs_recipient_email;
DROP INDEX IF EXISTS idx_email_logs_sent_at;
DROP INDEX IF EXISTS idx_email_logs_email_type;
