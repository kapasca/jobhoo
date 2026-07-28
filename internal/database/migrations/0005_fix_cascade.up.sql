-- Fix foreign key constraint: jobs.created_by should cascade delete
-- so that seeding (DELETE FROM users) properly cleans up dependent jobs.

ALTER TABLE jobs 
  DROP CONSTRAINT jobs_created_by_fkey,
  ADD CONSTRAINT jobs_created_by_fkey 
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE;
