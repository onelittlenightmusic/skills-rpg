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
import urllib.error
import urllib.request

RPG_SERVER_URL = os.environ.get("RPG_SERVER_URL", "http://127.0.0.1:7100").rstrip("/")


def clean(val) -> str:
    """An unexpanded %{placeholder} means "the want has no value for this yet"."""
    v = str(val).strip()
    return "" if (v.startswith("%{") and v.endswith("}")) else v


def read_state():
    with urllib.request.urlopen(f"{RPG_SERVER_URL}/api/v1/state", timeout=5) as resp:
        return json.loads(resp.read().decode())


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


def main() -> None:
    arg = {}
    if len(sys.argv) > 1:
        try:
            arg = json.loads(sys.argv[1])
        except json.JSONDecodeError:
            pass

    stage_id = clean(arg.get("stage_id", ""))
    device_id = clean(arg.get("device_id", ""))
    desired = clean(arg.get("desired", "")).lower()
    attempted = clean(arg.get("attempted", "")).lower()
    last_error = clean(arg.get("last_error", ""))

    if not stage_id or not device_id:
        print(json.dumps({}), flush=True)
        return

    try:
        state = read_state()
    except Exception as e:
        # Unreachable: say nothing about the device rather than guess at it.
        print(json.dumps({"error": f"cannot reach rpg-server: {e}"}, ensure_ascii=False), flush=True)
        return

    view = survey(state, stage_id, device_id)
    if view is None:
        print(json.dumps({
            "error": f'unknown device "{device_id}" in "{stage_id}"',
            "attempted": desired,
        }, ensure_ascii=False), flush=True)
        return

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
            view = survey(state := read_state(), stage_id, device_id) or view
        except Exception:
            pass  # The act landed; the next poll will read the result.

    if error:
        summary = f"{desired}: {error}"
    elif view["blocked"]:
        summary = f'{view["label"]} is held by {view["blocked_by"]}'
    else:
        summary = f'{view["label"]} is {"on" if view["on"] else "off"}'

    print(json.dumps({
        **view,
        "attempted": desired,
        "error": error,
        "summary": summary,
    }, ensure_ascii=False), flush=True)


if __name__ == "__main__":
    main()
