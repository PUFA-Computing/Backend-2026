-- Remove project permissions from roles
DELETE FROM role_permissions 
WHERE permission_id IN (
    SELECT id FROM permissions WHERE name LIKE 'project:%' OR name LIKE 'project_vote:%'
);

-- Remove project permissions
DELETE FROM permissions WHERE name LIKE 'project:%' OR name LIKE 'project_vote:%';
