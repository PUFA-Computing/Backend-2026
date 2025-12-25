-- Add approval-related fields to projects table
ALTER TABLE projects
ADD COLUMN approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
ADD COLUMN approved_at TIMESTAMP,
ADD COLUMN rejection_reason TEXT;

-- Create index for approved_by for faster queries
CREATE INDEX IF NOT EXISTS idx_projects_approved_by ON projects(approved_by);

-- Create index for filtering pending projects
CREATE INDEX IF NOT EXISTS idx_projects_pending ON projects(is_published, rejection_reason) 
WHERE is_published = false AND rejection_reason IS NULL;
