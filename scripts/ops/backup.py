"""Authenticated encrypted PostgreSQL backups; restore only into a new database."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tempfile
from cryptography.fernet import Fernet


def run(*args, **kwargs):
    return subprocess.run(args, check=True, capture_output=True, **kwargs)


def backup(destination, database):
    key = Fernet(os.environ['BACKUP_KEY'].encode())
    with tempfile.TemporaryDirectory() as directory:
        dump = Path(directory) / 'database.dump'
        run('pg_dump', '--format=custom', '--no-owner', '--no-acl', '--file', str(dump), '--dbname', database)
        raw = dump.read_bytes()
        envelope = key.encrypt(raw)
        with Path(destination).open('xb') as stream:
            os.chmod(destination, 0o600)
            stream.write(envelope)
    return {'bytes': len(envelope), 'sha256': hashlib.sha256(envelope).hexdigest()}


def restore(source, database):
    # Identifiers are deliberately limited; no connection URI or existing target is accepted.
    import re
    if not re.fullmatch(r'[a-z][a-z0-9_]{0,50}', database):
        raise ValueError('invalid target database name')
    raw = Fernet(os.environ['BACKUP_KEY'].encode()).decrypt(Path(source).read_bytes())
    with tempfile.TemporaryDirectory() as directory:
        dump = Path(directory) / 'database.dump'
        dump.write_bytes(raw)
        run('pg_restore', '--list', str(dump))
        run('createdb', '--', database)  # Fails if the database already exists.
        run('pg_restore', '--exit-on-error', '--single-transaction', '--no-owner', '--no-acl', '--dbname', database, str(dump))
    return {'restored_database': database}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('operation', choices=['backup', 'restore', 'verify'])
    parser.add_argument('file')
    parser.add_argument('--database')
    args = parser.parse_args()
    try:
        if args.operation == 'verify':
            Fernet(os.environ['BACKUP_KEY'].encode()).decrypt(Path(args.file).read_bytes())
            result = {'verified': True}
        elif not args.database:
            parser.error('--database required')
        elif args.operation == 'backup': result = backup(args.file, args.database)
        else: result = restore(args.file, args.database)
        print(json.dumps(result))
    except Exception as error:
        parser.exit(1, f'Backup operation failed ({type(error).__name__})\n')


if __name__ == '__main__': main()
