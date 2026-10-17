# HNU Tickets business logic

This document defines the intended business rules of the existing product. Source-observed deviations are explicitly listed in [Security](security.md#implementation-gaps). The [specification](specification.md) retains the original FR-01–FR-12 requirement identifiers.

## Purpose and boundaries

The system owns support tickets, their assignment, processing, discussion and change history within one organization. The central consistency boundary is a ticket mutation together with its history entries.

| Module | Owns | Dependencies and boundaries |
| --- | --- | --- |
| Auth | Registration, credential verification, JWT issuance and initial-admin bootstrap | Uses the users repository; registration cannot choose a privileged role |
| Users | Account identity, profile and role administration | Admin-only management; credentials are not public response fields |
| Categories | Ticket category directory | ADMIN manages; authenticated roles list; referenced categories cannot be deleted |
| Tickets | Ticket fields, visibility, assignment and lifecycle | Coordinates user lookup, ticket storage, transaction and history interfaces |
| Comments | Ticket discussion | Uses the ticket access checker; posting a comment does not itself change status |
| History | Append-only ticket change records exposed for reading | Shares ticket visibility; no update/delete API |
| Analytics | Counts, distributions, workload and resolution duration | ADMIN only; reads the relational model, not an event stream |

Browser rendering, token storage and client logout belong to the separate client. Email notifications, files, shared multi-tenancy, automated ticket assignment and background queues are outside this implementation. The FR-06 suggestion is a pending confirmed-choice aid, not automatic assignment.

## Entities and invariants

| Entity | Main fields | Invariants |
| --- | --- | --- |
| User | UUID, name, email, password hash, role, created/updated times | Stable ID; unique normalized email; role is USER/SPECIALIST/ADMIN; registration assigns USER; passwords stored as hashes |
| Category | UUID, name, description | Stable ID; unique valid name; a category referenced by a ticket cannot be deleted |
| Ticket | UUID, title, description, status, priority, category ID, creator ID, optional assignee ID, created/updated/resolved/closed times | Starts NEW; creator is the authenticated USER; valid category/priority; status follows the lifecycle; assignment selects a SPECIALIST; terminal tickets cannot be assigned or reopened |
| Comment | UUID, ticket ID, author ID, content, creation time | Existing accessible ticket; actor is the author; trimmed content has 1–5000 characters; no edit/delete operation in current API |
| History entry | UUID, ticket ID, actor ID, action, old/new values, creation time | Written with the ticket mutation; actor derives from authentication; entries are read-only through the API |

UUID identity and creation times are not client-editable fields. Optional assignee and lifecycle times use pointers/SQL NULL. `Status`, `Priority` and `Role` are named Go types with allowed values; no separate VO inheritance hierarchy exists. API responses use DTOs that omit password hashes.

## Access policy

| Operation | USER | SPECIALIST | ADMIN |
| --- | --- | --- | --- |
| Create ticket | Yes | No | No |
| View/list ticket, comments and history | Own tickets | Assigned tickets | All tickets |
| Edit title, description, category, priority | Own NEW ticket | No under intended role policy | Any ticket's permitted fields |
| Assign/reassign specialist | No | No | Nonterminal tickets |
| Add comment | Own ticket | Assigned ticket | Any ticket |
| Manage users/roles/categories | No | No | Yes |
| List categories / read own profile | Yes | Yes | Yes |
| View analytics | No | No | Yes |

Status changes additionally follow the matrix below. Listing always adds the actor's visibility restriction on the server. An incoming filter cannot widen it. The API has no ticket-delete operation. Field edits and comments are not currently restricted solely because a ticket is terminal; only lifecycle changes and reassignment are blocked by terminal status.

The intended edit policy requires the author's current role to be USER; the existing service only checks ownership and NEW for a non-admin (G-03). Current JWT roles may also be stale (G-01).

## Status lifecycle

The complete allowed transition matrix is below. Every omitted transition, including a transition to the same status, is forbidden. An ADMIN can perform only listed transitions and must preserve assignment invariants.

| From | To | Permitted actor / condition |
| --- | --- | --- |
| NEW | ASSIGNED | ADMIN through assignment to an existing SPECIALIST |
| NEW | CANCELLED | Ticket author with USER role, or ADMIN |
| ASSIGNED | IN_PROGRESS | Assigned SPECIALIST, or ADMIN |
| ASSIGNED | CANCELLED | Ticket author with USER role, or ADMIN |
| IN_PROGRESS | WAITING_FOR_USER | Assigned SPECIALIST, or ADMIN |
| IN_PROGRESS | RESOLVED | Assigned SPECIALIST, or ADMIN |
| IN_PROGRESS | CANCELLED | Ticket author with USER role, or ADMIN |
| WAITING_FOR_USER | IN_PROGRESS | Ticket author with USER role, assigned SPECIALIST, or ADMIN |
| RESOLVED | IN_PROGRESS | Assigned SPECIALIST, or ADMIN |
| RESOLVED | CLOSED | Ticket author with USER role, or ADMIN |
| CLOSED / CANCELLED | — | Terminal; no outgoing transitions |

`NEW → ASSIGNED` must use the assignment operation so that status and responsible person agree. The generic status endpoint currently omits the assignee precondition for ADMIN (G-02).

Entering RESOLVED sets `resolved_at`. Returning from RESOLVED to IN_PROGRESS clears it. Entering CLOSED sets `closed_at`. These times and `updated_at` use UTC. Adding a comment while WAITING_FOR_USER does not resume the ticket automatically: an explicit permitted status command is required.

## Assignment and processing flow

1. USER creates a validated NEW ticket. The service writes ticket and creation history in one transaction.
2. ADMIN reviews tickets and specialist workload, selects a SPECIALIST and confirms assignment. First assignment changes NEW to ASSIGNED. The FR-06 least-workload suggestion is not implemented in the API (G-06).
3. The assigned SPECIALIST starts work, requests information or records resolution through explicit status changes. Participants exchange comments under the same ticket visibility policy.
4. The author can resume a request waiting for information and close a resolved request. A specialist can return a resolved request to work. Cancellation is available only along the matrix edges.
5. Every ticket mutation is committed with its history entries; readers see the committed state. ADMIN can inspect aggregate analytics.

Reassignment is permitted for any nonterminal ticket and normally preserves its status; NEW becomes ASSIGNED. Old/new assignee IDs are recorded. Reassignment to the same ID is not explicitly rejected and may produce another history entry. A later user-role change does not automatically reassign that person's tickets; operational reassignment and authorization freshness require attention (G-01).

## Validation, commands and queries

Input bounds are in [Specification](specification.md#data-and-interface-constraints). Service validation enforces text, role, enum and workflow rules; foreign keys reject nonexistent referenced users/categories. Parameterized queries and identifier allowlists protect query construction.

| Area | Commands | Queries |
| --- | --- | --- |
| Accounts | Register, login, bootstrap administrator, update user/profile role | Own profile; ADMIN user list/details |
| Categories | Create, update, delete unused category | List categories |
| Tickets | Create, update fields, assign/reassign, change status | Get accessible ticket; list/search/filter/page |
| Discussion | Add comment | List comments for an accessible ticket |
| History | Internal write within a ticket transaction | List history for an accessible ticket |
| Analytics | None | Overview, status/priority/category distributions, specialist workload |

These are service operations, not separate CQRS command/query classes. Account updates and category changes are not currently written to `ticket_history`.

## History and consistency

The service records creation, field updates, category/priority changes, assignment/reassignment and lifecycle actions. Status transitions contain old/new statuses; assignment contains old/new assignee IDs. Initial assignment records an assignment action while also changing NEW to ASSIGNED. Creation does not contain a full old/new snapshot. Generic field-update entries contain old/new values without a separate field-name column.

The transactor encloses a ticket write and its history writes. Workflow mutations read the ticket `FOR UPDATE`; a failed validation or write aborts the transaction. History is not an event bus, an event-sourced model or a cryptographically immutable ledger. Direct database privileges can bypass API restrictions; G-04 describes the remaining protection work.

## Analytics definitions

- Period filters include tickets whose `created_at` is within the inclusive `from`/`to` interval, using RFC3339 boundaries. They do not represent a historical snapshot of status at that time.
- Open tickets exclude RESOLVED, CLOSED and CANCELLED. Counts and distributions use current ticket state.
- Active specialist workload counts assigned tickets in ASSIGNED, IN_PROGRESS and WAITING_FOR_USER, grouped for users currently having SPECIALIST role.
- Average resolution time is the mean number of seconds between `created_at` and non-null `resolved_at` for selected tickets. With no such tickets the value is zero. Reopening clears the timestamp, so the previous resolution no longer contributes until resolution is recorded again.
- The current data model is read directly. A least-loaded recommendation and automatic assignment are not implemented.
