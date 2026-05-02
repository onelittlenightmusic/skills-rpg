#!/usr/bin/env python3
"""rpg-load: restore rpg-server state from a save slot."""
import json
import os
import sys
import urllib.error
import urllib.request


def error_out(msg: str) -> None:
    print(json.dumps({"error": msg, "ok": False}, ensure_ascii=False), flush=True)
    sys.exit(1)


def main() -> None:
    raw = sys.argv[1] if len(sys.argv) > 1 else "{}"
    try:
        arg = json.loads(raw or "{}")
    except json.JSONDecodeError as e:
        error_out(f"invalid JSON arg: {e}")
        return

    slot = arg.get("slot", "").strip()
    if not slot:
        error_out("slot is required")
        return

    base = os.environ.get("RPG_SERVER_URL", "http://localhost:7100").rstrip("/")
    url = f"{base}/api/v1/saves/{slot}/load"
    req = urllib.request.Request(url, data=b"", method="POST",
                                  headers={"Content-Type": "application/json"})
    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            data = json.loads(resp.read())
    except urllib.error.HTTPError as e:
        error_out(f"HTTP {e.code}: {e.read().decode('utf-8','replace')}")
        return
    except urllib.error.URLError as e:
        error_out(f"cannot reach rpg-server: {e}")
        return

    print(json.dumps(data, ensure_ascii=False), flush=True)


if __name__ == "__main__":
    main()
