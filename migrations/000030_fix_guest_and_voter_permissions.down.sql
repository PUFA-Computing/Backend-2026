-- Revert Guest permissions (Role 6) back to original
-- NOTE: We just add back the exact permissions Guest had before.
-- From 000006_role_permissions.up.sql:
-- 1, 6, 11, 13, 18, 2, 3, 27, 34
INSERT INTO role_permissions (role_id, permission_id)
VALUES
    (6, 1),
    (6, 6),
    (6, 11),
    (6, 13),
    (6, 18),
    (6, 2),
    (6, 3),
    (6, 27),
    (6, 34)
ON CONFLICT DO NOTHING;

-- No need to explicitly re-add project permissions to Roles 3, 4, 5 
-- because they probably didn't have them originally, but we can't be sure 
-- without a full snapshot. This is a best-effort rollback.
