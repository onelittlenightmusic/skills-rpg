#!/usr/bin/env python3
"""rpg-observe: read a subtree of rpg-server state."""
import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request


def report_progress(p: int, m: str = "") -> None:
    print(json.dumps({"_progress": p, "_message": m}, ensure_ascii=False), flush=True)


def error_out(msg: str, **extra) -> None:
    print(json.dumps({"error": msg, **extra}, ensure_ascii=False), flush=True)
    sys.exit(1)


def main() -> None:
    raw = sys.argv[1] if len(sys.argv) > 1 else "{}"
    try:
        arg = json.loads(raw or "{}")
    except json.JSONDecodeError as e:
        error_out(f"invalid JSON arg: {e}")
        return

    target = arg.get("target", "")
    base = os.environ.get("RPG_SERVER_URL", "http://localhost:7100").rstrip("/")
    qs = {"actor": "chap"}
    if target:
        qs["target"] = target
    url = f"{base}/api/v1/observe?{urllib.parse.urlencode(qs)}"

    report_progress(20, f"observing {target or '<root>'}")
    try:
        with urllib.request.urlopen(url, timeout=15) as resp:
            data = json.loads(resp.read())
    except urllib.error.URLError as e:
        error_out(f"cannot reach rpg-server: {e}")
        return

    report_progress(100, "done")
    print(json.dumps(data, ensure_ascii=False), flush=True)


if __name__ == "__main__":
    main()
