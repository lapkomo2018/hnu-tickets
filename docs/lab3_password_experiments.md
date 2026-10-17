# Password experiments

## UTF-8 input and bcrypt boundary

Student string: `2026_ІПЗс-24-1_Лапко_Максим_варіант:7_`.

The seed contains 38 characters and 60 UTF-8 bytes. Append 12 underscores to obtain 72 bytes. Variant 2 appends `1`, variant 3 appends the recorded random suffix `wrXhruUuYb`, and variant 4 removes the final underscore.

The first four hashes use Linux libxcrypt's bcrypt implementation at cost 12, which demonstrates the traditional 72-byte prefix behavior without manually truncating the supplied input. A separate check with Python bcrypt 5.0.0 rejects the 73-byte and 82-byte inputs.

| Hash | UTF-8 bytes | Result |
| --- | --- | --- |
| 1 | 72 | `$2b$12$6vB.iEsZGWn1AZTgIKiUbO8r21epm9Zlk5WY/uilpbYLobKYoAp96` |
| 2 | 73 | `$2b$12$w15NXnD4pH811uJjGmAuE.e74.w4x7nc..cPR/1UyBmZtQHQNS4Z2` |
| 3 | 82 | `$2b$12$fBuz8g0gnq0g3rodq9ei1eHBX.hxY2QUCNksnbnGn3DtiAQEDxiQ6` |
| 4 | 71 | `$2b$12$nfrVVGKvcyMuItyGonwBn.q4T939efgOR.DzqQ9ttezp2OxY0O66q` |

### Verification matrix

Each row is the stored hash and each column is the complete supplied string.

| Hash / String | 1 (72 B) | 2 (73 B) | 3 (82 B) | 4 (71 B) |
| --- | --- | --- | --- | --- |
| 1 | match | match | match | no match |
| 2 | match | match | match | no match |
| 3 | match | match | match | no match |
| 4 | no match | no match | no match | match |

Hashes 1–3 are different strings because each uses an independently generated salt. Nevertheless, all three verify inputs 1–3: the additional bytes are outside the traditional bcrypt limit. Removing the final byte changes the effective password, so input 4 matches only hash 4. Strict implementations reject oversized passwords instead of demonstrating truncation.

## Dictionary comparison

Password: `pass: Лапко`. The JSON dictionary has 20 distinct entries, with the correct candidate last. The scripts stop at the first verified match.

- Hash 5 (SHA-256): `aef5f54b3a371fec2e09d46047321fe96b96af760af58be561b3852f6b2488bc`
- Hash 6 (bcrypt cost 12): `$2b$12$kya1MeEFjkf2qAQR0k.ZDOPLziYRNY.r2bcp1rusEQvqns.TeA6by`

| Algorithm | Attempts | Elapsed microseconds |
| --- | --- | --- |
| SHA-256 | 20 | 24.216 |
| bcrypt, cost 12 | 20 | 4955842.291 |

Measured at 2026-10-17T19:57:25.267615+00:00, using Python 3.12.14 and bcrypt 5.0.0 on Linux-6.18.55-x86_64-with-glibc2.39. Timing uses a monotonic nanosecond counter converted to microseconds; setup and initial hash creation are excluded.

This is one sequential run over a fixed 20-entry dictionary, not a throughput or hardware-capacity benchmark. Bcrypt deliberately makes each password check more expensive; SHA-256 alone is unsuitable for password storage.

Raw inputs, salts/hashes, the full matrix and timing results are in [results.json](../scripts/lab3/results.json). Reproduction steps are in [the script guide](../scripts/lab3/README.md).
