"""Reproduce equivalence checks and trace Appendix A's state machine."""
import contextlib
import io
import json
import platform
import runpy
import subprocess
import sys
from pathlib import Path
from unittest.mock import patch

ROOT = Path(__file__).resolve().parent
KEY = '2026_ІПЗс-24-1_Лапко_Максим_варіант:7_'


def execute(name, value):
    run = subprocess.run([sys.executable, str(ROOT / name)], input=value + '\n',
                         capture_output=True, text=True, timeout=30, check=True)
    return run.stdout.removeprefix('Введіть пароль: ').strip()


def trace(value):
    states = []
    filename = str(ROOT / 'appendix_obfuscated.py')

    def hook(frame, event, arg):
        if event == 'line' and frame.f_code.co_filename == filename and '_w' in frame.f_locals:
            pair = [frame.f_locals['_s'], frame.f_locals['_w']]
            # Record the dispatch line, after _w has been recomputed.
            if lines[frame.f_lineno - 1].strip() == 'match _w:':
                states.append(pair)
        return hook

    lines = Path(filename).read_text().splitlines()
    previous = sys.gettrace()
    try:
        with patch('builtins.input', return_value=value), contextlib.redirect_stdout(io.StringIO()):
            sys.settrace(hook)
            runpy.run_path(filename, run_name='__main__')
    finally:
        sys.settrace(previous)
    return states


def main():
    appendix = []
    for value in ['Secret123', 'wrong', '']:
        outputs = [execute(name, value) for name in
                   ['appendix_original.py', 'appendix_hash.py', 'appendix_obfuscated.py']]
        assert len(set(outputs)) == 1
        assert outputs[0] == ('Доступ дозволено' if value == 'Secret123' else 'Доступ заборонено')
        appendix.append({'input': value, 'outputs': outputs})
    cases = [KEY, '', 'wrong', KEY[:-1], KEY + ' ', KEY.lower(), KEY.replace(':7_', ':07_'), KEY + 'x', 'я' * 1000, 'я' * 40000]
    own = []
    for value in cases:
        outputs = [execute(name, value) for name in ['original.py', 'obfuscated.py']]
        assert outputs[0] == outputs[1]
        assert outputs[0] == ('Доступ дозволено' if value == KEY else 'Доступ заборонено')
        own.append({'input': value if len(value) < 200 else f'<{len(value)} Cyrillic characters>', 'outputs': outputs})
    result = {'python': platform.python_version(), 'key': KEY, 'appendix': appendix,
              'own': own, 'trace_success': trace('Secret123'), 'trace_failure': trace('wrong')}
    (ROOT / 'results.json').write_text(json.dumps(result, ensure_ascii=False, indent=2) + '\n')
    print('29 executions passed; success and failure traces recorded')


if __name__ == '__main__':
    main()
