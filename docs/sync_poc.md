# Remote Sync PoC Guide

This document explains how to run the minimal “remote backend pushes data into the central backend” workflow.

## 1. Central backend setup

1. **Apply database migration** (adds `source_host`, `remote_id`, and `sync_errors`):
   ```sql
   SOURCE deployment/migrations/20251129_sync_poc.sql;
   ```
2. **Set sync token** on the central node before starting the backend:
   ```powershell
   $env:SYNC_TOKEN = "sync-dev-token" # use a strong token in production
   ```
3. **Restart** the Go backend. It now exposes `POST /api/v1/sync/events` (see `internal/handlers/sync_handler.go`).
4. Optional sanity check:
   ```powershell
   curl -X POST http://127.0.0.1:8080/api/v1/sync/events `
        -H "Content-Type: application/json" `
        -H "X-SYNC-TOKEN: sync-dev-token" `
        -d '{"source_host":"test","table":"honeypot_instances","rows":[]}'
   ```
   Should return a success response (with zero rows processed).

## 2. Remote sync client

Files: `tools/remote_sync.py` (client) and `tools/remote_sync_config.example.json` (template).

1. Copy both files to each remote machine and create `remote_sync_config.json` from the template:
   - `source_host`: unique identifier for that machine (e.g., `10.0.0.12`).
   - `central_url`: e.g., `http://127.0.0.1:8080/api/v1/sync/events` (point to central IP).
   - `sync_token`: must match the central `SYNC_TOKEN`.
   - `mysql` block: credentials for the **local** database on that machine.
2. Install dependencies:
   ```bash
   pip install pymysql requests
   ```
3. Run once (for manual sync / testing):
   ```bash
   python remote_sync.py --config remote_sync_config.json --log-level DEBUG
   ```
4. Continuous sync options:
   - **Cron** (every minute):
     ```bash
     * * * * * /usr/bin/python3 /opt/remote-sync/remote_sync.py --config /opt/remote-sync/remote_sync_config.json >> /var/log/remote-sync.log 2>&1
     ```
   - **systemd timer** (Linux) – create a service running `python remote_sync.py --loop --interval 60 ...`.

State tracking (`sync_state.json`) stores the highest synced ID per table so repeated runs only push new rows. Failed pushes keep the old state, so the next run retries automatically.

## 3. Testing checklist (single machine)

1. Insert a dummy row into the local `honeypot_instance` table (ID > last synced value).
2. Run `remote_sync.py --config ... --run once` (default).
3. On the central DB, verify:
   ```sql
   SELECT id, source_host, remote_id FROM honeypot_instance WHERE source_host='127.0.0.2';
   ```
4. Run the script again without new data – `Inserted/Updated` should stay zero.
5. Intentionally break the token to confirm the script retries and `sync_errors` captures the failure.

## 4. Rolling out to 9 remote machines

For each machine `127.0.0.2` … `127.0.0.10`:
1. Copy `remote_sync.py` + configuration template.
2. Set a distinct `source_host` value (can be the machine IP or hostname).
3. Point `central_url` to the reachable IP of the central backend (e.g., `http://10.0.0.1:8080/api/v1/sync/events`).
4. Provide local DB credentials.
5. Start the sync script (cron/systemd or `--loop`).
6. On the central node, monitor `sync_errors` and application logs for failures.

## 5. Top issues & quick fixes

1. **ID conflicts / duplicates** – Ensure each remote keeps its own DB; the central table uses (`source_host`, `remote_id`) as a composite unique index, so duplicates are ignored automatically.
2. **Network timeout** – Adjust `timeout` and `batch_size` in the config or lower the cron frequency. The script retries next run without losing state.
3. **Bad token / auth error** – Verify `SYNC_TOKEN` on the central node matches `sync_token` in every remote config. Unauthorized pushes are logged in `sync_errors`.
4. **Large batches causing slow inserts** – Reduce `batch_size` (e.g., 50) and/or run the script more frequently. This keeps payloads small.
5. **Remote DB schema drift** – If remote tables lack new columns, JSON decoding may fail. Check `sync_errors.error_message`, update the remote schema, and re-run; rows remain unsynced until the issue is fixed.

With this PoC, the frontend continues to read from the central backend/database (`127.0.0.1`), while each remote node periodically ships only new rows using a lightweight script.
