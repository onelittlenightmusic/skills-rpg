#!/usr/bin/env python3
"""rpg-world: keep skills-rpg in the world this mywant world is for.

A want of this type lives in a generated mywant world and names the skills-rpg
world it belongs to. Opening that world in mywant brings the want with it, and
the want puts the game where the board already is.

It reconciles rather than fires once. mywant's `world open` restores a want with
the state it was saved in, so a one-shot doer that had already finished would
come back finished and never run again — and a want that only acts on creation
could not fix a drift that happened afterwards. Comparing on every poll has
neither problem, and costs one GET when the two already agree, which is almost
always.

Reads JSON arg from sys.argv[1] (or a job on stdin under serve mode):
  {"world": "fortress"}
"""
import json
import os
import sys
import urllib.error
import urllib.request

RPG_SERVER_URL = os.environ.get("RPG_SERVER_URL", "http://127.0.0.1:7100").rstrip("/")


def clean(val) -> str:
    """An unexpanded %{placeholder} means the want has no value for this yet."""
    v = str(val).strip()
    return "" if (v.startswith("%{") and v.endswith("}")) else v


def get_json(path: str):
    with urllib.request.urlopen(f"{RPG_SERVER_URL}{path}", timeout=5) as resp:
        return json.loads(resp.read().decode())


def post_json(path: str, payload: dict):
    req = urllib.request.Request(
        f"{RPG_SERVER_URL}{path}",
        data=json.dumps(payload).encode(),
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    with urllib.request.urlopen(req, timeout=10) as resp:
        return json.loads(resp.read().decode())


def observe(arg: dict) -> dict:
    want_world = clean(arg.get("world", ""))
    out = {
        "world": want_world,
        "current_world": "",
        "in_sync": False,
        "switched": False,
        "error": "",
        "summary": "",
    }
    if not want_world:
        out["error"] = "no world named"
        out["summary"] = "no world named"
        return out

    try:
        worlds = get_json("/api/v1/worlds")
    except Exception as e:
        # rpg-server unreachable. Say nothing rather than claiming a mismatch:
        # switching on a guess would move a game nobody asked to move.
        out["error"] = f"rpg-server unreachable: {e}"
        out["summary"] = "waiting for rpg-server"
        return out

    current = str(worlds.get("current") or "")
    out["current_world"] = current
    title = ""
    for w in worlds.get("worlds") or []:
        if w.get("id") == want_world:
            title = str(w.get("title") or "")
    out["title"] = title

    if current == want_world:
        out["in_sync"] = True
        out["summary"] = f"in {title or want_world}"
        return out

    try:
        post_json("/api/v1/worlds/switch", {"world": want_world})
    except Exception as e:
        out["error"] = f"switch failed: {e}"
        out["summary"] = f"could not enter {want_world}"
        return out

    out["in_sync"] = True
    out["switched"] = True
    out["summary"] = f"entered {title or want_world}"
    return out


def parse_arg(raw: str) -> dict:
    try:
        return json.loads(raw)
    except (json.JSONDecodeError, TypeError):
        return {}


def serve() -> None:
    """Answer jobs on stdin until it closes (MYWANT_MRS_SERVE=1)."""
    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue
        req = parse_arg(line)
        args = req.get("args") or []
        out = observe(parse_arg(args[0]) if args else {})
        out["_id"] = req.get("_id")
        print(json.dumps(out, ensure_ascii=False), flush=True)


def main() -> None:
    arg = parse_arg(sys.argv[1]) if len(sys.argv) > 1 else {}
    print(json.dumps(observe(arg), ensure_ascii=False), flush=True)


if __name__ == "__main__":
    if os.environ.get("MYWANT_MRS_SERVE") == "1":
        serve()
    else:
        main()
