DROP TRIGGER IF EXISTS enforce_same_major ON votes;
DROP FUNCTION IF EXISTS check_same_major();
DROP TABLE IF EXISTS votes CASCADE;
