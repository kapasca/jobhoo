-- Convert category from a rigid enum to free text so recruiters can enter
-- custom categories. Existing rows keep their values ('engineering_product'
-- etc.) as plain text; the seed will repopulate with human-readable strings.
ALTER TABLE jobs ALTER COLUMN category TYPE TEXT USING category::text;
DROP TYPE job_category;
