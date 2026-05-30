-- Google OAuth + account-linking support.
--
-- We treat Google as a *linked identity* on top of the existing email/password
-- account. The same row in `users` can be reached via either method:
--   - password   -> rows that were created via /auth/register
--   - google     -> rows that were created via /auth/google (or password rows
--                   that later linked their Google account)
--   - both       -> after a password user logs in with the same Google email
--
-- profile_completed is FALSE only for fresh Google sign-ups that still need
-- to enter Student ID + batch (the second-stage form). All existing users
-- have a populated student_id, so we backfill them to TRUE.

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS google_sub VARCHAR(255),
    ADD COLUMN IF NOT EXISTS auth_provider VARCHAR(20) NOT NULL DEFAULT 'password',
    ADD COLUMN IF NOT EXISTS profile_completed BOOLEAN NOT NULL DEFAULT TRUE;

-- Unique index (partial – NULLs are excluded so multiple un-linked rows are fine)
CREATE UNIQUE INDEX IF NOT EXISTS users_google_sub_unique
    ON users (google_sub)
    WHERE google_sub IS NOT NULL;

-- Existing accounts already have student data – mark them completed.
UPDATE users
SET profile_completed = TRUE
WHERE student_id IS NOT NULL AND student_id <> '';

-- A Google-only account doesn't need a password. Allow NULL.
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;
