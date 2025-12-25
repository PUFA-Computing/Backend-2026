-- Manually assign project permissions to Admin role (role_id = 1)
-- This avoids the ON CONFLICT issue

-- First, get the permission IDs
DO $$
DECLARE
    perm_id BIGINT;
BEGIN
    -- Loop through each project permission and assign to admin if not exists
    FOR perm_id IN 
        SELECT id FROM permissions WHERE name IN (
            'project:create',
            'project:edit',
            'project:delete',
            'project:publish',
            'project:view',
            'project_vote:create',
            'project_vote:delete',
            'project_vote:view'
        )
    LOOP
        -- Check if this permission is already assigned to admin
        IF NOT EXISTS (
            SELECT 1 FROM role_permissions 
            WHERE role_id = 1 AND permission_id = perm_id
        ) THEN
            INSERT INTO role_permissions (role_id, permission_id) 
            VALUES (1, perm_id);
            RAISE NOTICE 'Assigned permission % to Admin role', perm_id;
        END IF;
    END LOOP;
    
    -- Assign basic permissions to Computizen role (role_id = 2)
    FOR perm_id IN 
        SELECT id FROM permissions WHERE name IN (
            'project:create',
            'project_vote:create'
        )
    LOOP
        IF NOT EXISTS (
            SELECT 1 FROM role_permissions 
            WHERE role_id = 2 AND permission_id = perm_id
        ) THEN
            INSERT INTO role_permissions (role_id, permission_id) 
            VALUES (2, perm_id);
            RAISE NOTICE 'Assigned permission % to Computizen role', perm_id;
        END IF;
    END LOOP;
END $$;

-- Verify admin has project permissions
SELECT 'Admin role project permissions:' as info;
SELECT p.name, p.description 
FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
WHERE rp.role_id = 1 AND p.name LIKE 'project%'
ORDER BY p.name;
