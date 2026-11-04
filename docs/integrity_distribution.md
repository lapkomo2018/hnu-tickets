# Integrity and distribution

HNU Tickets is operated as a hosted service. Clients receive API access rather than server secrets or a trusted local license checker. The existing verified JWT, current database role, object ownership checks and revoked-token store enforce access on the server. Copying a browser client does not grant access to protected API operations. No paid subscription, device activation or billing system is introduced.

## Protected data

| Data and user influence | Consequence of unauthorized modification | Control |
| --- | --- | --- |
| Ticket descriptions, comments and history values supplied through the API | Altered support records or disclosure through substituted ciphertext | AES-256-GCM with authenticated version/key ID and table/record/column context; decryption fails closed |
| JWT supplied in Authorization | Impersonation or elevated role | HS256 validation, expiry, revocation and current server-side role |
| Ticket status, ownership and assignments | Unauthorized workflow transitions or access | Server authorization, validation and database constraints; encryption alone does not authorize a change |
| Release archive downloaded by an operator | Execution of substituted code | Detached Ed25519 signature over a domain-separated SHA-256 digest |

An ordinary SHA-256 checksum detects accidental corruption only when its expected value is trusted. An attacker able to replace both the file and checksum can recompute the checksum. Release authentication therefore uses a signature; recipients possess only the public verification key. Existing AES-GCM protects stored field authenticity but does not detect deletion or replay of an older valid record. The release verifier likewise does not enforce version freshness.

## Signed release workflow

Install `scripts/release/requirements.txt` in a dedicated Python environment. Generate the signing key outside the repository and keep it with the release operator. Private keys are not packaged, uploaded or required by the API.

```sh
python scripts/release/artifact.py keygen /secure/release-private.pem /secure/release-public.pem
python scripts/release/artifact.py sign dist/hnu-tickets.tar.gz /secure/release-private.pem dist/hnu-tickets.tar.gz.sig
python scripts/release/artifact.py verify dist/hnu-tickets.tar.gz /trusted/release-public.pem dist/hnu-tickets.tar.gz.sig
```

The archive must include every file used for deployment, including the executable and migrations. Verify before extracting or executing it. Only exit status 0 permits deployment; missing files, a malformed key, invalid signature or modified archive result in nonzero status. Verification streams the artifact and signs the exact message `hnu-tickets/release/v1\0 || SHA256(artifact)`.

Provision the public key through an independently trusted administration channel. A public key downloaded alongside an untrusted archive does not establish authenticity. Protect the verification tool and its dependencies as part of the deployment environment. Verify a private, immutable copy and deploy that same copy to avoid replacement after verification. Keep an approved release version to prevent rollback to an older valid archive. Key rotation requires explicitly replacing the trusted public key through the administration channel.

This is an operator-invoked release check, not automatic verification by the API at startup. A compromised release signing key or administrator remains outside its protection boundary. The public source repository does not provide technical secrecy or prevent independent copies; the hosted instance's access controls protect its service and data.

## Obfuscation experiment

`scripts/lab5` contains the three unmodified Appendix A programs and an original/obfuscated pair using the student's key. The obfuscated version adapts Appendix A's dispatcher, fragmented hash checks, opaque predicates, indirect calls and authenticated encrypted output. It is an educational console program, not the API authentication mechanism.

```sh
python scripts/lab5/run_experiments.py
python -m unittest discover -s scripts/release -v
```

The runner checks valid and invalid inputs and records actual dispatcher states using `sys.settrace`. The results are saved in `scripts/lab5/results.json`. Obfuscation increases analysis effort but does not make client-side secrets inaccessible or prevent patching of the local decision.
