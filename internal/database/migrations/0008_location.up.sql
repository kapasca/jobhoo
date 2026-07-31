ALTER TABLE jobs
    ADD COLUMN country TEXT,
    ADD COLUMN state   TEXT;

ALTER TABLE jobs DROP COLUMN location;
