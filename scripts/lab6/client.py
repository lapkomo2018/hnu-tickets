"""Memory-only bearer client for the educational Library API."""
import argparse
import getpass
import json
import time
import urllib.error
import urllib.parse
import urllib.request


class Client:
    def __init__(self, base, client_id, secret):
        url = urllib.parse.urlsplit(base)
        if url.scheme != 'https' and not (url.scheme == 'http' and url.hostname in ('127.0.0.1', 'localhost', '::1')):
            raise ValueError('HTTPS required except on loopback')
        if url.username or url.password or url.query or url.fragment:
            raise ValueError('invalid base URL')
        self.base, self.client_id, self.secret = base.rstrip('/'), client_id, secret
        self.token, self.expires = None, 0
        self.opener = urllib.request.build_opener(urllib.request.ProxyHandler({}), NoRedirect())

    def request(self, path, body=None, token=None):
        headers = {'Accept': 'application/json'}
        if token:
            headers['Authorization'] = 'Bearer ' + token
        data = None
        if body is not None:
            data = json.dumps(body).encode()
            headers['Content-Type'] = 'application/json'
        req = urllib.request.Request(self.base + path, data=data, headers=headers)
        with self.opener.open(req, timeout=10) as response:
            return json.load(response)

    def authenticate(self):
        result = self.request('/auth/token', {'client_id': self.client_id, 'client_secret': self.secret})
        if result.get('token_type', '').lower() != 'bearer':
            raise ValueError('unsupported token type')
        self.token = result['access_token']
        self.expires = time.monotonic() + max(0, int(result['expires_in']) - 5)

    def get(self, path):
        if not self.token or time.monotonic() >= self.expires:
            self.authenticate()
        try:
            return self.request(path, token=self.token)
        except urllib.error.HTTPError as error:
            if error.code != 401:
                raise
            self.token = None
            self.authenticate()
            return self.request(path, token=self.token)


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        raise urllib.error.HTTPError(req.full_url, code, 'redirect refused', headers, fp)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--base', default='http://127.0.0.1:8080')
    parser.add_argument('--client-id', default='peony-07')
    args = parser.parse_args()
    client = Client(args.base, args.client_id, getpass.getpass('Client secret: '))
    try:
        for path in ('/whoami', '/api/books', '/api/loans', '/api/summary'):
            print(path)
            print(json.dumps(client.get(path), ensure_ascii=False, indent=2))
    except (urllib.error.URLError, ValueError, KeyError) as error:
        parser.exit(1, f'Request failed ({type(error).__name__})\n')


if __name__ == '__main__':
    main()
