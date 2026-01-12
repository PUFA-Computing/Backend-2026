ALTER TABLE projects
ADD COLUMN project_members JSONB DEFAULT '[]'::jsonb,
ADD COLUMN linkedin_profiles JSONB DEFAULT '[]'::jsonb,
ADD COLUMN major VARCHAR(50),
ADD COLUMN batch INTEGER;

-- Add constraints
ALTER TABLE projects
ADD CONSTRAINT check_project_major CHECK (major IN ('information_system', 'informatics'));

ALTER TABLE projects
ADD CONSTRAINT check_project_batch CHECK (batch >= 2021 AND batch <= 2025);

-- Add comments for documentation
COMMENT ON COLUMN projects.project_members IS 'Array of team member full names (JSONB)';
COMMENT ON COLUMN projects.linkedin_profiles IS 'Array of LinkedIn profile URLs (JSONB)';
COMMENT ON COLUMN projects.major IS 'Project creator major: information_system or informatics';
COMMENT ON COLUMN projects.batch IS 'Project creator batch year (2021-2025)';
