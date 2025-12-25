-- Safe migration: Add project permissions if they don't exist
-- This script can be run multiple times safely

-- Insert project permissions (skip if already exists)
DO $$
BEGIN
    -- project:create
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'project:create') THEN
        INSERT INTO permissions (name, description) VALUES ('project:create', 'Can create projects');
    END IF;
    
    -- project:edit
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'project:edit') THEN
        INSERT INTO permissions (name, description) VALUES ('project:edit', 'Can edit projects');
    END IF;
    
    -- project:delete
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'project:delete') THEN
        INSERT INTO permissions (name, description) VALUES ('project:delete', 'Can delete projects');
    END IF;
    
    -- project:publish (CRITICAL for admin approval)
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'project:publish') THEN
        INSERT INTO permissions (name, description) VALUES ('project:publish', 'Can publish projects (admin only)');
    END IF;
    
    -- project:view
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'project:view') THEN
        INSERT INTO permissions (name, description) VALUES ('project:view', 'Can view all projects including unpublished');
    END IF;
    
    -- project_vote:create
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'project_vote:create') THEN
        INSERT INTO permissions (name, description) VALUES ('project_vote:create', 'Can vote for projects');
    END IF;
    
    -- project_vote:delete
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'project_vote:delete') THEN
        INSERT INTO permissions (name, description) VALUES ('project_vote:delete', 'Can delete project votes');
    END IF;
    
    -- project_vote:view
    IF NOT EXISTS (SELECT 1 FROM permissions WHERE name = 'project_vote:view') THEN
        INSERT INTO permissions (name, description) VALUES ('project_vote:view', 'Can view all project votes');
    END IF;
END $$;

-- Assign project permissions to Admin role (role_id = 1)
INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, id FROM permissions WHERE name IN (
    'project:create',
    'project:edit',
    'project:delete',
    'project:publish',
    'project:view',
    'project_vote:create',
    'project_vote:delete',
    'project_vote:view'
)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Assign basic project permissions to Computizen role (role_id = 2)
INSERT INTO role_permissions (role_id, permission_id)
SELECT 2, id FROM permissions WHERE name IN (
    'project:create',
    'project_vote:create'
)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Verify permissions were added
SELECT 'Project permissions added successfully!' as status;
SELECT name, description FROM permissions WHERE name LIKE 'project%' ORDER BY name;
