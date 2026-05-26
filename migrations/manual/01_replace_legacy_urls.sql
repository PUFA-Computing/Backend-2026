-- ============================================================================
-- STEP 1: REPLACE legacy image hosts with the new R2 public URL
-- ============================================================================

BEGIN;

\echo '>>> Updating events.thumbnail ...'
UPDATE events
SET thumbnail = REPLACE(thumbnail, 'https://pufacompsci.my.id', 'https://pufacompsci.online')
WHERE thumbnail LIKE '%pufacompsci.my.id%';

UPDATE events
SET thumbnail = REPLACE(thumbnail, 'https://id.pufacomputing.live', 'https://pufacompsci.online')
WHERE thumbnail LIKE '%id.pufacomputing.live%';

\echo '>>> Updating news.thumbnail ...'
UPDATE news
SET thumbnail = REPLACE(thumbnail, 'https://pufacompsci.my.id', 'https://pufacompsci.online')
WHERE thumbnail LIKE '%pufacompsci.my.id%';

UPDATE news
SET thumbnail = REPLACE(thumbnail, 'https://id.pufacomputing.live', 'https://pufacompsci.online')
WHERE thumbnail LIKE '%id.pufacomputing.live%';

\echo '>>> Updating projects.image_url ...'
UPDATE projects
SET image_url = REPLACE(image_url, 'https://pufacompsci.my.id', 'https://pufacompsci.online')
WHERE image_url LIKE '%pufacompsci.my.id%';

UPDATE projects
SET image_url = REPLACE(image_url, 'https://id.pufacomputing.live', 'https://pufacompsci.online')
WHERE image_url LIKE '%id.pufacomputing.live%';

\echo '>>> Updating users.profile_picture ...'
UPDATE users
SET profile_picture = REPLACE(profile_picture, 'https://pufacompsci.my.id', 'https://pufacompsci.online')
WHERE profile_picture LIKE '%pufacompsci.my.id%';

UPDATE users
SET profile_picture = REPLACE(profile_picture, 'https://id.pufacomputing.live', 'https://pufacompsci.online')
WHERE profile_picture LIKE '%id.pufacomputing.live%';

\echo '>>> Updating users.twofa_image ...'
UPDATE users
SET twofa_image = REPLACE(twofa_image, 'https://pufacompsci.my.id', 'https://pufacompsci.online')
WHERE twofa_image LIKE '%pufacompsci.my.id%';

UPDATE users
SET twofa_image = REPLACE(twofa_image, 'https://id.pufacomputing.live', 'https://pufacompsci.online')
WHERE twofa_image LIKE '%id.pufacomputing.live%';

\echo '>>> Updating candidates.profile_picture ...'
UPDATE candidates
SET profile_picture = REPLACE(profile_picture, 'https://pufacompsci.my.id', 'https://pufacompsci.online')
WHERE profile_picture LIKE '%pufacompsci.my.id%';

UPDATE candidates
SET profile_picture = REPLACE(profile_picture, 'https://id.pufacomputing.live', 'https://pufacompsci.online')
WHERE profile_picture LIKE '%id.pufacomputing.live%';

\echo ''
\echo '=== POST-MIGRATION VERIFICATION (all should be 0) ==='
SELECT 'events.thumbnail' AS location, COUNT(*) AS still_legacy FROM events WHERE thumbnail LIKE '%pufacompsci.my.id%' OR thumbnail LIKE '%id.pufacomputing.live%'
UNION ALL
SELECT 'news.thumbnail', COUNT(*) FROM news WHERE thumbnail LIKE '%pufacompsci.my.id%' OR thumbnail LIKE '%id.pufacomputing.live%'
UNION ALL
SELECT 'projects.image_url', COUNT(*) FROM projects WHERE image_url LIKE '%pufacompsci.my.id%' OR image_url LIKE '%id.pufacomputing.live%'
UNION ALL
SELECT 'users.profile_picture', COUNT(*) FROM users WHERE profile_picture LIKE '%pufacompsci.my.id%' OR profile_picture LIKE '%id.pufacomputing.live%'
UNION ALL
SELECT 'users.twofa_image', COUNT(*) FROM users WHERE twofa_image LIKE '%pufacompsci.my.id%' OR twofa_image LIKE '%id.pufacomputing.live%'
UNION ALL
SELECT 'candidates.profile_picture', COUNT(*) FROM candidates WHERE profile_picture LIKE '%pufacompsci.my.id%' OR profile_picture LIKE '%id.pufacomputing.live%';

\echo 'If all verification counts are 0, run COMMIT; otherwise ROLLBACK;'
