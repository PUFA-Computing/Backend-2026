-- ============================================================================
-- STEP 0: AUDIT — count rows with legacy URLs (read-only, safe to run anytime)
-- ============================================================================

\echo '=== Legacy URL audit ==='

SELECT 'events.thumbnail' AS location, COUNT(*) AS rows_with_legacy_url
FROM events
WHERE thumbnail LIKE '%pufacompsci.my.id%' OR thumbnail LIKE '%id.pufacomputing.live%'

UNION ALL

SELECT 'news.thumbnail', COUNT(*)
FROM news
WHERE thumbnail LIKE '%pufacompsci.my.id%' OR thumbnail LIKE '%id.pufacomputing.live%'

UNION ALL

SELECT 'projects.image_url', COUNT(*)
FROM projects
WHERE image_url LIKE '%pufacompsci.my.id%' OR image_url LIKE '%id.pufacomputing.live%'

UNION ALL

SELECT 'users.profile_picture', COUNT(*)
FROM users
WHERE profile_picture LIKE '%pufacompsci.my.id%' OR profile_picture LIKE '%id.pufacomputing.live%'

UNION ALL

SELECT 'users.twofa_image', COUNT(*)
FROM users
WHERE twofa_image LIKE '%pufacompsci.my.id%' OR twofa_image LIKE '%id.pufacomputing.live%'

UNION ALL

SELECT 'candidates.profile_picture', COUNT(*)
FROM candidates
WHERE profile_picture LIKE '%pufacompsci.my.id%' OR profile_picture LIKE '%id.pufacomputing.live%';

\echo ''
\echo 'Sample of legacy URLs found (first 5 per table):'
\echo ''
\echo '--- events ---'
SELECT id, title, thumbnail FROM events WHERE thumbnail LIKE '%pufacompsci.my.id%' OR thumbnail LIKE '%id.pufacomputing.live%' LIMIT 5;

\echo '--- projects ---'
SELECT id, title, image_url FROM projects WHERE image_url LIKE '%pufacompsci.my.id%' OR image_url LIKE '%id.pufacomputing.live%' LIMIT 5;

\echo '--- news ---'
SELECT id, title, thumbnail FROM news WHERE thumbnail LIKE '%pufacompsci.my.id%' OR thumbnail LIKE '%id.pufacomputing.live%' LIMIT 5;
