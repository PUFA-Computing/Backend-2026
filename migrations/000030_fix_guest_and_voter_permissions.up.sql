-- Fix Guest permissions (Role 6) to strictly be read-only
-- Remove all existing permissions for role 6
DELETE FROM role_permissions WHERE role_id = 6;

-- Assign ONLY read permissions to Guest (Role 6)
INSERT INTO role_permissions (role_id, permission_id)
SELECT 6, id FROM permissions WHERE name IN (
    'users:get',
    'users:list',
    'events:get',
    'events:list',
    'news:get',
    'news:list',
    'roles:get',
    'roles:list',
    'permissions:list',
    'aspirations:list',
    'project:view',
    'candidate:view',
    'vote:view',
    'project_vote:view'
) ON CONFLICT DO NOTHING;

-- Remove voting and sensitive write permissions from Roles 3, 4, 5, 6
-- Only Role 1 (Admin) and Role 2 (Computizen) should have vote:cast, project_vote:create, project_vote:delete
DELETE FROM role_permissions
WHERE role_id NOT IN (1, 2)
AND permission_id IN (
    SELECT id FROM permissions WHERE name IN (
        'vote:cast',
        'project_vote:create',
        'project_vote:delete'
    )
);

-- Ensure Role 1 and Role 2 have the necessary voting permissions just in case
INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, id FROM permissions WHERE name IN (
    'vote:cast',
    'project_vote:create',
    'project_vote:delete'
) ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 2, id FROM permissions WHERE name IN (
    'vote:cast',
    'project_vote:create',
    'project_vote:delete'
) ON CONFLICT DO NOTHING;
