"""Reproduce UTF-8/bcrypt and 20-candidate password experiments for variant 7."""

import ctypes
import ctypes.util
import datetime
import hashlib
import hmac
import json
import platform
import secrets
import string
import sys
from pathlib import Path

import bcrypt

from dictionary_bcrypt import search as bcrypt_search
from dictionary_sha256 import search as sha_search


def main():
    # libxcrypt exposes the traditional bcrypt behaviour without pre-truncating input.
    library = ctypes.util.find_library("crypt")
    if library is None:
        raise RuntimeError("the 72-byte demonstration requires Linux libxcrypt")
    crypt = ctypes.CDLL(library)
    crypt.crypt.argtypes = [ctypes.c_char_p, ctypes.c_char_p]
    crypt.crypt.restype = ctypes.c_char_p
    crypt.crypt_gensalt_rn.argtypes = [ctypes.c_char_p, ctypes.c_ulong,
                                     ctypes.c_char_p, ctypes.c_int,
                                     ctypes.c_char_p, ctypes.c_int]
    crypt.crypt_gensalt_rn.restype = ctypes.c_char_p

    def legacy_hash(password):
        output = ctypes.create_string_buffer(128)
        salt = crypt.crypt_gensalt_rn(b"$2b$", 12, secrets.token_bytes(16), 16, output, len(output))
        if not salt:
            raise RuntimeError("bcrypt salt generation failed")
        digest = crypt.crypt(password, salt)
        if not digest or not digest.startswith(b"$2b$12$"):
            raise RuntimeError("bcrypt hashing failed")
        return digest.decode("ascii")

    seed = "2026_ІПЗс-24-1_Лапко_Максим_варіант:7_"
    size = len(seed.encode("utf-8"))
    if size > 72:
        raise ValueError("student seed already exceeds 72 UTF-8 bytes")
    base = seed + "_" * (72 - size)
    suffix = "".join(secrets.choice(string.ascii_letters + string.digits) for _ in range(10))
    variants = [base, base + "1", base + suffix, base[:-1]]
    hashes = [legacy_hash(value.encode("utf-8")) for value in variants]
    matrix = []
    for reference in hashes:
        row = []
        for value in variants:
            actual = crypt.crypt(value.encode("utf-8"), reference.encode("ascii"))
            row.append(hmac.compare_digest(actual, reference.encode("ascii")))
        matrix.append(row)

    strict = []
    for value in variants:
        try:
            bcrypt.hashpw(value.encode("utf-8"), bcrypt.gensalt(rounds=4))
            strict.append("accepted")
        except ValueError:
            strict.append("rejected: password exceeds 72 bytes")

    password = "pass: Лапко"
    candidates = ["pass: " + surname for surname in [
        "Бондар", "Мельник", "Шевченко", "Коваленко", "Бойко", "Ткаченко", "Кравченко",
        "Олійник", "Шевчук", "Поліщук", "Савченко", "Козак", "Мороз", "Петренко",
        "Лисенко", "Ткачук", "Кравчук", "Клименко", "Гриценко", "Лапко"]]
    hash5 = hashlib.sha256(password.encode("utf-8")).hexdigest()
    hash6 = bcrypt.hashpw(password.encode("utf-8"), bcrypt.gensalt(rounds=12)).decode("ascii")
    results = {
        "measured_at_utc": datetime.datetime.now(datetime.timezone.utc).isoformat(),
        "environment": {"platform": platform.platform(), "python": sys.version.split()[0],
                        "bcrypt": bcrypt.__version__, "legacy_library": library},
        "seed": seed, "seed_characters": len(seed), "seed_utf8_bytes": size,
        "padding_underscores": 72-size, "random_suffix": suffix,
        "variants": [{"id": i+1, "value": value, "utf8_bytes": len(value.encode("utf-8")),
                      "hash": digest, "strict_python_bcrypt": accepted}
                     for i, (value, digest, accepted) in enumerate(zip(variants, hashes, strict))],
        "verification_matrix_hash_rows_string_columns": matrix,
        "dictionary": candidates, "password": password, "hash5_sha256": hash5, "hash6_bcrypt_cost12": hash6,
        "sha256_search": sha_search(hash5, candidates), "bcrypt_search": bcrypt_search(hash6, candidates),
        "clock": "time.perf_counter_ns; elapsed_us = delta_ns / 1000",
    }
    destination = Path(__file__).resolve().parent / "results.json"
    destination.write_text(json.dumps(results, ensure_ascii=False, indent=2)+"\n", encoding="utf-8")
    (destination.parent / "dictionary.json").write_text(json.dumps(candidates, ensure_ascii=False, indent=2)+"\n", encoding="utf-8")
    print(json.dumps(results, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
