-- Remove role_permissions entries for candidate and vote permissions
DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE name IN (
        'candidate:create',
        'candidate:edit',
        'candidate:delete',
        'candidate:view',
        'vote:cast',
        'vote:view',
        'vote:delete'
    )
);

-- Remove candidate and vote permissions
DELETE FROM permissions WHERE name IN (
    'candidate:create',
    'candidate:edit',
    'candidate:delete',
    'candidate:view',
    'vote:cast',
    'vote:view',
    'vote:delete'
);
