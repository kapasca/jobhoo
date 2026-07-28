-- Rollback: revert jobs.created_by foreign key to not cascade delete

ALTER TABLE jobs 
  DROP CONSTRAINT jobs_created_by_fkey,
  ADD CONSTRAINT jobs_created_by_fkey 
    FOREIGN KEY (created_by) REFERENCES users(id);
