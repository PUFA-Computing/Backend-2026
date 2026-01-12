ALTER TABLE projects
DROP CONSTRAINT IF EXISTS check_project_major,
DROP CONSTRAINT IF EXISTS check_project_batch;

ALTER TABLE projects
DROP COLUMN IF EXISTS project_members,
DROP COLUMN IF EXISTS linkedin_profiles,
DROP COLUMN IF EXISTS major,
DROP COLUMN IF EXISTS batch;
