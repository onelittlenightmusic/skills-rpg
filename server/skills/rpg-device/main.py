#!/usr/bin/env python3
"""
Device <-> rpg-server monitor (rpg_generator / rpg_alarm).

Polls one skills-rpg device (GET /api/v1/state on the running rpg-server,
default port 7100, no auth) and mirrors it into the want's state, together with
the doors that device gates — a generator is only interesting because of the
door it powers, so the door travels with it.

The want may also hold an intent: "desired" says which way the device should be,
and chap is sent to make it so (POST /api/v1/control). This is a level, not an
edge — the want states what it wants and this script closes the gap whenever it
can. That matters because the want's fetchFrom fields are re-expanded from the
last output on every reconcile cycle, so a one-shot request written by the card
and cleared by this script could never survive; a standing intent can.

Who writes what, so the two never fight:
  desired            - the card (webhook). Never written here.
  attempted / error  - this script. Never written by the card.

When to act, so the game's event log does not fill with refusals:
  * desired differs from the intent last acted on -> a fresh instruction: act
    (and forget any earlier error).
  * same intent, previous attempt errored -> stay put; clicking again is what
    asks for a retry.
  * same intent, no error, but the world has drifted -> re-assert it.
  * blocked by another device -> never call: it would only be refused. The
    block lifting is itself a drift, so the intent lands on its own.

Args (JSON):
  stage_id    - skills-rpg stage id (e.g. "stage4")
  device_id   - device id within that stage (e.g. "generator")
  desired     - "on" | "off" | "" (optional; "" = observe only)
  attempted   - the intent this script last acted on (echoed back by the want)
  last_error  - the error from that attempt, if any (echoed back by the want)

stage_id or device_id missing/empty -> prints {} and exits (no-op).
"""

import json
import os
import sys
import time
import urllib.error
import urllib.request

RPG_SERVER_URL = os.environ.get("RPG_SERVER_URL", "http://127.0.0.1:7100").rstrip("/")


def clean(val) -> str:
    """An unexpanded %{placeholder} means "the want has no value for this yet"."""
    v = str(val).strip()
    return "" if (v.startswith("%{") and v.endswith("}")) else v


# One snapshot serves every device. Under `serve` mode this process answers all
# the generators and alarms of a world, and they were each issuing the same
# argument-independent GET. The TTL sits just under the poll interval so a
# device still sees a change on the tick after it happens.
STATE_TTL_SECONDS = 1.5
_snapshot = {"at": 0.0, "data": None}


def read_state(force: bool = False):
    """The whole game state. Raises if rpg-server cannot be reached.

    force=True skips the cache — required after control(), where the point of
    the read is to observe the change we just caused. It also refreshes the
    shared snapshot, so the other devices see that change immediately.
    """
    now = time.monotonic()
    if not force and _snapshot["data"] is not None and now - _snapshot["at"] < STATE_TTL_SECONDS:
        return _snapshot["data"]
    with urllib.request.urlopen(f"{RPG_SERVER_URL}/api/v1/state", timeout=5) as resp:
        _snapshot["data"] = json.loads(resp.read().decode())
        _snapshot["at"] = now
    return _snapshot["data"]


def control(action: str, target: str) -> str:
    """chap performs an action. Returns "" on success, else the reason."""
    body = json.dumps({"actor": "chap", "action": action, "target": target}).encode()
    req = urllib.request.Request(
        f"{RPG_SERVER_URL}/api/v1/control",
        data=body,
        headers={"Content-Type": "application/json"},
    )
    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            data = json.loads(resp.read())
    except urllib.error.HTTPError as e:
        try:
            data = json.loads(e.read().decode("utf-8", "replace"))
        except Exception:
            return f"HTTP {e.code}"
    except Exception as e:
        return f"cannot reach rpg-server: {e}"
    if data.get("ok"):
        return ""
    return data.get("reason") or "rejected"


def survey(state, stage_id, device_id):
    """The device, and every door it gates, as the card needs them."""
    stage = state.get("stages", {}).get(stage_id, {})
    device = stage.get("devices", {}).get(device_id)
    if device is None:
        return None

    devices = stage.get("devices", {})
    doors = stage.get("doors", {})

    blocked_by = device.get("blocked_by_device", "") or ""
    blocked = bool(devices.get(blocked_by, {}).get("on")) if blocked_by else False

    powers, blocks, gates = [], [], []
    for door_id, door in sorted(doors.items()):
        gated = False
        if door.get("requires_device") == device_id:
            powers.append(door_id)
            gated = True
        if door.get("blocked_by_device") == device_id:
            blocks.append(door_id)
            gated = True
        if gated:
            gates.append({
                "id": door_id,
                "open": bool(door.get("open")),
                "locked": bool(door.get("locked", True)),
            })

    return {
        "on": bool(device.get("on")),
        "label": device.get("label") or device_id,
        "blocked_by": blocked_by,
        "blocked": blocked,
        "powers": powers,
        "blocks": blocks,
        "gates": gates,
        "stage_title": stage.get("title", ""),
    }


def observe(arg: dict) -> dict:
    stage_id = clean(arg.get("stage_id", ""))
    device_id = clean(arg.get("device_id", ""))
    desired = clean(arg.get("desired", "")).lower()
    attempted = clean(arg.get("attempted", "")).lower()
    last_error = clean(arg.get("last_error", ""))

    if not stage_id or not device_id:
        return {}

    try:
        state = read_state()
    except Exception as e:
        # Unreachable: say nothing about the device rather than guess at it.
        return {"error": f"cannot reach rpg-server: {e}"}

    view = survey(state, stage_id, device_id)
    if view is None:
        return {
            "error": f'unknown device "{device_id}" in "{stage_id}"',
            "attempted": desired,
        }

    fresh_intent = desired != attempted
    error = "" if fresh_intent else last_error

    act = (
        desired in ("on", "off")
        and desired != ("on" if view["on"] else "off")
        and not view["blocked"]
        and not (not fresh_intent and last_error)
    )

    if act:
        error = control("activate" if desired == "on" else "deactivate", device_id)
        try:
            view = survey(state := read_state(force=True), stage_id, device_id) or view
        except Exception:
            pass  # The act landed; the next poll will read the result.

    if error:
        summary = f"{desired}: {error}"
    elif view["blocked"]:
        summary = f'{view["label"]} is held by {view["blocked_by"]}'
    else:
        summary = f'{view["label"]} is {"on" if view["on"] else "off"}'

    return {
        **view,
        "attempted": desired,
        "error": error,
        "summary": summary,
    }


def parse_arg(raw) -> dict:
    try:
        return json.loads(raw)
    except (json.JSONDecodeError, TypeError):
        return {}


def serve() -> None:
    """Answer jobs on stdin until it closes (MYWANT_MRS_SERVE=1).

    Starting an interpreter and importing urllib costs ~57ms; the observation
    itself costs a fraction of a millisecond. Staying alive is the whole point.
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
