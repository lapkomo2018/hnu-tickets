import datetime
import json
import unittest
from monitor import evaluate


class MonitorTests(unittest.TestCase):
    def test_window_thresholds_and_malformed(self):
        now = datetime.datetime.now(datetime.timezone.utc)
        def row(status, duration, seconds=0):
            return json.dumps({'time': (now-datetime.timedelta(seconds=seconds)).isoformat(), 'msg': 'http.operation', 'status': status, 'duration_ms': duration})
        lines = [row(401, 1)] * 10 + [row(200, 1)] * 9 + [row(500, 3001)] * 2
        lines += [row(500, 9000, 301), '{}', '[]', '{bad']
        result = evaluate(lines, now.timestamp())
        self.assertEqual(result['requests'], 21)
        self.assertEqual(set(result['alerts']), {'access_failures', 'server_errors', 'latency'})
        self.assertEqual(result['malformed_lines'], 2)

    def test_empty(self):
        self.assertEqual(evaluate([], 0)['alerts'], [])


if __name__ == '__main__': unittest.main()
