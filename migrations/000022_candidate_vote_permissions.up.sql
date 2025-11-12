-- Add permissions for candidate and vote management
INSERT INTO permissions (name, description) VALUES
    ('candidate:create', 'Create new candidates'),
    ('candidate:edit', 'Edit existing candidates'),
    ('candidate:delete', 'Delete candidates'),
    ('candidate:view', 'View candidate details'),
    ('vote:cast', 'Cast a vote'),
    ('vote:view', 'View all votes (admin)'),
    ('vote:delete', 'Delete votes (admin)');

-- Assign candidate and vote permissions to admin role (role_id = 1)
-- Get the permission IDs dynamically
INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, id FROM permissions WHERE name IN (
    'candidate:create',
    'candidate:edit',
    'candidate:delete',
    'candidate:view',
    'vote:cast',
    'vote:view',
    'vote:delete'
);

-- Assign basic voting permission to user role (role_id = 2)
INSERT INTO role_permissions (role_id, permission_id)
SELECT 2, id FROM permissions WHERE name IN (
    'candidate:view',
    'vote:cast'
);
