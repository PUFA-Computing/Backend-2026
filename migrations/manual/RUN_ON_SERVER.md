# Run-on-server: Migrate legacy image URLs → new R2 domain

These steps are executed via SSH on the production server (`188.166.212.22`).

**Target:** Replace `https://pufacompsci.my.id` and `https://id.pufacomputing.live`
with `https://pufacompsci.online` in all image-URL columns.

---

## Step 1 — SSH into the server

```bash
ssh root@188.166.212.22     # or whatever user you use
```

## Step 2 — BACKUP the database (mandatory)

```bash
# Make a backup directory
mkdir -p ~/db-backups
cd ~/db-backups

# Dump the database. If you're using Docker, swap to the docker exec form below.
PGPASSWORD='<YOUR_DB_PASSWORD>' pg_dump \
  -h 127.0.0.1 \
  -p 5432 \
  -U pufa_user \
  -d pufa_26 \
  -F p \
  -f pufa_26_before_url_migration_$(date +%Y%m%d_%H%M%S).sql

# OR if Postgres runs in Docker (adjust container name):
# docker exec -t <postgres_container_name> pg_dump -U pufa_user pufa_26 \
#   > pufa_26_before_url_migration_$(date +%Y%m%d_%H%M%S).sql

ls -lh pufa_26_before_url_migration_*.sql    # confirm file is not empty
```

**Do not proceed until you see a backup file with a reasonable size (MB-range).**

## Step 3 — Upload the SQL migration files

From your **local machine** (not server) run:

```bash
# From the Backend-2026 directory
scp migrations/manual/00_check_legacy_urls.sql root@188.166.212.22:~/db-backups/
scp migrations/manual/01_replace_legacy_urls.sql root@188.166.212.22:~/db-backups/
```

## Step 4 — Audit: see how many rows will change (read-only)

Back in your SSH session:

```bash
cd ~/db-backups

PGPASSWORD='<YOUR_DB_PASSWORD>' psql \
  -h 127.0.0.1 -p 5432 -U pufa_user -d pufa_26 \
  -f 00_check_legacy_urls.sql
```

This only does `SELECT COUNT(*)`. You'll see something like:

```
        location           | rows_with_legacy_url
---------------------------+----------------------
 events.thumbnail          |                   12
 news.thumbnail            |                    8
 ...
```

**Note the numbers. They tell you exactly how many rows will change.**

## Step 5 — Run the migration (still inside a transaction, NOT committed)

```bash
PGPASSWORD='<YOUR_DB_PASSWORD>' psql \
  -h 127.0.0.1 -p 5432 -U pufa_user -d pufa_26
```

You're now in the `psql` prompt. Run:

```sql
\i /root/db-backups/01_replace_legacy_urls.sql
```

The script will:
- Begin a transaction
- Run all `UPDATE` statements
- Show post-migration verification counts (all should be `0`)
- Stop **without** committing

## Step 6 — Verify and commit

Look at the verification output. If **all `still_legacy` counts are 0**:

```sql
COMMIT;
```

If anything looks wrong:

```sql
ROLLBACK;
```

After `COMMIT`, exit psql:

```sql
\q
```

## Step 7 — Test the API

Still on the server (or from your laptop):

```bash
curl -s https://compsci.president.ac.id/api/v1/events \
  | grep -oE 'https://[^"]+\.(jpg|jpeg|png)' \
  | sort -u | head -20
```

Every URL should now start with `https://pufacompsci.online/...`. None with
`pufacompsci.my.id` or `id.pufacomputing.live`.

## Step 8 — If something went horribly wrong, restore

```bash
PGPASSWORD='<YOUR_DB_PASSWORD>' psql \
  -h 127.0.0.1 -p 5432 -U pufa_user -d pufa_26 \
  -f pufa_26_before_url_migration_YYYYMMDD_HHMMSS.sql
```

---

## Notes on production safety

- Run during low-traffic hours if possible.
- The migration touches at most a few hundred rows — it completes in seconds.
- All writes are inside `BEGIN` ... `COMMIT`, so readers see consistent data.
- The frontend caches images on the client; users may need a hard refresh
  (Cmd+Shift+R / Ctrl+Shift+R) to see the new URLs immediately.
