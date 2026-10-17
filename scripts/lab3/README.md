# Laboratory 3 password experiments

Requirements: Python 3.12, Linux libxcrypt (for the traditional bcrypt demonstration), and the pinned Python bcrypt package.

```sh
python -m pip install -r scripts/lab3/requirements.txt
python scripts/lab3/run_experiments.py
```

The run creates `results.json` and `dictionary.json`. Random salts and the ten-character suffix change on each execution; timings depend on the machine. The committed files record one completed run.

To verify an existing laboratory hash with the 20-entry dictionary:

```sh
python scripts/lab3/dictionary_sha256.py '<hash5>' scripts/lab3/dictionary.json
python scripts/lab3/dictionary_bcrypt.py '<hash6>' scripts/lab3/dictionary.json
```

Each script takes its algorithm from the script name, receives the reference hash and dictionary, visits candidates in order and stops on a match. The output contains the matched candidate, number of attempts and elapsed microseconds. Initial hashing, imports and JSON loading are outside the measured search.

`run_experiments.py` supplies the complete 72/73/82/71-byte inputs to libxcrypt and verifies all 16 pairs. It also records that bcrypt 5 rejects oversized inputs. The HNU Tickets application rejects passwords longer than 72 UTF-8 bytes rather than truncating them.
