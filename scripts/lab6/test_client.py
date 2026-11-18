import time
import unittest
import urllib.error
from client import Client


class ClientTests(unittest.TestCase):
    def test_rejects_insecure_remote_and_url_credentials(self):
        for url in ['http://example.com', 'https://user:pass@example.com', 'https://example.com?token=x']:
            with self.assertRaises(ValueError): Client(url, 'id', 'secret')

    def test_401_reauthenticates_once(self):
        client = Client('https://example.test', 'id', 'secret')
        client.token, client.expires = 'old', time.monotonic() + 100
        calls = []
        def request(path, body=None, token=None):
            calls.append((path, token))
            if path == '/auth/token': return dict(token_type='Bearer', access_token='new', expires_in=1800)
            if token == 'old': raise urllib.error.HTTPError('', 401, '', {}, None)
            return {'ok': True}
        client.request = request
        self.assertEqual(client.get('/whoami'), {'ok': True})
        self.assertEqual(calls, [('/whoami', 'old'), ('/auth/token', None), ('/whoami', 'new')])

    def test_forbidden_is_not_retried(self):
        client = Client('https://example.test', 'id', 'secret')
        client.token, client.expires = 'token', time.monotonic() + 100
        calls = []
        def request(*args, **kwargs):
            calls.append(args)
            raise urllib.error.HTTPError('', 403, '', {}, None)
        client.request = request
        with self.assertRaises(urllib.error.HTTPError): client.get('/whoami')
        self.assertEqual(len(calls), 1)


if __name__ == '__main__': unittest.main()
