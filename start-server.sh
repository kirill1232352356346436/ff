#!/usr/bin/env bash
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")"
if [ ! -x server/.venv/bin/python ]; then
  python3 -m venv server/.venv
fi
server/.venv/bin/python -m pip install -r server/requirements.txt
exec server/.venv/bin/python -m uvicorn server.main:app --host 127.0.0.1 --port 8010
