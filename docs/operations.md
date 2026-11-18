# Logging, monitoring and recovery

## Request audit

`http.operation` JSON records include timestamp with zone, server-generated request ID, verified actor ID/role, HTTP method, route template, first UUID object parameter, socket peer, status, response bytes and duration. Anonymous requests use `anonymous`; successful login/registration attach the verified account. GET/POST/PATCH/DELETE plus route identify reads, creation, updates, deletion, assignment and role operations. Object IDs are present for routes with UUID parameters; list/create operations have no object parameter. Ticket history separately retains domain changes.

Status <400 is INFO, 4xx WARN, 5xx ERROR. Recoverer converts a panic to a generic 500; raw panic/error text and stacks are excluded. No passwords, bearer tokens, body fields, email addresses, query strings or raw unmatched paths are logged. Forwarded IP/identity headers are untrusted. Socket peer is the proxy address when deployed behind one; ingress must separately record trusted client attribution.

Compose limits API stdout logs to five 10 MB files. Restrict host log access to operators. Production policy: forward to a separate append-only collector, retain 30 days online and 90 days in restricted archive, then delete; preserve incident material under a separate retention decision. Rotation is configured locally; an external collector is a deployment requirement, not provisioned here.

## Basic monitoring

```sh
python3 scripts/ops/monitor.py /var/log/hnu-tickets/api.jsonl --ready http://127.0.0.1:8080/ready
```

Feed decoded application JSON lines (not Docker's outer JSON envelope). Run every minute. It examines the last five minutes, reporting request count, 5xx, 401/403/429, p95 latency, malformed lines and readiness. Alert thresholds: at least 10 denied/rate-limited requests, at least 5% 5xx with 20+ requests, p95 above 2 seconds, or readiness failure. Exit 1 signals an alert; connect this exit/result to the operator's notification system. Empty traffic is not by itself a failure. Baseline thresholds need tuning to real traffic. Also inspect backup job exit status, age of last successful off-host copy (over 4 hours is overdue), free disk space and archive capacity.

## Recovery inventory and policy

Back up the full PostgreSQL database: accounts and password hashes, categories, tickets, comments, history, persistent token revocations and encryption metadata. Preserve the corresponding signed release, migrations, deployment configuration and all still-required DATA_KEYS. Keep JWT and backup encryption keys in a separate access-controlled recovery vault. Source/releases are recoverable from the repository; binaries, caches and temporary files can be rebuilt. This application has no uploaded attachment store.

For the small project, use a full custom-format logical dump every four hours. Retain 42 four-hour copies (7 days), 30 daily and 12 monthly copies. Keep an encrypted local copy plus a copy in independent off-host storage with a separate account and retention lock; retain a monthly offline copy. API credentials must not delete backups. RPO target: 4 hours; RTO target: 1 hour including deployment, keys and functional checks. These are policy targets, not guarantees from the small drill. Schedule and remote storage require operator configuration; this repository implements the repeatable backup/verification/restore commands.

## Commands

Install PostgreSQL client tools matching the server major version and Python dependencies:

```sh
python3 -m pip install -r scripts/ops/requirements.txt
```

Set PGHOST, PGPORT, PGUSER and use a protected PGPASSFILE for database authentication. Load BACKUP_KEY from the recovery vault: it must be a Fernet key generated once with `Fernet.generate_key()` and retained for all copies encrypted by it. Keep DATA_KEYS separately; a database backup alone cannot decrypt application fields.

```sh
umask 077
python3 scripts/ops/backup.py backup /secure/backups/tickets-YYYYMMDD-HHMM.enc --database tickets
python3 scripts/ops/backup.py verify /secure/backups/tickets-YYYYMMDD-HHMM.enc
python3 scripts/ops/backup.py restore /secure/backups/tickets-YYYYMMDD-HHMM.enc --database tickets_recovered
```

Choose a unique timestamped filename. Existing files/databases are not overwritten. Only upload/mark a copy successful after both commands succeed, and verify the off-host copy after download. Save the returned SHA-256 in a protected inventory. Fernet authentication detects tampering; a hash stored beside an attacker-controlled file alone does not. Raw dumps exist temporarily in a private directory; use an encrypted temporary volume for sensitive production data. Encryption currently holds a dump in memory, appropriate for this small dataset; use a streaming backup system for large deployments.

Stop writes before cutover. Verify the decrypted archive with `pg_restore --list`, create a new database, and restore in one transaction. A failed restore may leave an empty new database; inspect it before retrying with another name. Restore the matching configuration and keys, switch POSTGRES_DB, restart the API, check `/ready`, login, role/object authorization, data counts and critical reads/writes. After a compromise, revoke old sessions/rotate signing credentials before reopening traffic; do not trust restored sessions automatically.

## Reproducible loss-and-restore exercise

Run manually under an unprivileged OS account with PostgreSQL 13+ server tools and free loopback ports 15488/18081:

```sh
go build -o bin/api ./cmd/api
go build -o bin/cryptodata ./cmd/cryptodata
PG_BIN=/usr/lib/postgresql/17/bin python3 scripts/ops/recovery_drill.py
```

The script creates a temporary cluster, synthetic account/category/ticket and ephemeral keys. It initializes schema and encryption, creates an encrypted backup, verifies it, rejects a corrupted copy, deletes only its own temporary source database, restores into a new database and compares the full API ticket response. It also verifies login/readiness and absence of secrets in request logs, then stops and removes the cluster. PostgreSQL 13+ provides `gen_random_uuid()` natively, so the drill omits the otherwise redundant pgcrypto extension declaration; application schema migrations are unchanged. The drill's trust authentication is restricted to its temporary loopback cluster. Do not expose its port.

`drill_results.json` and `drill.jsonl` record the run on 18 November 2026: encrypted backup 20,132 bytes; tampering rejected; source deleted; ticket unchanged; login/readiness successful; no password/token/email in the log. Measured recovery phase: 0.570 seconds for this tiny local dataset, excluding incident investigation and infrastructure provisioning. No synthetic changes occurred between backup and deletion, so observed loss was zero. The drill is a manual lab exercise, not a database service added to CI.
