# Libraxis incident, 17 November 2026

The supplied application and access logs describe one incident. Access timestamps use UTC+02:00; application timestamps omit the timezone. Correlation assumes the same clock and zone, supported by matching login/error times. No database audit or response bodies are available.

## Application log alone

Successful logins are followed by SQL syntax/column errors, creation of readers and loans export files, changes to account 1's display name/avatar, a reader-note update, and database errors. Error 1046 says no database is selected; subsequent 1049 errors say `catalog1` is unknown. Thus the configured database is unavailable, although this does not establish that the database server is down. SQL injection and deletion are plausible, but the log alone does not identify an attacker or prove either. Export creation does not prove delivery or a data leak.

## Access log alone

IP `185.220.101.22` repeatedly receives 401 at login, then 200, probes routes, sends SQL injection payloads, reads session-related data, enters admin endpoints, downloads exports and sends a `DROP DATABASE catalog1` payload. Other clients subsequently receive 500. This establishes malicious requests and successful HTTP responses, not the exact returned records or execution of every SQL statement. The export responses strongly indicate data disclosure; the reason for the database connection failure remains an inference without application/database evidence.

## Correlated timeline (UTC+02:00)

| Time | Evidence and interpretation |
| --- | --- |
| 14:03:04 | Legitimate administrator login from 203.0.113.10. |
| 14:04:02–14:04:59 | Repeated 401 login responses from 185.220.101.22. |
| 14:05:02 | Login 200 matches application login for `shelby`. |
| 14:05:12–14:05:20 | Route enumeration; `/doc` returns 200. |
| 14:05:39 | Quote probe returns 500, matching SQL syntax error 1064. |
| 14:05:45–14:05:56 | Boolean and ORDER BY probes; column 5 fails, column 4 succeeds. |
| 14:06:04–14:06:19 | Four-column UNION and metadata enumeration return 200. |
| 14:06:27 | Query selects user ID, access token and expiry from sessions; HTTP 200, 2,387 bytes. |
| 14:06:41 | Same IP enters admin dashboard without a recorded admin login. Session theft is the likely bridge; token contents are not logged. |
| 14:06:44–14:07:10 | Five reader pages and three loan pages fetched. |
| 14:07:16; 14:07:24 | Reader and loan export responses: 684,217 and 1,128,946 bytes. Application log confirms export creation. Total response volume: 1,813,163 bytes; not a count of exposed records. |
| 14:07:31–14:07:38 | Audit and system information read. |
| 14:07:54–14:07:58 | Account 1 display name changed to `StupedPotato`, avatar to `monkey_face.png`; matching requests from the suspicious IP. |
| 14:08:08 | Reader-note update from the legitimate administrator IP. It must not be attributed to the attacker merely because it also uses account 1. |
| 14:08:16 | Database-drop payload, HTTP 500 and application error 1046. |
| From 14:08:19 | Widespread HTTP 500 and error 1049 (`catalog1` unknown). |

The most consistent chain is credential compromise → SQL injection → session disclosure → privileged access and export → database deletion. Exact leaked records and the stolen session remain unverified without response/database evidence. The final failure is consistent with a missing database rather than network connectivity loss.

## Logging gaps

The application log omits failed login attempts, IPs, request IDs, consistent levels and structured outcomes. Several entries omit date/zone, actor or target. Stack details expose internal paths. Access logs add sequence, source, method, URL, status and response size, but lack verified identity, application action and database outcome. Neither log independently proves session theft or exact disclosure. Use server-generated correlation IDs, verified actor IDs, object IDs, consistent timestamps and sanitized error categories. Never log session tokens, passwords or full sensitive query strings.

## Recovery order

1. Restrict affected routes and privileged sessions; preserve read-only copies of both logs and record hashes.
2. Revoke sessions and rotate compromised credentials. Investigate database/server state and preserve forensic evidence before repair.
3. Fix SQL construction with parameters; remove DDL/drop privileges from the runtime account. Review administrative access and rate limits.
4. Restore a verified pre-incident backup into an isolated new database; inspect schema, row counts and privileged account changes. Recover later legitimate changes only from trustworthy evidence.
5. Deploy the corrected application and protected configuration/keys; verify login, authorization, reads and writes. Assess the actual recovery point and data-loss interval.
6. Reopen access with monitoring, document the incident and handle confirmed disclosure according to the project's incident policy.
