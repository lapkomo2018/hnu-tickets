# Privacy policy

| Field | Value |
| --- | --- |
| Version | 1.0.0 |
| Updated | 2026-10-17 |
| Responsible role | Operator of the organization using HNU Tickets |
| Scope | Accounts, support tickets, comments, history, logs and backups in one deployment |

## Purpose and data minimization

HNU Tickets processes information necessary to register support requests, assign responsibility, communicate and measure service performance. One deployment serves one organization. The organization designates an operator and a contact for access, correction and removal requests before operational use.

Do not request passwords, access tokens, private keys, financial identifiers or unrelated personal details in a ticket. Free-text messages can nevertheless contain sensitive information, so their access and retention must be controlled.

## Categories and sources

| Data | Source | Purpose and storage |
| --- | --- | --- |
| Name, email and account role | Registration and administrator updates | Identification, login and access control; `users` in PostgreSQL |
| Password hash | Password supplied during registration/bootstrap and hashed by the API | Credential verification; `users.password_hash`; never returned in the API |
| Ticket title, description, category, priority, author, assignee and timestamps | Requester input and authorized workflow operations | Support processing and workload visibility; `tickets` |
| Comment content, author and time | Participants with ticket access | Discussion and clarification; `comments` |
| Action, actor, time and old/new values | Ticket mutations | Traceability; `ticket_history`; old values may contain previous personal text |
| Access token with user ID, role and issue/expiry times | API token issuance | Bearer authentication; token delivered to the client; only revoked token UUIDs and expiry are retained in PostgreSQL |
| Revoked token UUID and expiry | Server logout | Block replay until expiry; `revoked_tokens`; expired entries are pruned on later logout writes |
| Method, route, status, duration, request ID and internal error context | API operation | Diagnostics and security investigation; operator-managed logs |
| Database backup and restoration records | Operator's backup process | Recovery; restricted storage outside the database host |

The server does not currently require attachments, payment details, a postal address or advertising identifiers. No email-delivery or advertising service is integrated into the documented server flow. Do not promise email notifications from collection of an email address.

## Access and transmission

USER can access their own tickets, comments and history; SPECIALIST can access assigned tickets; ADMIN can access all tickets and administer accounts. Only ADMIN can list/manage user records and request administrative analytics. Safe profile responses omit password hashes.

Database, log and backup access is restricted to designated operators. Application administrators do not automatically receive direct database or backup access. Support records and credentials must not be committed to the public source repository.

Transmit API data over HTTPS and use verified TLS for remote database connections. Encrypt off-host backups and protect encryption keys separately. Passwords are hashed rather than reversibly encrypted. Ticket descriptions, comment content and history values use AES-256-GCM field encryption. Titles, names, email and operational metadata remain readable in the database. See [Cryptography](cryptography.md) for key separation, rotation and protection limits.

The server has no documented automatic business-data transfer to third-party analytics or mail providers. Hosting operators may process stored data under the organization's deployment arrangement; the organization must identify actual providers and locations before operational use.

## Retention schedule

The following is the deployment policy to configure and enforce, not a claim that automatic retention jobs already exist.

| Record | Retention rule |
| --- | --- |
| Account profile | While the account is needed for organizational support; review after departure or an account removal request |
| Ticket, comments and associated history | While open, then 12 months after CLOSED/CANCELLED; periodically review necessity and any recorded incident hold |
| Routine operational logs | 30 days; a selected incident extract may be retained until investigation closure, then reviewed within 30 days |
| Rolling database backups | 30 days; expire copies through the backup storage lifecycle |
| Temporary test exports and restored copies | Remove when verification is complete; no later than 7 days unless a recorded investigation requires longer |

An incident hold must specify owner, reason, restricted scope and next review date. It must not silently turn normal retention into indefinite storage. The operator may choose a different justified schedule before deployment and record that choice consistently.

## Correction and removal

Participants contact the designated organization operator through its existing support channel. Verify identity and scope before returning data or applying a correction. Return only the person's permitted data, with other participants' information excluded where appropriate.

The API supports administrator profile updates but has no account-delete or ticket-delete endpoint and no retention job. Until a controlled removal procedure is implemented, the operator must schedule and verify restricted maintenance; immediate automated erasure is not available.

Removal must consider linked tickets, comments, history old/new text and backup copies. Do not simply delete a user row that remains referenced by foreign keys. Use a reviewed purge or consistent anonymization plan with schema checks and preserved necessary audit accountability. Record who performed the maintenance and its outcome without duplicating the removed content.

If sensitive material is present in a ticket, restrict exposure first and assess retained history before correcting the current text. After restoration, reapply recorded removals/anonymizations so old backup content does not silently reintroduce deleted data. Expire older backups according to the schedule.

## Logs, development and incidents

Do not log passwords, tokens, signing secrets, full request bodies or database credentials. Internal errors may carry sensitive context, so restrict access, review message construction and apply retention. Current request metadata is not a complete security-event log.

Use synthetic data in development, tests and screenshots. Operational data exports require a documented purpose, restricted recipient and deletion date. Never place a database dump in a GitHub issue or public artifact.

For suspected disclosure, restrict the affected interface/storage, revoke or rotate exposed credentials, preserve a minimal restricted incident record and assess affected data and recipients. The organization operator decides and records necessary communication through its established process.

## Verification and review

Check the SEC-R10 acceptance criteria in [Security requirements](security.md), including response fields, unauthorized object access, retention configuration and removal after restore. Review this policy when data categories, integrations, storage, retention or deployment providers change.

[Security policy](security_policy.md) defines the required technical controls. [Threat model](threat_model.md) covers disclosure, alteration and availability risks across the relevant data flows.
