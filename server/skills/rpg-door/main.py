#!/usr/bin/env python3
"""
Door <-> rpg-server monitor (rpg_door).

Mirrors one skills-rpg door (GET /api/v1/state on the running rpg-server,
default port 7100, no auth) into the want's state.

The predecessor of this script mirrored "locked" alone, which is all the canvas
tile needs to decide whether CursorMan may walk through. A card has room for the
rest of the answer — which key opens it, whether chap is carrying that key, and
which device is holding it — so all of it is read here. Everything is derived
from the same single GET, so the extra fields cost nothing.

Args (JSON):
  stage_id  - skills-rpg stage id (e.g. "stage4")
  door_id   - door id within that stage (e.g. "power_door")

Either missing/empty -> prints {} and exits immediately (no-op: an unsynced
door keeps whatever "locked" its own params gave it).
"""

import json
import os
import sys
import time
import urllib.request

RPG_SERVER_URL = os.environ.get("RPG_SERVER_URL", "http://127.0.0.1:7100").rstrip("/")


def clean(val) -> str:
    """An unexpanded %{placeholder} means "the want has no value for this yet"."""
    v = str(val).strip()
    return "" if (v.startswith("%{") and v.endswith("}")) else v


# One snapshot serves every door. Under `serve` mode this process answers all
# 17 doors of a world, and they were each fetching the same argument-independent
# GET /api/v1/state — 17 identical requests every two seconds. The TTL is set
# just under the poll interval so a door still sees a change on the tick after
# it happens.
STATE_TTL_SECONDS = 1.5
_snapshot = {"at": 0.0, "data": None}


def read_state():
    """The whole game state, refetched at most once per STATE_TTL_SECONDS.

    Returns None when rpg-server cannot be reached, which callers must treat as
    "say nothing" rather than "everything is false" — see observe().
    """
    now = time.monotonic()
    if _snapshot["data"] is not None and now - _snapshot["at"] < STATE_TTL_SECONDS:
        return _snapshot["data"]
    try:
        with urllib.request.urlopen(f"{RPG_SERVER_URL}/api/v1/state", timeout=5) as resp:
            _snapshot["data"] = json.loads(resp.read().decode())
            _snapshot["at"] = now
    except Exception:
        _snapshot["data"] = None
    return _snapshot["data"]


def observe(arg: dict) -> dict:
    stage_id = clean(arg.get("stage_id", ""))
    door_id = clean(arg.get("door_id", ""))

    if not stage_id or not door_id:
        return {}

    state = read_state()
    if state is None:
        # rpg-server not running / unreachable — leave every mirrored field
        # untouched rather than clobbering "locked" with a guess. A door that
        # forgot it was locked would let CursorMan walk through a locked door.
        return {}

    stage = state.get("stages", {}).get(stage_id, {})
    door = stage.get("doors", {}).get(door_id)
    if door is None:
        return {}

    waypoints = stage.get("waypoints", {})
    devices = stage.get("devices", {})
    items = stage.get("items", {})

    def label_of(waypoint_id):
        return waypoints.get(waypoint_id, {}).get("label") or waypoint_id

    between = door.get("between") or ["", ""]
    key = door.get("key", "") or ""
    requires = door.get("requires_device", "") or ""
    blocked_by = door.get("blocked_by_device", "") or ""


    locked = bool(door.get("locked", True))
    is_open = bool(door.get("open"))

    # Why it will not open yet, in the order the game itself checks. The card
    # shows the first unmet condition, which is the one worth acting on.
    device_on = bool(devices.get(requires, {}).get("on")) if requires else False
    blocker_on = bool(devices.get(blocked_by, {}).get("on")) if blocked_by else False

    if is_open:
        summary = f"{door_id} is open"
    elif requires and not device_on:
        summary = f"{door_id} needs {requires} running"
    elif blocked_by and blocker_on:
        summary = f"{door_id} is blocked by {blocked_by}"
    elif locked:
        summary = f"{door_id} is locked" + (f" ({key})" if key else "")
    else:
        summary = f"{door_id} is unlocked"

    return {
        "locked": locked,
        "open": is_open,
        "key": key,
        "key_held_by_chap": items.get(key, {}).get("held_by") == "chap" if key else False,
        "requires_device": requires,
        "device_on": device_on,
        "device_label": devices.get(requires, {}).get("label", "") if requires else "",
        "blocked_by_device": blocked_by,
        "blocker_on": blocker_on,
        "blocker_label": devices.get(blocked_by, {}).get("label", "") if blocked_by else "",
        "between_from": label_of(between[0] if len(between) > 0 else ""),
        "between_to": label_of(between[1] if len(between) > 1 else ""),
        "stage_title": stage.get("title", ""),
        "summary": summary,
    }


def parse_arg(raw: str) -> dict:
    try:
        return json.loads(raw)
    except (json.JSONDecodeError, TypeError):
        return {}


def serve() -> None:
    """Answer jobs on stdin until it closes (MYWANT_MRS_SERVE=1).

    Starting an interpreter and importing urllib costs ~57ms; the observation
    itself costs 0.11ms. Staying alive is the whole point — the request/response
    loop below is what lets mywant stop paying the former per door per tick.
    """
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
