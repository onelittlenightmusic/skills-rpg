#!/usr/bin/env python3
"""rpg-control: send a control action to rpg-server as `chap`.

Reads JSON arg from sys.argv[1] (or stdin if absent):
  {"action": "open", "target": "door1", "args": {...}}
"""
import json
import os
import sys
import urllib.error
import urllib.request


def report_progress(p: int, m: str = "") -> None:
    print(json.dumps({"_progress": p, "_message": m}, ensure_ascii=False), flush=True)


def error_out(msg: str, **extra) -> None:
    payload = {"error": msg, "ok": False, **extra}
    print(json.dumps(payload, ensure_ascii=False), flush=True)
    sys.exit(1)


def main() -> None:
    raw = sys.argv[1] if len(sys.argv) > 1 else sys.stdin.read()
    try:
        arg = json.loads(raw or "{}")
    except json.JSONDecodeError as e:
        error_out(f"invalid JSON arg: {e}")
        return

    action = arg.get("action", "")
    target = arg.get("target", "")
    if not action:
        error_out("action is required")
        return

    body = {"actor": "chap", "action": action, "target": target}
    if "args" in arg and arg["args"]:
        body["args"] = arg["args"]

    base = os.environ.get("RPG_SERVER_URL", "http://localhost:7100").rstrip("/")
    url = f"{base}/api/v1/control"

    report_progress(20, f"chap → {action} {target}")
    req = urllib.request.Request(
        url,
        data=json.dumps(body).encode(),
        headers={"Content-Type": "application/json"},
    )
    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            data = json.loads(resp.read())
    except urllib.error.HTTPError as e:
        body_text = e.read().decode("utf-8", "replace")
        try:
            data = json.loads(body_text)
        except Exception:
            error_out(f"HTTP {e.code}: {body_text}")
            return
        # control-level rejection: still emit JSON (with ok=false) so the
        # mywant want can pick up reason/hints/next_goal as fields.
        report_progress(100, "rejected")
        print(json.dumps(data, ensure_ascii=False), flush=True)
        return
    except urllib.error.URLError as e:
        error_out(f"cannot reach rpg-server at {url}: {e.reason}")
        return

    report_progress(100, "done")
    print(json.dumps(data, ensure_ascii=False), flush=True)


if __name__ == "__main__":
    main()
