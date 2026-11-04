# Lab 5 experiments

Requires Python 3.10+ and the dependencies pinned in `../release/requirements.txt`.

```sh
python -m pip install -r scripts/release/requirements.txt
python scripts/lab5/run_experiments.py
python scripts/lab5/original.py
python scripts/lab5/obfuscated.py
```

Key: `2026_ІПЗс-24-1_Лапко_Максим_варіант:7_`.

`appendix_original.py`, `appendix_hash.py` and `appendix_obfuscated.py` preserve the supplied Appendix A examples. The student's obfuscated version adapts the third example, replacing the expected hash fragments and regenerating the authenticated success payload for the new key. The plaintext key remains in the original program and experiment runner for reproducibility, not as a production credential.

Successful Appendix A dispatch: `1, 8, 15, 2, 5, 4, 11, 6, 9, 0, 7`. Invalid input follows the same sequence, fails AES-GCM authentication at state 7 and continues to state 10, which decrypts the denial message. `_t6(x)` is always true for integers because `x*(x+1)` is even. Names `decrypt` and `exec` are reconstructed from character codes; the decrypted snippets print the result.

The hash-only version conceals the literal password but permits offline candidate checks and local branch patching. The encrypted version also ties the success payload to the input, so changing a comparison alone does not decrypt it. A researcher controlling the process can still inspect decrypted code or replace the entire decision. No claim of absolute reverse-engineering resistance is made.
