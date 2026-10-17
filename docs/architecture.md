# Architecture

## Context and architectural concerns

HNU Tickets runs as one Go process and one PostgreSQL database for one organization. A separate web client and authorized integration clients access the REST API. The server owns authentication, access decisions, ticket processing and persistence.

| Stakeholder | Concern | Architectural response |
| --- | --- | --- |
| Requester/specialist | Correct visibility and workflow | Service-level role/object checks and ticket lifecycle policy |
| Administrator | Assignment, traceability and analytics | Ticket transactions/history and admin-only relational reports |
| Maintainer | Changes remain understandable and testable | Feature modules, repository interfaces, explicit wiring and tests |
| Operator | Availability, confidential configuration and recovery | Health/readiness routes, timeouts, external secrets and persistent PostgreSQL; deployment protection and backup verification remain operator tasks |

Scope, acceptance and constraints are in [Specification](specification.md); protection requirements and limitations are in [Security](security.md).

## Modular monolith and components

Code is grouped by feature instead of horizontal controller/service/repository folders. This keeps deployment simple while preserving responsibilities. Requests pass through authentication middleware into a feature handler, then a service/policy and repository interface before the PostgreSQL adapter performs storage operations.

| Component | Responsibility and dependencies |
| --- | --- |
| `cmd/api` | Composition root: configuration, database pool, handlers/services, routes and HTTP server |
| `internal/auth` | Credentials, bcrypt, JWT and bootstrap; uses the users repository |
| `internal/users` | Profiles and administrator-controlled roles |
| `internal/categories` | Category directory and administrator-controlled changes |
| `internal/tickets` | Domain types, lifecycle, visibility, assignment, application service and PostgreSQL adapter; depends on users, history and a transaction interface |
| `internal/comments` | Discussion; uses a narrow ticket access-checker interface |
| `internal/history` | Audit persistence and read service; uses a ticket access-checker interface |
| `internal/analytics` | Admin-only current-state aggregates from relational data |
| `internal/platform` | Configuration, database/transactions, HTTP errors, principal context, CORS, logging and API generation |
| `migrations`, `docs` | Versioned database schema and maintained/generated documentation |

The module and invariant map is in [Business logic](business_logic.md).

## Layers and dependencies

- Domain code defines tickets, roles, priorities, statuses and lifecycle rules without HTTP or SQL operations.
- Application services authorize actors and coordinate repository interfaces. They contain orchestration and some validation/business decisions; this is not a separate class-per-command CQRS architecture.
- HTTP handlers decode and validate transport data. Feature HTTP adapters map expected errors to public status codes.
- PostgreSQL adapters implement repositories with parameterized SQL through `pgx`. SQL identifiers for sorting/grouping come from controlled choices.
- The composition root constructs dependencies explicitly; no framework container is required.

An existing coupling is that services and ticket policies use `Principal` from `internal/platform/middleware`. The type is shared without requiring an HTTP request, but its package location couples the application to a transport-named package. Moving the actor type to a neutral package is a possible later cleanup.

## Error boundaries

PostgreSQL adapters translate expected database conditions into feature errors and wrap unexpected storage failures with operation context. Services return those errors and create feature errors for business validation. Feature HTTP adapters assign public statuses/codes; shared error middleware writes JSON responses, logs unexpected failures and hides their internal details. Panic recovery returns a generic internal error and keeps stack details in operator-accessible logs.

## Transaction and concurrency view

Ticket creation, editing, assignment and status changes write the ticket and history in one transaction. Repository adapters obtain the transaction executor from context. Mutable ticket rows are selected with `FOR UPDATE` before workflow changes; checks and updates operate on that locked state. A failed history write rolls back the corresponding ticket mutation.

After committing a mutation, the service reads the response record separately. Another committed mutation may therefore be reflected in that response; it is not a frozen snapshot of the previous write. Account role changes are separate operations and are not synchronized with token claims. Dynamic transaction guarantees require the integration checks in the specification.

## Authentication and authorization

JWT authentication establishes a principal with a user UUID and token role. Issued tokens contain `sub`, `jti`, `role`, issue time and expiry. Middleware requires a token UUID and expiry, checks persistent revocation state and loads the current role from PostgreSQL before creating the request principal.

Authorization remains in services and policies:

- USER sees/comments on tickets they created.
- SPECIALIST sees/comments on assigned tickets.
- ADMIN sees all tickets and manages users, categories, assignments and analytics.

Comments and history call the ticket access checker. The full intended status transition matrix is in [Business logic](business_logic.md#status-lifecycle). CLOSED/CANCELLED are terminal, and entering ASSIGNED should require a valid assigned specialist; the generic status route currently has G-02.

## Persistence and audit view

PostgreSQL contains users, categories, tickets, comments and ticket history; see [Database](database.md). UUID foreign keys, uniqueness constraints and allowed-value checks complement application validation. This schema does not implement row-level tenant isolation.

History is append-only through the API, not immutable against privileged database access. It is neither event sourcing nor a message bus. Analytics reads the current relational model directly. Stronger history privilege protection is G-04.

## Deployment and operational view

The development Compose workflow runs the API, PostgreSQL and migration step; PostgreSQL data requires persistent storage. Local setup and required configuration are in the root [README](../README.md). The HTTP process uses header/read/write/idle timeouts, graceful shutdown and health/readiness routes.

For network deployment, put HTTPS termination at a reverse proxy or equivalent trusted ingress, restrict direct API/database exposure and configure verified TLS for a remote database. These are deployment requirements, not controls provisioned by the current local setup. Keep live credentials outside Git and backups separately protected. Proposed recovery steps and outstanding verification are in [Security](security.md#backup-and-recovery-procedure).

## Decisions and limitations

| Decision | Reason and consequence |
| --- | --- |
| One process and one relational database | Suitable for the educational scope; avoids distributed coordination, but deployment scales as one service |
| Feature grouping with repository interfaces | Keeps related behavior together and allows isolated tests; does not require a full DDD framework |
| Transactional audit writes and ticket row locks | Preserve mutation/history consistency and serialize competing ticket workflows |
| JWT identity with current database role and revocation checks | Signed authentication plus current permissions; logout invalidation persists across API restarts |
| Relational aggregates for analytics | No extra reporting pipeline; large-period report performance must be measured |

The [threat model](threat_model.md) defines data flows and trust boundaries for this architecture. [Security requirements](security.md) records controls and remaining implementation gaps.

