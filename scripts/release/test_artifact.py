import tempfile
import unittest
from pathlib import Path

from cryptography.exceptions import InvalidSignature
from artifact import generate, sign, verify


class ArtifactTests(unittest.TestCase):
    def test_authenticity_and_tampering(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            artifact, private, public, signature = [root / n for n in ('api', 'private', 'public', 'signature')]
            artifact.write_bytes(b'HNU Tickets release')
            generate(private, public)
            sign(artifact, private, signature)
            verify(artifact, public, signature)
            artifact.write_bytes(b'HNU Tickets modified')
            with self.assertRaises(InvalidSignature):
                verify(artifact, public, signature)
            # Recomputing the public hash does not repair the signature.
            from artifact import message
            message(artifact)
            with self.assertRaises(InvalidSignature):
                verify(artifact, public, signature)
            artifact.write_bytes(b'HNU Tickets release')
            signature.write_bytes(bytes(64))
            with self.assertRaises(InvalidSignature):
                verify(artifact, public, signature)

    def test_wrong_signer_and_exclusive_key_creation(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            artifact = root / 'api'
            artifact.write_bytes(b'payload')
            generate(root / 'a', root / 'a.pub')
            generate(root / 'b', root / 'b.pub')
            sign(artifact, root / 'a', root / 'sig')
            with self.assertRaises(InvalidSignature):
                verify(artifact, root / 'b.pub', root / 'sig')
            with self.assertRaises(FileExistsError):
                generate(root / 'a', root / 'a.pub')


if __name__ == '__main__':
    unittest.main()
