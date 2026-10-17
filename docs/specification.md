# System specification

Version 1.0.0 · 2026-10-17 · Documentation baseline for Data and Information Security, laboratory No. 1.

## Purpose and scope

HNU Tickets centralizes technical support requests, responsibility, communication and performance information for one organization. The system replaces disconnected messages with a controlled ticket register. Users submit problems, specialists process assigned requests, and administrators distribute work and monitor execution.

This repository supplies the Go REST API and PostgreSQL persistence. A separate browser client consumes the API. The product baseline includes a Ukrainian browser interface, but UI implementation and client-side logout are outside this server repository. Shared multi-tenancy, billing, attachments, email delivery, offline mode and reopening terminal tickets are outside the current scope.

The specification retains FR-01–FR-12 from the [Software Construction technical specification](https://docs.google.com/document/d/17-yJyBPrhCOHj3Ee_OQHcGjJoC46sjPGtDXMujjBZ0U/edit). The table below states intended behavior. Source-review notes identify partial implementation without changing the baseline requirements.

## Stakeholders and roles

| Stakeholder | Need |
| --- | --- |
| Requester (`USER`) | Register a problem, track own requests, provide information and confirm resolution |
| Support specialist (`SPECIALIST`) | See assigned work, request details and record resolution |
| Service administrator (`ADMIN`) | Allocate responsibility, manage accounts/categories and inspect analytics |
| Operator | Configure deployment, protect credentials, monitor availability and recover data |
| Maintainer | Preserve consistent rules, API compatibility and repeatable verification |

Permissions are specified in [Business logic](business_logic.md#access-policy). Operators are a deployment responsibility, not a fourth application role.

## Functional requirements

| ID | Required behavior | Acceptance criterion |
| --- | --- | --- |
| FR-01 | Register with name, unique email and password; assign only `USER` | A valid registration creates a USER; invalid fields and duplicate email fail; a supplied privileged role cannot create an ADMIN |
| FR-02 | Authenticate by email/password, provide a session and logout, and reject invalid or expired credentials | Login returns a token and safe profile; invalid login or altered/expired token fails; client logout removes its token. Server-side revocation is additional security work (G-01) |
| FR-03 | Only USER can create a ticket with title, description, category and priority | Ticket starts as `NEW`, identifies the authenticated author and has a creation-history entry; failure leaves no partial ticket/history |
| FR-04 | Provide paginated lists and cards with role-based visibility | USER sees own, SPECIALIST assigned, ADMIN all; guessing an inaccessible UUID does not disclose ticket data |
| FR-05 | Permit author edits of own NEW tickets and administrator edits of permitted fields | Validate title, description, category and priority; reject ownership/status violations. Explicit current-role checking needs G-03 |
| FR-06 | ADMIN assigns/reassigns a SPECIALIST to a nonterminal ticket; first assignment changes NEW to ASSIGNED; suggest the least-loaded specialist for administrator confirmation | Reject a non-specialist or terminal ticket; save assignment/history atomically. Workload analytics exists, but a recommendation operation is not implemented (G-06); assignment invariant needs G-02 |
| FR-07 | Enforce the defined lifecycle and actor permissions | All listed transitions succeed only for permitted actors; unlisted transitions leave data unchanged; CLOSED/CANCELLED cannot reopen |
| FR-08 | Participants with ticket access can add/view comments | Nonempty comment of at most 5000 characters is attributed to the actor; unauthorized access fails |
| FR-09 | Search visible tickets by title or identifier | Title search and UUID lookup return only visible tickets |
| FR-10 | Filter visible tickets by status, category, priority and creation date; provide ADMIN an assignee filter | Filters combine with mandatory visibility restrictions; a specialist cannot replace their visibility restriction with another assignee |
| FR-11 | ADMIN views counts, status/priority/category distributions, specialist workload and average resolution time for a period | Values match an independently calculated dataset; non-admin requests fail |
| FR-12 | Record major ticket changes with actor, time, action and previous/new values; participants can read history but cannot edit it | Ticket mutations and history are atomic; no history-edit/delete API exists. Stronger storage protection is G-04 |

FR-06 recommendation and the server-side part of FR-02 logout are not claimed complete. [Security](security.md#implementation-gaps) records the source-observed gaps; runtime acceptance is still required.

## Data and interface constraints

- Name: 2–100 characters; syntactically valid, normalized unique email; password: at least 8 characters. A clear maximum compatible with bcrypt remains part of G-05.
- Ticket title: 3–255 characters after trimming; description: nonempty; category: an existing UUID; priority: `LOW`, `MEDIUM`, `HIGH` or `CRITICAL`.
- Category name: 2–100 characters and unique; referenced categories cannot be deleted. Comments: 1–5000 characters after trimming.
- Statuses and transitions are defined in [Business logic](business_logic.md#status-lifecycle). UUIDs identify resources; date/time values use RFC3339 and lifecycle times are stored in UTC.
- REST routes use `/api/v1`; JSON bodies accept one value, reject unknown fields and have a 1 MiB decoding limit. Error bodies contain `error.code` and `error.message`.
- Ticket pages default to 20 items and are capped at 100. Detailed routes and request/response schemas are in [API](api.md) and [OpenAPI](swagger.yaml).

## Nonfunctional requirements

| ID | Requirement | Verification |
| --- | --- | --- |
| NFR-01 | Server-side least-privilege access and protected credentials | Run SEC-01–SEC-05 from [Security](security.md#verification-scenarios) |
| NFR-02 | Each ticket mutation and audit write is atomic; concurrent mutations preserve the lifecycle | Inject a history-write failure; verify rollback, then exercise conflicting operations in PostgreSQL (SEC-06) |
| NFR-03 | Committed records survive application restart; recovery is possible from a verified backup | Restart and inspect committed data; restore an isolated database and compare counts and relationships (SEC-09) |
| NFR-04 | Main ticket operations complete within 2 seconds in the agreed acceptance environment | Before measurement record commit, CPU/RAM, data volume, concurrency, operation mix and timing method. Report actual results; capacity is not yet established |
| NFR-05 | Failures return consistent errors without internal details; health/readiness expose basic availability | Exercise malformed input, storage failure, `/health` and `/ready`; inspect responses/logs |
| NFR-06 | Network deployment uses HTTPS; remote DB connections use verified TLS | Inspect deployment configuration and verify clients reject invalid certificates; not guaranteed by local Compose defaults |
| NFR-07 | Changes remain reproducible and maintainable | Pinned modules, migrations, reviewed changes, CI and synchronized documentation as defined in [Policy](project_policy.md) |

Long-period reports require separate performance measurement. No measured availability percentage, guaranteed recovery time or demonstrated concurrency capacity is asserted.

## Environment and architecture constraints

Use the toolchain declared in [`go.mod`](../go.mod) (currently Go 1.27), PostgreSQL 17 or newer as stated in the repository setup, and Docker/Compose for the container workflow. A 2-core, 4-GB host is the inherited demonstration baseline, subject to test confirmation rather than a capacity guarantee. Persistent storage must account for tickets, comments, history and separate backups.

The system is a modular monolith with one database; see [Architecture](architecture.md). Database constraints complement application validation. The service does not implement an organization/tenant isolation model and must not be described as a shared multi-tenant service.

## Acceptance and traceability

Acceptance records must identify the tested commit, environment, dataset, scenario, expected/actual outcome and unresolved defects. Execute both successful and denied operations for all three roles, every permitted/forbidden lifecycle edge, rollback/restart checks, analytics on a known dataset and the defined security scenarios. A source review alone does not satisfy runtime acceptance.

| Requirement area | Implementation reference | Verification basis |
| --- | --- | --- |
| FR-01–FR-02 | `internal/auth`, `internal/users` | Password/token tests, HTTP scenarios, SEC-01/SEC-03; logout gap G-01 |
| FR-03–FR-07 | `internal/tickets`, transaction adapter | Domain/policy tests, workflow and rollback scenarios; G-02/G-03/G-06 |
| FR-08 | `internal/comments`, ticket access checker | Comment validation and cross-ticket access scenarios |
| FR-09–FR-10 | Ticket filters and PostgreSQL adapter | Combined filters, pagination and visibility scenarios |
| FR-11 | `internal/analytics` | Independent counts/time calculation, admin-only checks |
| FR-12 | `internal/history`, ticket transactions | Read-only API checks, failure injection and database privilege review |

Source-observed implementation is recorded as of the baseline commit in the [documentation index](README.md). Future fixes must update this assessment and preserve the FR identifiers.
