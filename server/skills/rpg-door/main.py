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
import urllib.request

RPG_SERVER_URL = os.environ.get("RPG_SERVER_URL", "http://127.0.0.1:7100").rstrip("/")


def clean(val) -> str:
    """An unexpanded %{placeholder} means "the want has no value for this yet"."""
    v = str(val).strip()
    return "" if (v.startswith("%{") and v.endswith("}")) else v


def main() -> None:
    arg = {}
    if len(sys.argv) > 1:
        try:
            arg = json.loads(sys.argv[1])
        except json.JSONDecodeError:
            pass

    stage_id = clean(arg.get("stage_id", ""))
    door_id = clean(arg.get("door_id", ""))

    if not stage_id or not door_id:
        print(json.dumps({}), flush=True)
        return

    try:
        with urllib.request.urlopen(f"{RPG_SERVER_URL}/api/v1/state", timeout=5) as resp:
            state = json.loads(resp.read().decode())
    except Exception:
        # rpg-server not running / unreachable — leave every mirrored field
        # untouched rather than clobbering "locked" with a guess. A door that
        # forgot it was locked would let CursorMan walk through a locked door.
        print(json.dumps({}), flush=True)
        return

    stage = state.get("stages", {}).get(stage_id, {})
    door = stage.get("doors", {}).get(door_id)
    if door is None:
        print(json.dumps({}), flush=True)
        return

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

    print(json.dumps({
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
    }, ensure_ascii=False), flush=True)


if __name__ == "__main__":
    main()
