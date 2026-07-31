-- Job lifecycle: recruiters can schedule when a job automatically becomes
-- visible/invisible to candidates, on top of manual open/close via status.
--
-- Deliberately NOT using a background scheduler/cron to flip `status` when
-- these dates pass — that would mean job visibility depends on a worker
-- process staying alive and running on time. Instead, "is this job
-- currently open" is computed at query time in JobsRepo.ListPublished
-- (status = 'published' AND opens_at/closes_at bracket now()), so
-- visibility is always correct regardless of any scheduler.

ALTER TABLE jobs ADD COLUMN opens_at TIMESTAMPTZ;
ALTER TABLE jobs ADD COLUMN closes_at TIMESTAMPTZ;

COMMENT ON COLUMN jobs.opens_at IS 'If set, job is hidden from public listing until this time even if status=published.';
COMMENT ON COLUMN jobs.closes_at IS 'If set, job is hidden from public listing after this time even if status=published.';
