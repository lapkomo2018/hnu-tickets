# Authentication and authorization

## Identity and methods

Email is the normalized login identifier: trim surrounding whitespace and lowercase it during registration and login. The database guarantees uniqueness. A UUID identifies the account in resource references and the JWT subject; clients do not choose the acting user ID.

Password authentication is the available mandatory factor. MFA is not enabled. A future TOTP factor would require separate enrollment, encrypted factor secrets, recovery and verification before token issuance; no second-factor protection is claimed for the current API.

## Password storage and input

Use bcrypt at cost 12 through `golang.org/x/crypto/bcrypt`; the library generates a separate salt for each hash. Existing supported bcrypt hashes remain verifiable. Never store or return plaintext passwords or hashes in profile responses.

Passwords must be valid UTF-8, contain at least 8 characters and occupy at most 72 bytes. Check the byte limit during registration, bootstrap and login; do not truncate or pre-hash a long password. A multibyte password can reach the byte limit before 72 characters.

Wrong passwords and unknown emails return the same `INVALID_CREDENTIALS` response. Unknown accounts perform a dummy bcrypt comparison to avoid a fast missing-account path; this does not establish strict constant-time response behavior across all accounts and network conditions.

## Access tokens and current permissions

Issued HS256 JWTs contain `sub`, `jti`, `role`, `iat` and `exp`. Require a valid signature, allowed algorithm, nonempty UUID subject/token ID and expiration. Tokens issued before the `jti` requirement require a new login.

For each protected request, middleware checks persistent revocation state and loads the account's **current role** from PostgreSQL. The JWT role is informational and never overrides that current role. Missing accounts and revoked/invalid/expired tokens receive 401. Database failures deny access with a generic server error rather than allowing the request.

The checked principal is stored in request context. Application services then enforce role, ownership/assignment and ticket-state rules. A role change affects subsequent authorization checks, including requests using an older token. Requests that already passed authentication are not retroactively cancelled.

## Endpoints and logout

| Endpoint | Behavior |
| --- | --- |
| `POST /api/v1/auth/register` | Validate profile/password, create USER and issue a token |
| `POST /api/v1/auth/login` | Verify credentials and issue an independently identified token |
| `GET /api/v1/auth/me` | Return the current authenticated account through a safe DTO |
| `POST /api/v1/auth/logout` | Persist the current token ID and its expiry in `revoked_tokens`; return 204 |

Clients call logout before deleting their stored token. Subsequent protected requests using that token, including a copied token, fail with 401. Other tokens for the same account remain usable. Revocation survives API restart because it is stored in PostgreSQL. If logout fails, the client must not claim that the server revoked the token.

Apply migration `000006` before deploying this version. Expired revocation entries no longer affect access and are pruned during later logout writes. Operators may additionally schedule expiry cleanup. No plaintext token is persisted in the revocation table.

## Access control model

Role-based permissions are combined with resource attributes: creator, assigned specialist and ticket status. USER can create tickets and access their own records; SPECIALIST can access assigned records; ADMIN manages accounts, roles, categories and assignment. Comments/history reuse ticket visibility. Author field edits explicitly require the current USER role, ownership and NEW status.

This combines RBAC with object/state checks and avoids granting all members of a role access to every ticket. Full permissions and lifecycle rules are in [Business logic](business_logic.md).

## Authentication request limits

Per API process and socket peer, allow 10 login requests and 5 registration requests per 60-second fixed window. Rejections return 429 and `Retry-After`; successful and failed attempts both consume the budget. The limiter is mutex-protected and holds at most 4096 peer buckets; when full, new peers are denied until expired entries can be cleared.

Forwarded headers are not trusted for identity. A reverse-proxy deployment must configure an appropriate trusted peer/rate-limit policy at ingress. Multiple API replicas require shared or ingress enforcement; the in-process limit is not a global distributed quota.

## Verification

Tests cover registration, duplicate email, normalized login, invalid credentials, 72-byte/multibyte boundaries, missing token claims, signature/expiry/algorithm rejection, current role changes, deleted accounts, per-token logout, independent sessions, storage failure and concurrent/rate-limited attempts. An isolated PostgreSQL test verifies persistence, idempotent revocation and expired-entry cleanup.

Run `make test`, `make lint`, `go test -race ./...` and `go build ./...`. Set `TEST_DATABASE_URL` to an isolated PostgreSQL database to include the revocation integration check. Regenerate the endpoint contract with `make swagger` after API changes.
