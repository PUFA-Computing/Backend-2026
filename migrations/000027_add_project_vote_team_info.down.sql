-- First, remove the columns from project_votes if they were added
ALTER TABLE project_votes
DROP CONSTRAINT IF EXISTS check_major,
DROP CONSTRAINT IF EXISTS check_batch;

ALTER TABLE project_votes
DROP COLUMN IF EXISTS project_members,
DROP COLUMN IF EXISTS linkedin_profiles,
DROP COLUMN IF EXISTS major,
DROP COLUMN IF EXISTS batch;
