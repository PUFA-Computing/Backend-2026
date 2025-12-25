-- Ensure permissions table exists
CREATE TABLE IF NOT EXISTS permissions (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT now()
);

-- Ensure unique constraint exists for permissions.name
CREATE UNIQUE INDEX IF NOT EXISTS idx_permissions_name_unique
ON permissions(name);

-- Ensure role_permissions table exists
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    PRIMARY KEY (role_id, permission_id)
);

-- Insert project permissions
INSERT INTO permissions (name, description) VALUES
    ('project:create', 'Can create projects'),
    ('project:edit', 'Can edit projects'),
    ('project:delete', 'Can delete projects'),
    ('project:publish', 'Can publish projects (admin only)'),
    ('project:view', 'Can view all projects including unpublished'),
    ('project_vote:create', 'Can vote for projects'),
    ('project_vote:delete', 'Can delete project votes'),
    ('project_vote:view', 'Can view all project votes')
ON CONFLICT (name) DO NOTHING;

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
ON CONFLICT DO NOTHING;

-- Assign basic project permissions to Computizen role (role_id = 2)
INSERT INTO role_permissions (role_id, permission_id)
SELECT 2, id FROM permissions WHERE name IN (
    'project:create',
    'project_vote:create'
)
ON CONFLICT DO NOTHING;
