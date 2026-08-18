# Actions

Everything that changes the game goes through one endpoint, as one shape. This
is the reference for what may be asked, by whom, and what has to be true for it
to happen.

```
POST /api/v1/control
{ "actor": "you" | "chap", "action": "...", "target": "...", "args": { ... } }
```

The same call is what `mywant rpg control` sends (see
[mywant-rpg-cli.md](mywant-rpg-cli.md)) and what the MCP tool
`rpg_control_system` exposes to an agent.

## The two actors

The game is played by two, and the split is the point of it: `you` is the person
inside the dungeon, `chap` is the AI agent on the other end of an MCP cable.
Neither can do the other's half.

| | `you` | `chap` |
|---|---|---|
| `observe` | ✅ | ✅ |
| `move` | ✅ | — |
| `pickup` | ✅ | — |
| `advance` | ✅ | — |
| `return` | ✅ | — |
| `inspect` | ✅ | — |
| `name_thing` | ✅ | — |
| `pin_thing` | ✅ | — |
| `open` | — | ✅ |
| `activate` | — | ✅ |
| `deactivate` | — | ✅ |
| `state` | — | ✅ |

chap cannot walk and cannot pick anything up; you cannot open a door or touch a
device. The three fortress verbs are `you`'s for the same reason the split
exists: naming a value, putting it on the board and reading a card are not
interventions in the world, they are things the player says about it and sees in
it. chap can be *asked about* them — it just cannot do them for you. An action asked by the wrong actor is refused with
`"<actor> cannot perform \"<action>\""` — and, like every refusal, still records
an event (see [Refusals are events](#refusals-are-events)).

## The actions

### `observe`
Either actor. No state change. Recording an observe event is the whole effect —
achievements and narrations can match on it. Reading the world without recording
anything is `GET /api/v1/state` (or `mywant rpg observe [path]`).

### `move` — `you`
`target` is a waypoint id in the current stage.

Moves if the waypoint is adjacent and no closed door sits between the two;
otherwise a breadth-first search looks for a path through open waypoints and,
finding one, puts you at the destination. Refused when the target does not exist
in this stage, or when every route to it is behind a closed door
(`no open path from "a" to "b"`).

### `pickup` — `you`
`target` is an item id.

Requires the item to be at your current position and not already held. The item
moves into `you.inventory`.

### `open` — `chap`
`target` is a door id. Unlock and open folded into one, because from the
outside they are one act.

Refused unless every condition on the door is met:

| Condition | Requirement |
|---|---|
| `key` | chap holds it, or you do. Passing `args.key` names a specific key and fails if it is the wrong one — which is how a wrong-key attempt becomes an event worth narrating. |
| `requires_device` | that device is on |
| `blocked_by_device` | that device is off |

Opening an already-open door succeeds and changes nothing. That is deliberate:
achievements fire on the event, so a repeat must not fail.

### `activate` / `deactivate` — `chap`
`target` is a device id. Idempotent in the same way and for the same reason.
`activate` is refused while another device blocks it.

### `state` — `chap`
`target` is a device id. Reads `on` without changing anything.

### `advance` — `you`
Moves to `next_stage`, the stage this one leads to.

Requires the current stage to be cleared — its `cleared_when` achievement must
be held — **unless you have been in the next stage before**. Re-entering a stage
you have already walked out of is not the same act as earning it for the first
time, and without that exception, stepping back one stage would strand you: the
clear condition was met in a life the stage no longer remembers.

### `return` — `you`
Moves to the stage that leads here — `next_stage` inverted. No clear
requirement: leaving a stage unfinished and coming back to it is ordinary.

Refused with `no previous stage to return to` in the first stage.

### `inspect` — `you`

Read something without changing it: an item you are carrying, a door, a device.

```
{ "actor": "you", "action": "inspect", "target": "north_gate" }
{ "actor": "you", "action": "inspect", "target": "lira_mark" }
```

A door reports what it is waiting for, which is the point of it. `waiting_for`,
`subtype` and `satisfied` appear on a door whose condition is a named or pinned
thing — that is how the player finds out a gate has a maker field with nothing
in it, and it is not something chap can be asked, because it is a reading rather
than an act.

An item reports its `value` — what is written on it. A maker's mark carries a
name; reading it is how you learn the name, and the board still does not know it
until you say so.

### `name_thing` — `you`

Tell the city a value's name, so it exists as a **thing** and wants can refer to
it.

```
{ "actor": "you", "action": "name_thing", "target": "lira_mark",
  "args": { "subtype": "maker" } }
```

`subtype` is what kind of thing this is — a `maker`, a `station`, a `level`. The
value comes from the item named in `target`, or from `args.value` when there is
no item to read it off:

```
{ "actor": "you", "action": "name_thing",
  "args": { "subtype": "level", "value": "3.5" } }
```

Naming is not using. The value exists to the system once it has a name, whether
or not anything refers to it yet — and a want that was waiting for a value of
that kind can resolve the moment it does.

It needs a running mywant: this writes to the city's own catalog, not to the
game's state. Without one the action is refused and says so.

### `pin_thing` — `you`

Put a named value on the board **to stay**.

```
{ "actor": "you", "action": "pin_thing", "target": "lira_mark",
  "args": { "subtype": "maker" } }
```

The difference from `name_thing` is the whole of what the fortress's second
stage is about. A value is drawn on the canvas when something refers to it **or**
when it has been pinned. The first is contingent — remove the references and it
is gone from the board — and the second is not.

Pinning something the city has not been told about names it on the way, rather
than refusing: mywant has no such ordering rule and neither does this.


## Travelling does not reset

`advance` and `return` change where you are and nothing else. Both stages are
left exactly as they stand: doors as opened, devices as switched, items as
carried, achievements as earned. Only the per-stage event log is cleared, and
your position in the stage you left is remembered, so coming back puts you at
the door you went out of rather than at the entrance.

This is worth stating because there is a call that does the opposite.
`POST /api/v1/debug/jump` (`mywant rpg debug jump <stage>`) restores the stage
it lands on from its pristine YAML and empties inventory, achievements and
history. It is a testing restart, and using it to travel means every trip out
and back re-arms the alarms and re-shuts the doors. **To move between stages,
use `advance` and `return`.**

## Refusals are events

A refused action is not nothing. It records an event with `result: "rejected"`
and a reason, and achievements can match on exactly that — `attempted_self_unlock`
exists because you trying to open your own door is a thing worth noticing. So a
refusal may still advance the story, and the reason string is written for a
person to read.

## What comes back

```jsonc
{
  "ok": true,
  "actor": "you", "action": "move", "target": "control_room",
  "reason": "",                       // why not, when ok is false
  "changes": { "you.position": "control_room" },
  "achievements_unlocked": ["escaped_control_room"],
  "narration": { ... },               // if one matched this event
  "next_goal": { "text": "...", "hint": "...", "required_skill": "..." }
}
```

`changes` is a flat map of dotted paths into the state, so a caller can apply
what happened without re-reading the world. `next_goal` is recomputed on every
action — it is the answer to "what now", and it is always current.

HTTP status is `200` when the action was performed, `409` when it was refused
for a game reason, `400` for a malformed request.
