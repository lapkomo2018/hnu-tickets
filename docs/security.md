# Security requirements

Reference implementation: `00c84f2c7daa020f50fd16f2857a240ffc7a92bd`, reviewed 2026-10-17.

[Security policy](security_policy.md) defines mandatory controls; [Privacy policy](privacy_policy.md) defines data handling; [Threat model](threat_model.md) identifies the relevant STRIDE scenarios. Requirements and acceptance criteria below are distinct from the observed implementation status.

## Assets and trust boundaries

| Asset | Protection goal |
| --- | --- |
| Account names/emails, ticket text and comments | Confidentiality: only permitted participants and administrators |
| Password hashes, JWT signing key and database credentials | Prevent disclosure and unauthorized account/session creation |
| Ticket status, category, priority and assignment | Integrity: only valid, authorized changes |
| Ticket history and operational logs | Traceability, restricted access and protection from alteration |
| PostgreSQL data and backups | Availability, confidentiality and recoverability |
| Source, migrations and release artifacts | Reviewed, reproducible changes |

Requests cross the untrusted client boundary before authentication, validation and authorization. The API accesses a separate PostgreSQL trust boundary using a service identity. Deployment configuration, signing secrets and backups are controlled by the operator. A signed JWT authenticates claims but does not make its role current; CORS does not authorize a caller.

## Threats and required controls

| Threat | Required control | Requirement |
| --- | --- | --- |
| Guessing a ticket UUID or changing a list filter | Object-level and list visibility enforcement, including comments/history | SEC-R02 |
| Privilege escalation through registration or stale credentials | USER-only registration, admin-only role changes, fresh authorization/revocation policy | SEC-R01–SEC-R03 |
| Forged or disclosed access token | Verify signature/algorithm/expiry, protect signing key, use HTTPS, limit lifetime | SEC-R01, SEC-R04 |
| Injection or malformed/oversized request | Parameterized SQL, identifier allowlists, strict bounded JSON and field validation | SEC-R05 |
| Partial workflow update, competing mutations or audit tampering | Transactions, row locking, restricted history permissions | SEC-R06 |
| Password guessing or expensive authentication abuse | Throttling, explicit password bounds and monitored authentication failures | SEC-R07 |
| Data loss or backup disclosure | Restricted encrypted backups and tested restoration | SEC-R08 |
| Dependency or unauthorized repository changes | Reviewed PRs, locked dependencies, CI and vulnerability checks | SEC-R09 |

## Security requirements and observed status

| ID | Required behavior | Source-observed status |
| --- | --- | --- |
| SEC-R01 | Store salted password hashes; verify bearer JWT signature, accepted algorithm and expiry; omit credentials from responses | bcrypt with default cost; issued HS256 tokens contain expiry; parser restricts HS256 and validates token; safe user DTOs. Explicit password upper bound is missing (G-05) |
| SEC-R02 | Enforce current role, ownership/assignment and valid workflow on every protected operation | Service policies and scoped listing exist; G-01/G-02/G-03 remain |
| SEC-R03 | Privileged role changes and logout must have a defined effect on existing sessions | Registration forces USER and role management is ADMIN-only; immediate revocation/current-role lookup is absent (G-01) |
| SEC-R04 | Use unique deployment secrets, HTTPS and verified TLS for remote PostgreSQL; do not expose the database publicly | Required secret settings and CORS allowlist exist. HTTPS termination, verified DB TLS, network restrictions and key rotation require operator configuration (G-07) |
| SEC-R05 | Reject invalid JSON, unknown fields and oversized bodies; validate input; parameterize values and allowlist SQL identifiers | JSON decoder caps bodies at 1 MiB, rejects extra JSON values/unknown fields; SQL values are parameterized and sorting is allowlisted |
| SEC-R06 | Commit ticket changes and history atomically; protect history from API and direct runtime SQL modification | Transactions and row locks exist; no history mutation API. Database-level immutability is not established (G-04) |
| SEC-R07 | Rate-limit authentication and protect resources from abuse; keep credentials out of logs | HTTP timeouts/body bounds and metadata logs exist; authentication throttling and explicit long-password validation are pending (G-05) |
| SEC-R08 | Retain restricted backups, restore them in isolation and document recovery results | Backup/restore procedure below is proposed; no successful restoration is asserted (G-08) |
| SEC-R09 | Review changes, run lint/tests/build and inspect dependencies for known vulnerabilities | CI covers lint, generated specification, test and build; dependency vulnerability scanning is not configured (G-09) |
| SEC-R10 | Minimize personal data, restrict its disclosure, enforce documented retention and controlled removal across current data/history/backups | Safe DTOs and ticket visibility exist; automated retention, account/ticket removal and restore-time removal replay are not implemented (G-10) |

## Implementation gaps

Priorities below are proposed remediation order, not measured severity scores. These items describe reviewed code paths; they have not been dynamically exploited in this documentation change.

| ID | Priority | Observation and evidence | Required follow-up / completion evidence |
| --- | --- | --- | --- |
| G-01 | High | [`auth/token.go`](../internal/auth/token.go) places the role in the token; [`auth/middleware.go`](../internal/auth/middleware.go) trusts it without loading current account role. [`users/service.go`](../internal/users/service.go) changes a role without invalidating issued tokens. Config defaults token lifetime to 24h. No revocation/logout endpoint exists | Define and implement session invalidation/current-role verification. Demonstrate an old privileged token loses permissions after demotion. Client token deletion alone cannot revoke a copied token |
| G-02 | High | [`tickets/service.go`](../internal/tickets/service.go) permits ADMIN to use `ChangeStatus`; [`tickets/domain.go`](../internal/tickets/domain.go) validates an edge but not an assignee. Thus NEW can become ASSIGNED without `Assign`. The schema does not enforce the assignee/status invariant | Restrict entry into ASSIGNED to assignment or check an existing valid assignee. Verify the generic status endpoint cannot create an unassigned ASSIGNED ticket |
| G-03 | Medium | Non-admin `tickets.Service.Update` checks creator ID and NEW status but not an explicit USER role. A former creator now acting as SPECIALIST can pass that condition | Require the intended current USER role for author edits, with a regression test alongside G-01 |
| G-04 | Medium | [`ticket_history` migration](../migrations/000005_create_ticket_history.up.sql) creates an ordinary mutable table with cascading ticket deletion. API append-only behavior does not establish database-level tamper resistance | Define separate runtime/migration roles; prevent runtime UPDATE/DELETE on history and uncontrolled destructive operations. Consider external protected audit storage for stronger threat models; verify actual privileges |
| G-05 | Medium | Router has no login/register throttling; [`auth/service.go`](../internal/auth/service.go) validates only a minimum password length, while bcrypt hashing has a byte limit | Add bounded throttling and explicit user-facing validation compatible with bcrypt's 72-byte input maximum; verify multibyte passwords and abuse/error behavior |
| G-06 | Functional | FR-06 requires a least-active-workload suggestion for confirmation. [`analytics/postgres.go`](../internal/analytics/postgres.go) reports workloads, but no recommendation service/route exists | Implement a deterministic recommendation with a tie rule and explicit admin confirmation, or formally revise FR-06. Do not describe analytics alone as the completed recommendation |
| G-07 | Deployment | [`main.go`](../cmd/api/main.go) starts plain HTTP; database configuration defaults `sslmode` to `disable`. Presence of a secret is checked, but secret strength/rotation is not enforced | Configure HTTPS, verified TLS for remote DB, private network access and independent random secrets; record deployment checks before network use |
| G-08 | Operational | Backup scheduling and a successful isolated restore are not established by the deployment configuration | Execute and record the isolated restoration procedure below; agree retention/frequency with the operator |
| G-09 | Process | Current CI has no dependency-vulnerability scan | Add a reproducible vulnerability scan and a process for reviewing/remediating findings; record its actual output |
| G-10 | Privacy | No account/ticket-delete API or retention job; history and backups may retain removed text | Implement controlled purge/anonymization respecting foreign keys, retention and recorded removal after restore; verify SEC-11 |

## Handling data, logs and secrets

Use synthetic ticket content for demonstrations. Support messages can contain personal information or credentials supplied accidentally; participants must avoid including secrets in tickets, comments and history. History old/new values may retain previous text, so redacting only the current ticket is insufficient.

Application request logs contain method, path, status, duration and request ID. Unexpected errors are logged internally, so operators must restrict log access and prevent sensitive values from being included in error context. Do not log request bodies, `Authorization`, passwords or connection secrets. A complete security-event audit of login failures and role changes is additional work; `ticket_history` covers ticket mutations only.

If a secret is exposed, replace it, invalidate affected sessions as applicable, review access logs and remove the exposed value from accessible artifacts. Removing a secret from the latest Git file does not remove earlier versions. JWT signing-key replacement invalidates existing tokens and must be coordinated with clients.

## Backup and recovery procedure

The operator must agree backup frequency, retention and acceptable data loss before deployment; no recovery-time guarantee is set for the educational environment.

1. Use a restricted backup identity and an authenticated PostgreSQL connection; supply credentials through protected configuration rather than committed scripts or command-line literals.
2. Create a consistent logical backup with `pg_dump` in custom format. Store an encrypted copy outside the database host with restricted access; record time, schema/migration version and backup identifier without secrets.
3. Restore using `pg_restore` into a new isolated database. Keep it disconnected from normal clients; do not overwrite the working database during a verification exercise.
4. Verify schema/migrations, account/ticket/comment/history counts, foreign-key consistency and a sample of authorized ticket reads. Record expected/actual results and restoration duration.
5. For a real incident, stop writes, preserve incident evidence, select a verified backup, restore into a clean database and point the service at it after checks. Rotate compromised secrets and record any data-loss interval before reopening access.

Retention and removal rules are defined in [Privacy policy](privacy_policy.md#retention-schedule). After restoring a backup, reapply recorded removals before reopening access.

## Verification scenarios

These are acceptance scenarios, not test execution results. Run on isolated synthetic data and record the commit/environment and actual outcome.

| Scenario | Steps and expected result |
| --- | --- |
| SEC-01 Authentication | Missing, altered and expired tokens return 401; invalid credentials fail; registration cannot choose ADMIN; returned profiles contain no password hash |
| SEC-02 Object access | USER A requests USER B's ticket/comments/history, and an unassigned specialist does likewise: all fail without disclosure. List/search/filter results stay scoped to the actor |
| SEC-03 Role freshness | Issue an ADMIN token, demote that account, then repeat an admin operation with the old token. Required result: denied. Currently expected to reveal G-01; do not mark passed |
| SEC-04 Workflow integrity | Attempt NEW→ASSIGNED through the generic status route without an assignee, invalid terminal transitions, and assignment to USER. All must fail without partial writes; first case is G-02 |
| SEC-05 Author edit | Create as USER, change that creator to SPECIALIST, obtain a fresh token and edit the still-NEW ticket. Required result: denied; current non-admin ownership check is G-03 |
| SEC-06 Atomicity/concurrency | Fail the history write after a ticket mutation and verify full rollback; race conflicting status/assignment changes and confirm only valid serialized outcomes persist |
| SEC-07 Input and abuse | Exercise malformed/extra JSON, unknown fields, oversized bodies, invalid filters, long and multibyte passwords, repeated login attempts. No internal errors/secrets leak; missing abuse controls remain G-05 |
| SEC-08 Audit privileges | Confirm no history edit/delete API exists; try UPDATE/DELETE using the deployment runtime database identity. Required database denial must be verified after G-04 |
| SEC-09 Recovery | Restart after committed writes, then restore a backup in isolation and compare rows/relationships and authorized reads; record results rather than assuming success |
| SEC-10 Deployment | Verify HTTPS and DB certificate checking, restricted DB/log/backup access, secret handling and vulnerability-scan output under the actual deployment configuration |
| SEC-11 Privacy lifecycle | Confirm DTOs/logs omit secrets, select records older than the configured retention period, verify restricted purge/anonymization and foreign-key consistency, then restore an older copy and reapply recorded removals before allowing access |

## Acceptance and threat traceability

| Requirement | Threats | Acceptance evidence |
| --- | --- | --- |
| SEC-R01 | TM-01, TM-22 | SEC-01/SEC-07: invalid credentials, signatures and expired tokens fail; hashes/secrets are absent from responses |
| SEC-R02 | TM-05–TM-07 | SEC-02/SEC-04/SEC-05: role/object/workflow violations leave data unchanged |
| SEC-R03 | TM-01, TM-06 | SEC-03: old privileged sessions stop working after demotion; copied tokens follow the defined invalidation policy |
| SEC-R04 | TM-03, TM-11, TM-17, TM-19, TM-20, TM-22–TM-24 | SEC-10: protected channels, credential access and network restrictions are verified |
| SEC-R05 | TM-07, TM-14 | SEC-07: malformed/boundary input is rejected without internal data disclosure |
| SEC-R06 | TM-02, TM-04, TM-05, TM-08, TM-10, TM-13, TM-21, TM-23, TM-25 | SEC-06/SEC-08: rollback, privileged-operation evidence and denied runtime history modification |
| SEC-R07 | TM-09, TM-12, TM-15 | SEC-07/SEC-10: bounded authentication/resource behavior and operational limits |
| SEC-R08 | TM-04, TM-12, TM-16–TM-18, TM-23, TM-24 | SEC-09: isolated recovery with schema, relationship and data checks |
| SEC-R09 | Development process; TM-03, TM-22 | SEC-10: reviewed build configuration and recorded dependency scan |
| SEC-R10 | TM-07, TM-11, TM-14–TM-17, TM-21, TM-25 | SEC-11: data minimization, retention, removal and removal replay after restore |

Record the tested commit, configuration, dataset, expected and actual outcome. Close implementation gaps only after the corresponding control and regression checks are complete.
