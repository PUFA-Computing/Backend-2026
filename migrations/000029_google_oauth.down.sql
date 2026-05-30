-- Best-effort rollback. Note: any Google-only rows (where password IS NULL)
-- will block the NOT NULL restore. Drop / migrate those first if needed.

DROP INDEX IF EXISTS users_google_sub_unique;

ALTER TABLE users
    DROP COLUMN IF EXISTS google_sub,
    DROP COLUMN IF EXISTS auth_provider,
    DROP COLUMN IF EXISTS profile_completed;

ALTER TABLE users ALTER COLUMN password SET NOT NULL;
