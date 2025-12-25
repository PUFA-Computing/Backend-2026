-- Drop indexes
DROP INDEX IF EXISTS idx_projects_pending;
DROP INDEX IF EXISTS idx_projects_approved_by;

-- Remove approval-related columns from projects table
ALTER TABLE projects
DROP COLUMN IF EXISTS rejection_reason,
DROP COLUMN IF EXISTS approved_at,
DROP COLUMN IF EXISTS approved_by;
