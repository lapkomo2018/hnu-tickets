# Lab 4: classical ciphers and AES-CTR

Run from the repository root:

```sh
go run ./scripts/lab4 -output scripts/lab4/results.json
```

The UTF-8 input uses variant `07`; the alternative uses `99` so the byte lengths stay equal. Caesar uses shift 7 separately within the Ukrainian and Latin uppercase/lowercase alphabets; digits, underscores, punctuation and hyphens stay unchanged. Column transposition writes runes in rows of 7 and reads columns in zero-based order `[2,0,6,1,4,3,5]`, omitting missing cells with no padding. Both classical transformations are reversed and checked.

AES-256-CTR uses a random 32-byte key and random 16-byte initial counter. Go increments the full 128-bit counter in big-endian order; no padding is used. The attack computes `gamma = C1 XOR P1`, then `C2 = gamma XOR P2`, without receiving the key. The owner decrypts C2 using the original key/IV and verifies the text containing variant `99`.

The known pair reveals only the keystream for that message span, not the AES key. Bare CTR does not authenticate data, so the altered ciphertext is accepted. These results do not show a break of AES itself. The application uses AES-GCM instead.

`results.json` records one actual run, including its intentionally public **experiment-only** key/IV. They must never be used as application keys. A new run changes the key, IV and derived hex values; the verified substitution remains the same.
