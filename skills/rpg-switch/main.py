#!/usr/bin/env python3
"""rpg-switch: activate or deactivate a device based on on/off state.

Reads JSON arg from sys.argv[1] (or stdin if absent):
  {"target": "generator", "on": true}   -> activate
  {"target": "generator", "on": false}  -> deactivate
"""
import json
import os
import sys
import urllib.error
import urllib.request

BASE = os.environ.get("RPG_SERVER_URL", "http://localhost:7100").rstrip("/")


def report_progress(p: int, m: str = "") -> None:
    print(json.dumps({"_progress": p, "_message": m}, ensure_ascii=False), flush=True)


def error_out(msg: str, **extra) -> None:
    print(json.dumps({"error": msg, "ok": False, **extra}, ensure_ascii=False), flush=True)
    sys.exit(1)


def main() -> None:
    raw = sys.argv[1] if len(sys.argv) > 1 else sys.stdin.read()
    try:
        arg = json.loads(raw or "{}")
    except json.JSONDecodeError as e:
        error_out(f"invalid JSON arg: {e}")
        return

    target = arg.get("target", "")
    on = arg.get("on")

    if not target:
        error_out("target device is required")
        return

    if on is None:
        # State query mode
        action = "state"
        report_progress(50, f"chap → check state of {target}")
    else:
        # Control mode
        action = "activate" if on else "deactivate"
        report_progress(20, f"chap → {action} {target}")

    body = json.dumps({"actor": "chap", "action": action, "target": target}).encode()
    req = urllib.request.Request(
        f"{BASE}/api/v1/control",
        data=body,
        headers={"Content-Type": "application/json"},
    )
    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            data = json.loads(resp.read())
    except urllib.error.HTTPError as e:
        data = json.loads(e.read().decode("utf-8", "replace"))
        report_progress(100, "rejected")
        print(json.dumps(data, ensure_ascii=False), flush=True)
        return
    except urllib.error.URLError as e:
        error_out(f"cannot reach rpg-server: {e.reason}")
        return

    report_progress(100, "done")
    print(json.dumps(data, ensure_ascii=False), flush=True)


if __name__ == "__main__":
    main()
