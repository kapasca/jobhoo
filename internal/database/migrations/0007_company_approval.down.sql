DROP INDEX IF EXISTS idx_companies_status;
ALTER TABLE companies DROP COLUMN IF EXISTS rejection_reason;
ALTER TABLE companies DROP COLUMN IF EXISTS approved_by;
ALTER TABLE companies DROP COLUMN IF EXISTS approved_at;
ALTER TABLE companies DROP COLUMN IF EXISTS status;
DROP TYPE IF EXISTS company_status;
