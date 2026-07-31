-- Removing enum values requires recreating the type; mark blacklisted rows
-- as rejected before any manual cleanup.
UPDATE companies SET status = 'rejected' WHERE status = 'blacklisted';
