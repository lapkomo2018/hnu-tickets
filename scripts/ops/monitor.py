"""Evaluate one UTC time window of JSON request logs and readiness."""
import argparse
import datetime
import json
import urllib.request


def evaluate(lines, now, window=300):
    rows = []
    malformed = 0
    for line in lines:
        try:
            row = json.loads(line)
            if not isinstance(row, dict):
                raise ValueError('object required')
            if row.get('msg') != 'http.operation':
                continue
            if not isinstance(row.get('status'), int) or not isinstance(row.get('duration_ms'), (int, float)):
                raise ValueError('invalid request record')
            stamp = datetime.datetime.fromisoformat(row['time'].replace('Z', '+00:00')).timestamp()
            if row.get('msg') == 'http.operation' and 0 <= now - stamp <= window:
                rows.append(row)
        except (ValueError, KeyError, TypeError):
            malformed += 1
    failures = sum(r['status'] >= 500 for r in rows)
    denied = sum(r['status'] in (401, 403, 429) for r in rows)
    durations = sorted(r['duration_ms'] for r in rows)
    p95 = durations[max(0, (len(durations) * 95 + 99) // 100 - 1)] if durations else 0
    alerts = []
    if denied >= 10: alerts.append('access_failures')
    if len(rows) >= 20 and failures / len(rows) >= .05: alerts.append('server_errors')
    if p95 > 2000: alerts.append('latency')
    return dict(requests=len(rows), server_errors=failures, access_failures=denied,
                p95_ms=p95, malformed_lines=malformed, alerts=alerts)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('log')
    parser.add_argument('--ready', default='http://127.0.0.1:8080/ready')
    args = parser.parse_args()
    with open(args.log) as stream:
        result = evaluate(stream, datetime.datetime.now(datetime.timezone.utc).timestamp())
    try:
        opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))
        with opener.open(args.ready, timeout=3) as response:
            healthy = response.status == 200 and json.load(response).get('status') == 'ready'
    except (OSError, ValueError):
        healthy = False
    result['ready'] = healthy
    if not healthy: result['alerts'].append('not_ready')
    print(json.dumps(result))
    raise SystemExit(bool(result['alerts']))


if __name__ == '__main__': main()
