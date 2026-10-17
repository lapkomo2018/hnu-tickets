"""Verify an offline SHA-256 hash against exactly 20 supplied candidates."""

import argparse
import hashlib
import hmac
import json
import time
from pathlib import Path


def search(reference, candidates):
    if len(candidates) != 20:
        raise ValueError("the laboratory dictionary must contain exactly 20 entries")
    start = time.perf_counter_ns()
    for attempts, candidate in enumerate(candidates, 1):
        digest = hashlib.sha256(candidate.encode("utf-8")).hexdigest()
        if hmac.compare_digest(digest, reference):
            return {"candidate": candidate, "attempts": attempts,
                    "elapsed_us": round((time.perf_counter_ns() - start) / 1000, 3)}
    return {"candidate": None, "attempts": len(candidates),
            "elapsed_us": round((time.perf_counter_ns() - start) / 1000, 3)}


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("hash")
    parser.add_argument("dictionary", type=Path)
    args = parser.parse_args()
    print(json.dumps(search(args.hash, json.loads(args.dictionary.read_text())), ensure_ascii=False, indent=2))
