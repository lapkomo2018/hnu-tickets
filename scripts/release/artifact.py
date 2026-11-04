"""Sign and verify a single HNU Tickets release artifact with Ed25519."""
import argparse
import hashlib
from pathlib import Path

from cryptography.exceptions import InvalidSignature
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric.ed25519 import Ed25519PrivateKey, Ed25519PublicKey

DOMAIN = b"hnu-tickets/release/v1\0"


def message(path):
    digest = hashlib.sha256()
    with Path(path).open("rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(chunk)
    return DOMAIN + digest.digest()


def generate(private, public):
    key = Ed25519PrivateKey.generate()
    # Exclusive creation prevents accidental replacement of an existing key.
    with Path(private).open("xb") as stream:
        Path(private).chmod(0o600)
        stream.write(key.private_bytes(serialization.Encoding.PEM,
                     serialization.PrivateFormat.PKCS8, serialization.NoEncryption()))
    with Path(public).open("xb") as stream:
        stream.write(key.public_key().public_bytes(serialization.Encoding.PEM,
                     serialization.PublicFormat.SubjectPublicKeyInfo))


def sign(artifact, private, signature):
    key = serialization.load_pem_private_key(Path(private).read_bytes(), password=None)
    if not isinstance(key, Ed25519PrivateKey):
        raise ValueError("expected an Ed25519 private key")
    with Path(signature).open("xb") as stream:
        stream.write(key.sign(message(artifact)))


def verify(artifact, public, signature):
    key = serialization.load_pem_public_key(Path(public).read_bytes())
    if not isinstance(key, Ed25519PublicKey):
        raise ValueError("expected an Ed25519 public key")
    key.verify(Path(signature).read_bytes(), message(artifact))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)
    gen = sub.add_parser("keygen")
    gen.add_argument("private")
    gen.add_argument("public")
    for name in ("sign", "verify"):
        command = sub.add_parser(name)
        command.add_argument("artifact")
        command.add_argument("key")
        command.add_argument("signature")
    args = parser.parse_args()
    try:
        if args.command == "keygen":
            generate(args.private, args.public)
        elif args.command == "sign":
            sign(args.artifact, args.key, args.signature)
        else:
            verify(args.artifact, args.key, args.signature)
        print("OK")
    except (OSError, ValueError, TypeError, InvalidSignature):
        parser.exit(1, "Artifact operation failed\n")


if __name__ == "__main__":
    main()
