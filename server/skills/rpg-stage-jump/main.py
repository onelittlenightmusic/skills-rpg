#!/usr/bin/env python3
"""
Stage travel trigger monitor.

Watches a want's "pending_jump" state field — set to a stage id by a card
button, cleared back to "" once handled — and walks the game there through
rpg-server's ordinary control endpoint, then clears the field so it fires once
per press.

Serves both directions, because travelling is travelling: `next_stage` sends
"advance" ("進む"), `startpoint` sends "return" ("前の階に戻る"). The want type
declares which way it faces; this only has to carry it.

Deliberately NOT /api/v1/debug/jump. That endpoint restores the stage it lands
on from its pristine YAML and empties inventory, achievements and history — a
testing restart. Used for travel it meant a trip out and back re-armed every
alarm and re-shut every door, so a stage forgot you had ever been in it. The
normal actions leave both stages exactly as you left them.

Args (JSON):
  pending_jump  - the stage id the button asked for, or "" when idle (no-op)
  direction     - "advance" (default) or "return"

Always prints a JSON object (never errors out the poll): {} when idle,
otherwise {"pending_jump": "", "last_jump_result": "..."} once handled —
cleared even on refusal, so an uncleared stage or a down rpg-server does not
retry forever. The reason is recorded for the card to show.
"""

import json
import os
import sys
import urllib.error
import urllib.request

RPG_SERVER_URL = os.environ.get("RPG_SERVER_URL", "http://127.0.0.1:7100")


def clean(val) -> str:
    v = str(val).strip()
    return "" if (v.startswith("%{") and v.endswith("}")) else v


def get(path: str) -> dict:
    with urllib.request.urlopen(f"{RPG_SERVER_URL}{path}", timeout=5) as resp:
        return json.loads(resp.read() or b"{}")


def previous_of(stages: dict, stage_id: str) -> str:
    """Who leads into stage_id — the next_stage links, inverted."""
    for sid, st in (stages or {}).items():
        if isinstance(st, dict) and st.get("next_stage") == stage_id:
            return sid
    return ""


def post(path: str, payload: dict) -> dict:
    req = urllib.request.Request(
        f"{RPG_SERVER_URL}{path}",
        data=json.dumps(payload).encode(),
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=5) as resp:
            return json.loads(resp.read() or b"{}")
    except urllib.error.HTTPError as e:
        # The refusals worth showing ("current stage is not cleared yet") come
        # back as 4xx with a body, so read it rather than losing it to a raise.
        try:
            return json.loads(e.read() or b"{}")
        except Exception:
            return {"ok": False, "reason": f"HTTP {e.code}"}


def parse_arg(raw) -> dict:
    if isinstance(raw, dict):
        return raw
    try:
        val = json.loads(raw)
        return val if isinstance(val, dict) else {}
    except (TypeError, json.JSONDecodeError):
        return {}


def travel(arg: dict) -> dict:
    pending = clean(arg.get("pending_jump", ""))
    if not pending:
        return {}

    direction = clean(arg.get("direction", "")) or "advance"
    action = "return" if direction == "return" else "advance"

    # The card names a stage; the action only names a direction, so the two have
    # to be reconciled against where the game actually is before moving it.
    #
    # A press can reach this more than once — the poll runs again before the
    # cleared pending_jump has landed — and with a relative action a second run
    # is not the harmless repeat it was with an absolute jump: it would advance
    # a stage further than asked. So the target is checked first, and only a
    # move that lands exactly where the card asked is made at all.
    try:
        state = get("/api/v1/state")
        current = state.get("current_stage", "")
        if current == pending:
            result = "ok"                      # already there; nothing to do
        else:
            stages = state.get("stages") or {}
            here = stages.get(current) or {}
            target = here.get("next_stage", "") if action == "advance" else previous_of(stages, current)
            if target != pending:
                result = f"{pending} is not the stage {action} leads to from {current}"
            else:
                res = post("/api/v1/control", {"actor": "you", "action": action})
                result = "ok" if res.get("ok") else (res.get("reason") or res.get("error") or "refused")
    except Exception as e:
        result = f"error: {e}"

    return {"pending_jump": "", "last_jump_result": result}


def serve() -> None:
    """Answer jobs on stdin until it closes (MYWANT_MRS_SERVE=1).

    The same bargain every rpg skill makes: mywant keeps one process alive
    rather than paying ~57ms of interpreter startup per want per tick. A script
    that only reads argv is not merely slower here — under serve it is handed
    no argv at all, answers "nothing to do" and exits, so the button appears to
    work (the field clears) while the game never moves.
    """
    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue
        req = parse_arg(line)
        args = req.get("args") or []
        out = travel(parse_arg(args[0]) if args else {})
        out["_id"] = req.get("_id")
        print(json.dumps(out, ensure_ascii=False), flush=True)


def main() -> None:
    arg = parse_arg(sys.argv[1]) if len(sys.argv) > 1 else {}
    print(json.dumps(travel(arg), ensure_ascii=False), flush=True)


if __name__ == "__main__":
    if os.environ.get("MYWANT_MRS_SERVE") == "1":
        serve()
    else:
        main()
