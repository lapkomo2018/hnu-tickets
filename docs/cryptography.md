# Cryptography

## Protected data

AES-256-GCM encrypts `tickets.description`, `comments.content`, and non-NULL `ticket_history.old_value` / `new_value`. History protection prevents old descriptions from remaining readable after a ticket edit. PostgreSQL adapters encrypt before INSERT/UPDATE and decrypt scanned values; domain objects and API DTOs contain ordinary text. Authorization remains mandatory.

Titles remain readable for title search. Names, email, category, role, status, priority, relationships, timestamps and approximate text length remain visible. This protection limits disclosure from a database-only copy; an attacker controlling the API process or its keys can decrypt the fields. It does not prevent deletion, replay of an older valid value for the same field, or privileged modification of unencrypted metadata. HTTPS and encrypted whole-database backups are still required.

## Format and authentication

Format: `enc:v1:<key-id>:<base64(nonce || ciphertext || tag)>`. Nonce: 12 bytes; tag: 16 bytes. Go `crypto/cipher.NewGCMWithRandomNonce` generates/prepends the nonce for each write. The envelope header and `hnu-tickets/<table>/<row-uuid>/<column>` are authenticated as AAD. UUIDs are assigned before insertion. Copying a field to another record/column fails authentication. NULL history fields stay NULL; an empty string is encrypted normally.

Malformed data, unknown key/version, a changed nonce/tag/ciphertext or wrong AAD return an error and no plaintext. The API returns its generic internal-error response. There is no automatic plaintext fallback. Keys and decrypted text must not be logged.

## Keys

Generate a separate random 32-byte key with `openssl rand -base64 32`. Set `DATA_ACTIVE_KEY=k1` and `DATA_KEYS={"k1":"<base64 key>"}` in the runtime environment. IDs allow 1–32 ASCII letters, digits, underscores or hyphens. Startup validates configuration; `.env.example` contains only a placeholder. Do not reuse `JWT_SECRET` as an encryption key. Restrict environment/secret-file access and keep real keys outside Git and database backups. Use the deployment's secret manager where available; this repository does not provision one.

Rotate at least annually, immediately after suspected disclosure, and before one million encryptions under a key. This operating threshold is well below Go's maximum of 2^32 messages per random-nonce GCM key and applies across every process/field using that key. Operators track key age and write volume; automatic scheduling/counting is not implemented.

## Initial conversion

1. Stop every API writer and take a restricted backup. Preserve the original database and keys until conversion is verified.
2. Apply schema migrations through `000007`. Optionally load demo accounts/categories with `make seed`.
3. Configure `DATA_KEYS` / `DATA_ACTIVE_KEY` and run `make crypto-migrate` (`go run ./cmd/cryptodata -mode migrate`). An empty database also needs this initialization.
4. Start the API. Its startup requires `data_crypto_state.encrypted = TRUE`.

The command locks the state, ticket, comment and history tables, processes fields in pages of 500, and commits every rewritten value and the marker in one transaction. The marker determines whether input is plaintext; content beginning with `enc:` is not guessed to be encrypted. Repeating migration on an initialized database is rejected. Failure or interruption rolls back the transaction. Run maintenance in an outage window; all writers must be stopped, including older API versions. Do not insert plaintext ticket data through SQL after conversion.

Docker equivalent after migrations: `docker compose run --build --rm --no-deps --entrypoint /cryptodata api -mode migrate`. The image includes both `/api` and `/cryptodata`.

## Rotation

1. Stop all API instances and take a restricted backup.
2. Add a fresh key under a new ID to `DATA_KEYS`, retain every old key needed to read existing rows, and select the new ID in `DATA_ACTIVE_KEY`.
3. Run `make crypto-rotate` (`go run ./cmd/cryptodata -mode rotate`). Docker: use the same command above with `-mode rotate`.
4. Verify authorized reads, then restart every instance with the new active ID. Remove old keys from live configuration only after successful conversion.

Rotation authenticates/decrypts each existing value before re-encrypting; an unknown key or damaged field aborts the transaction. Retain old keys separately for retained backups that still require them. After suspected compromise, rotation cannot undo earlier disclosure. SQL rollback of migration `000007` does not decrypt fields: restore the matching pre-conversion backup for a downgrade, or implement a reviewed explicit recovery procedure.

## Verification

`go test ./...` covers Unicode/empty text, nonce variation, nonce/ciphertext/tag tampering, wrong key, wrong record/column/table, unknown envelope versions/IDs, old-key reads and new-key rotation. In-memory transaction/row fakes exercise all four fields, 500-row pagination, the zero UUID, initial plaintext beginning with `enc:`, state guards and corrupt rotation input. No database service is required by these tests.

## References

- [Go crypto/cipher](https://pkg.go.dev/crypto/cipher#NewGCMWithRandomNonce)
- [NIST SP 800-38D](https://csrc.nist.gov/pubs/sp/800/38/d/final)
