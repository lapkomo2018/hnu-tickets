# Security policy

| Field | Value |
| --- | --- |
| Version | 1.0.0 |
| Updated | 2026-10-17 |
| Owner | HNU Tickets repository maintainer |
| Scope | Server source, configuration, database migrations, dependencies, build workflow, deployment and support data |

## Purpose

Protect account credentials, support records, workflow integrity and service availability throughout development and operation. This policy defines mandatory practices; [Security requirements](security.md) records implementation status and verification criteria. The [threat model](threat_model.md) identifies the scenarios these controls address.

## Scope

The policy applies to the Go API, PostgreSQL, operational logs, configuration, backups, GitHub repository and CI workflow. It covers developers, reviewers and operators. Browser clients must respect the API contract, protect access tokens and render user content safely; their internal implementation is maintained separately.

One deployment serves one organization. File uploads, payment processing, external mail delivery and shared multi-tenancy are outside the current product scope.

## Roles and responsibilities

| Responsibility | Owner |
| --- | --- |
| Maintain policy, accept time-limited exceptions and set remediation priority | Repository maintainer |
| Implement validation, authorization and regression coverage; update relevant documentation | Developer |
| Review security-sensitive changes and required verification evidence | Reviewer |
| Manage runtime identities, secrets, network access, logs and verified backups | Operator |
| Limit ticket access and avoid putting credentials into support messages | Application administrator and participants |

A small team may combine these responsibilities. Privileged changes still require an explicit review record. Application ADMIN rights do not confer repository or database administration rights.

## Security principles

- Deny operations unless a server-side rule explicitly permits the actor, resource and action.
- Grant the minimum application, repository and database privileges needed for the task.
- Treat request fields, identifiers, filters and client-provided metadata as untrusted.
- Protect data at each boundary: client/API, API/database and operator/storage.
- Fail closed when authentication, configuration or authorization cannot be established.
- Keep credentials outside source code, generated specifications, logs, images and release artifacts.
- Minimize retained personal information and exposure through responses, history and operational tools.

## Development environment

Use individual GitHub identities with multi-factor authentication and device-specific SSH keys or narrowly scoped access credentials. Remove access when it is no longer needed. Never share a personal signing key or a privileged account between developers.

Changes to `main` go through reviewed branches. Configure branch protection and required checks where available. Changes to authentication, authorization, database privileges, migrations, secrets and CI permissions require security-focused review.

Pin Go dependencies and build tools in module files. Review origin, necessity, licence and known vulnerabilities before introducing or updating a dependency. CI must use the least permissions needed; the existing workflow requests `contents: read`. External contributions must not receive deployment secrets.

Keep development, test and network deployment configurations separate. Use synthetic accounts and messages outside the operational environment. Do not copy a working database or its backups onto an unprotected developer machine.

## Source and artifact protection

Build from a reviewed commit with recorded dependency versions. Identify each release by its source commit and preserve enough information to reproduce it. Restrict write access to build configuration and deployment settings.

Exclude `.env`, private keys, database exports and tokens from Git. Example configuration and seeded accounts are development material only. If a secret appears in Git history or logs, rotate it and inspect exposure; deleting the latest file does not revoke the secret.

The API runtime identity must not own the schema or have unrestricted DDL rights. Use a separate migration identity. Protect the ticket history table from runtime UPDATE/DELETE and prevent uncontrolled destructive operations on referenced tickets.

## Secure implementation

### Authentication and sessions

Hash passwords with bcrypt using the maintained library. Validate password input against both the minimum character requirement and the library's byte limit. Do not silently truncate passwords or return password hashes.

Verify the token signature, allowed algorithm and expiration. Derive the actor from verified authentication, never from a request body. Network deployments must invalidate affected sessions or revalidate current account permissions after a role change. Deleting a client token alone is not revocation.

Use independent random JWT signing secrets for each environment. Restrict their readers, record rotation responsibility and replace them after suspected compromise. Authentication responses must not expose internal failures. Apply bounded throttling to login and registration.

### Authorization and workflow

Check role and object access for every protected read or mutation, including comments, history and analytics. Listing restrictions must remain effective when filters are supplied.

Require a valid SPECIALIST assignment before entering ASSIGNED. Enforce the current USER role, ownership and NEW status for author edits. ADMIN remains bound by the ticket lifecycle. Never use CORS or hidden UI controls as authorization.

### Input and storage

Decode one bounded JSON value, reject unknown fields and validate text lengths, UUIDs, enums and references. Parameterize SQL values and allowlist dynamic identifiers. HTML clients must render ticket text as text or apply context-appropriate escaping.

Write each ticket mutation and its history atomically. Lock the ticket before checking and applying a competing workflow change. Preserve database constraints and test rollback on failed audit writes.

### Transport and configuration

Use HTTPS for browser/integration traffic. For a remote PostgreSQL connection, use TLS with certificate and hostname verification. Restrict database ports to approved hosts. Local Compose defaults are not network-deployment controls.

Validate required configuration on startup. Do not introduce a fallback production secret or proceed with missing credentials. Keep timeouts and body-size limits; add operation limits where unbounded work can exhaust resources.

### Errors and logging

Return stable public error codes and safe messages. Keep stack traces and storage details in restricted operational logs.

Log security-relevant failures and privileged changes with time, actor when authenticated, action, target and result. Treat client request IDs as correlation metadata, not proof of identity. Do not record passwords, bearer tokens, signing keys, connection strings or full request bodies. Ticket history is a business audit record and does not replace authentication/administration event logging.

## Privacy and recovery

Apply the data categories, access restrictions and retention schedule in [Privacy policy](privacy_policy.md). History and backups can retain old personal text; removal must account for those copies.

Create encrypted backups outside the database host and verify restoration into an isolated database. Record backup age, restored schema, row/reference checks and the outcome. Restrict restored environments as strictly as the original data. Agree acceptable data loss and recovery time with the operator before deployment.

## Security verification

| Change or release condition | Required evidence |
| --- | --- |
| Every code change | Review, lint, relevant tests and build |
| Authentication or authorization change | Negative tests for invalid tokens, changed roles, foreign tickets and unassigned specialists |
| Workflow/storage change | Valid/forbidden lifecycle tests, rollback and competing-mutation checks |
| API input/response change | Malformed, unknown, oversized and boundary inputs; safe response inspection |
| Dependency/build change | Dependency review and vulnerability scan, such as `govulncheck` |
| Network deployment | HTTPS/DB TLS checks, secret handling, runtime privileges and restricted database access |
| Recovery readiness | Successful isolated restore with documented checks |

The acceptance scenarios SEC-01–SEC-11 and implementation gaps G-01–G-10 are in [Security requirements](security.md). A network release must not leave a known authorization bypass, exposed secret or unverified critical protection unresolved. Any exception must explicitly identify the remaining exposure and compensating controls.

## Vulnerability management

Record the affected component and version, prerequisites, impact, reproducible scenario, priority, owner and required fix. Keep secrets, real personal data and reusable credentials out of public reports; use a restricted reporting channel when details would expose operational data.

Contain active exposure first, then correct the cause, add regression coverage and repeat the relevant checks. Close a finding only after checking the fix. Review related flows and update the model if the failure reveals a wider design issue.

Prioritize remotely usable authorization/credential failures ahead of lower-impact weaknesses. Record a remediation deadline with the owner; do not imply that a source-review observation is a demonstrated exploitation.

## Exceptions and maintenance

An exception must identify the requirement, affected environment, reason, risk owner, compensating measure and expiry date. The maintainer reviews it; an expired exception does not authorize continued deployment.

Review this policy after changes to authentication, architecture, data categories, dependencies, deployment or an incident. Version substantive changes and keep the threat model, privacy policy and security requirements consistent.

## Related documents

- [Project policy](project_policy.md)
- [Architecture](architecture.md)
- [Business logic](business_logic.md)
- [Privacy policy](privacy_policy.md)
- [Threat model](threat_model.md)
- [Security requirements and implementation gaps](security.md)
