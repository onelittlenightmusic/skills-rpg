#!/usr/bin/env python3
"""rpg-try-keys: try all of chap's keys on a target door until one works.

Reads JSON arg from sys.argv[1] (or stdin if absent):
  {"target": "vault_door"}
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


def get(path: str) -> dict:
    with urllib.request.urlopen(BASE + path, timeout=15) as resp:
        return json.loads(resp.read())


def post(path: str, body: dict) -> tuple[dict, int]:
    req = urllib.request.Request(
        BASE + path,
        data=json.dumps(body).encode(),
        headers={"Content-Type": "application/json"},
    )
    try:
        with urllib.request.urlopen(req, timeout=15) as resp:
            return json.loads(resp.read()), resp.status
    except urllib.error.HTTPError as e:
        data = json.loads(e.read().decode("utf-8", "replace"))
        return data, e.code


def main() -> None:
    raw = sys.argv[1] if len(sys.argv) > 1 else sys.stdin.read()
    try:
        arg = json.loads(raw or "{}")
    except json.JSONDecodeError as e:
        error_out(f"invalid JSON arg: {e}")
        return

    target = arg.get("target", "")
    if not target:
        error_out("target door is required")
        return

    report_progress(10, "observing game state...")
    try:
        state = get("/api/v1/observe?actor=chap")
    except Exception as e:
        error_out(f"cannot reach rpg-server: {e}")
        return

    current_stage = state.get("value", {}).get("current_stage", "")
    if not current_stage:
        error_out("could not determine current stage")
        return

    items = state.get("value", {}).get("stages", {}).get(current_stage, {}).get("items", {})
    chap_keys = [item_id for item_id, item in items.items() if item.get("held_by") == "chap"]

    if not chap_keys:
        error_out(f"chap holds no keys in stage {current_stage}")
        return

    report_progress(20, f"chap holds {len(chap_keys)} key(s): {', '.join(chap_keys)}")

    total = len(chap_keys)
    for i, key_id in enumerate(chap_keys):
        progress = 20 + int(70 * i / total)
        report_progress(progress, f"trying {key_id} on {target}...")

        data, status = post("/api/v1/control", {
            "actor": "chap",
            "action": "open",
            "target": target,
            "args": {"key": key_id},
        })

        if data.get("ok"):
            report_progress(100, f"opened {target} with {key_id}")
            print(json.dumps(data, ensure_ascii=False), flush=True)
            return

        reason = data.get("reason", "rejected")
        report_progress(progress, f"{key_id}: {reason}")

        if "already open" in reason:
            report_progress(100, f"{target} is already open")
            print(json.dumps(data, ensure_ascii=False), flush=True)
            return

    error_out(f"no key opened {target}", tried=chap_keys)


if __name__ == "__main__":
    main()
