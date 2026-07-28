-- Adds a job category taxonomy. Categories are a fixed enum (not a free-text
-- field or separate table) because JOBHOO's brief calls for a small,
-- deliberately curated set of categories used for filtering and
-- classification — not an open-ended taxonomy recruiters can expand.

CREATE TYPE job_category AS ENUM (
    'engineering_product',
    'design_creative',
    'sales_marketing',
    'data_analytics',
    'operations_support'
);

ALTER TABLE jobs ADD COLUMN category job_category NOT NULL DEFAULT 'engineering_product';
ALTER TABLE jobs ALTER COLUMN category DROP DEFAULT;

CREATE INDEX idx_jobs_category ON jobs (category);
