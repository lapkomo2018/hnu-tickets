# Client–server trust

The server owns accounts, password hashes, encryption keys, tickets, comments, history and authorization decisions. The client owns presentation, input collection and an in-memory access token. Client-side validation improves usability; it never grants permission.

HNU Tickets uses `POST /api/v1/auth/login` to issue a short-lived HS256 JWT. The client sends `Authorization: Bearer <token>` over HTTPS. Middleware validates the signature, algorithm, subject, token ID and expiration, checks revocation, and loads the current account role. Services enforce ticket ownership, assignment and state. An edited JWT role, hidden UI control or supplied user ID cannot bypass these checks. `POST /api/v1/auth/logout` revokes the current token before the client clears its state. Expiration requires login again; no refresh token is issued. See [authentication](authentication.md) for endpoint and permission details.

## Educational Library client

Run the supplied `lab6_library_server.py` locally (port 8080, or pass another port), then:

```sh
python3 scripts/lab6/client.py --base http://127.0.0.1:8080 --client-id peony-07
```

Enter variant 7's secret from the assignment at the hidden prompt. The client authenticates automatically, retains the token only in memory, and displays the reader, books, active loans and summary. It refreshes an expired token before requesting data and retries once after 401. Redirects are refused to avoid forwarding credentials to another endpoint. Plain HTTP is accepted only on loopback; normal connections require HTTPS with certificate validation. The secret remains available in process memory for reauthentication and must not be embedded in a distributed client.

`manual_results.json` records curl requests to the public routes, token endpoint and all five protected routes. The token value is omitted. Results: `/doc` 200; `/coffee_maker` 418; GET `/auth/token` 405; unauthenticated `/whoami` and incorrect credentials 401. Valid authentication yields a Bearer token with a 1,800-second lifetime. Reader R-007 has six books, four available and two active loans, watermark `LAB6-V07-R007`. Book `BK-07-01` is *Code Complete* (2004). The client was also exercised with an invalid cached token and forced local expiry; both paths recovered successfully.

The training API is an educational token protocol, not a complete OAuth implementation. Its embedded shared credential is suitable for the exercise only. HNU Tickets uses individual user credentials and server-side authorization.
