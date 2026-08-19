# World 2 — The Fortress City

> **Status: design.** Nothing here is implemented yet. When these stages ship,
> `STAGES.md` gets the usual per-stage entries and this file stays as the record
> of why they are shaped the way they are.
>
> World 3 (the forms, and the Legacy) is [docs/world-monolith.md](world-monolith.md).
> The world-switching both need is specified at the bottom of this file.

The dungeon (`stage1`–`stage9`) teaches **MCP → Skills → Wants** through doors,
because a door is something the player already wants opened. It is played almost
entirely through Claude Code and the CLI.

The fortress teaches **the board**: things, and the roads between wants. Every
stage is one operation on the want canvas, performed by the character standing on
it, with a result the player can see without asking anyone.

Fifteen stages, because someone who has never used mywant learns exactly one new
thing at a time. The count is not padding — an earlier draft of this world fit
the same material into eight and had to assume, at four separate points, that the
player already knew something nobody had taught them.

One constraint they have to respect, learned the hard way: **a stage has to be
finishable by walking.** The canvas lays a stage's rooms in a chain and puts the
exit portal in the last one, so any door between rooms stands between the player
and the way out. An earlier draft split "read the gate's card and see what it
wants" from "give it the name so it opens" into two stages over the same gate,
which cleared in the state machine and walled the player in on the board.
`TestFortressStagesAreWalkable` now fails on any door a stage cannot open.

---

## The Story

You are out of the dungeon. The fortress city runs on the Monolith — the lights
are on, machinery turns, and every want in the city reports healthy.

Nothing works.

The dungeon already said what the Empire does to people who challenge the
Monolith: *the Empire erased them, but the tools remained.* The fortress is where
the player finds out what erasing someone actually meant.

**It meant taking their name off everything they built.**

The Keymakers signed their work. The city's gates, its pumps, its lifts were built
by people the Empire then unmade, which in practice meant scraping the maker's
name off every one of them. Breaking the gates would have invited repair. Taking
the names left them running and unaddressable: a gate whose maker has no name is a
gate nobody can call to. The roads between the wants were cut for the same reason
— everything still produces what it always produced, and there is nowhere for it
to go.

And it is not finished. **The redaction crews are still working.** A name only
lives on this board while something refers to it, so keeping a city nameless is
not an act but a routine: find the references, remove them, move on. The player
will watch it done to a name they personally restored, which is how they find out
what they are actually up against — not the ruins of an old crime, but a shift
that comes round again.

A city that cannot be accused of being broken, and cannot be used.

**Lira** was your master, a Keymaker, and she died in the dungeon. What you carried
out is her **maker's mark** — and because the Empire scraped her name off
everything she built, the mark in your pocket is the last place that name still
exists.

The dungeon taught that defining an intent makes repetition scale. The fortress is
the other half: **an intent with nothing named to act on, and no road to act
along, scales to nothing.** Tools endure and intent endures, as the Keymakers
wrote — but a tool whose maker has no name cannot be called.

`fortress15` opens the Monolith's outer shell.

---

## Act 1 — Names

### fortress1 — The Maker's Mark

**Learns**: that the board answers questions chap cannot, and that naming a value
is what puts it into the world.
**Kata**: 留 initiated.

| Item | Details |
|---|---|
| Rooms | `north_gate_approach` → `gate_arch` |
| Blocker | `north_gate` — built by Lira, and it answers to its maker's name |
| Item held | `lira_mark` — her maker's mark, carried out of the dungeon |
| Clear Condition | `passed_north_gate` |
| Next Stage | fortress2 |

Above ground for the first time, and the city's own board is visible — every want
in the district standing where it stands. The north gate is Lira's work and will
not open. Not broken; the Empire scraped her name off it, and a gate whose maker
has no name has nobody to answer to.

Asked to open it, chap does nothing at all — no lock turning, no refusal, the way
a bell rings when nobody is holding it. It can act on the world; it cannot read
what a want is waiting for. **That is a reading, and readings come from the
board** — which makes this the first thing in the game the player does rather
than asks for: walk onto the gate's tile, open its card, and find one field named
for a maker with nothing in it.

The player has been carrying the answer without knowing it is one. The mark in
their inventory is the last surviving copy of that name, and an item in your
pocket is something *you* know; the board does not read pockets.

Naming it puts the value into the world. The gate resolves and opens — and a tile
appears by itself with one road running to the gate, **because the gate is now
using it**. The player did not place it.

**Goal Steps**: observe → ask chap (nothing happens) → read the gate's card → read the mark → name the value → the gate opens → go through.

**Why one stage and not two**: reading the card and naming the value are one
continuous discovery, and an earlier draft that split them left a gate the first
stage could not open standing between the player and its own exit.

---

### fortress2 — The Name That Vanishes

**Learns**: pinning, and what it is *for*.

Through the gate — and behind you a redaction crew is already at the gate,
scraping her name back off it. **The tile winks out.** Not because anything broke
or finished: the name is gone from the only thing that was referring to it, and a
name nothing refers to is not on the board.

The player has just watched the erasure happen, at working speed, to a name they
put back five minutes ago. The Empire never destroyed Lira's name. It made her
name **contingent** — present only while something happened to refer to it — and
then it went around removing the references. It is still going around removing
them. That is not history, it is the job somebody in this city has today.

Pinning is the answer, and it is not a convenience: a pinned thing stands on the
board whether or not anything refers to it, which is precisely the dependency the
Empire has been exploiting. **Pin is the undoing of the erasure**, and the crew
can scrape that gate as often as they like.

**Goal**: walk through → watch the crew scrape the name and the tile vanish → name it again → pin it → they scrape it again and it stays.

> **Verified** (against a running mywant): the thing↔want relation is derived on
> read — a want names a value while one of its subtyped parameters *holds* that
> value. Want status is not part of it. A want that has finished its work
> (`achieved`) still holds the edge; the edge disappears when the parameter stops
> holding the value, or when the want is deleted. So the redaction crew's move is
> to **clear the gate's maker parameter** — an earlier draft had the tile vanish
> when the gate's work finished, which does not happen and would have been odd if
> it did.
>
> What this needs: an `rpg_redactor` behaviour that blanks a named parameter on a
> named want, so a stage can stage the erasure on cue.

---

### fortress3 — The Third Road

**Learns**: reading a thing tile — it shows who uses it, and a name reaches
further than the one thing you gave it to.

The cistern is draining and nothing here was touched. The pinned tile, which had
one road a minute ago, now has **three** — the gate, and two wants the player has
never seen.

The gate was not the only thing Lira built. Putting her name back did not open one
door; it woke everything she ever signed, and one of them is pulling the cistern
down.

**Goal**: the cistern is draining → look at the tile: three roads → follow them → find the want draining it → stop it.

---

### fortress4 — The Name You Typed

**Learns**: things are usually made *implicitly* — type a value into a want's
field and it is remembered.

So far naming has been a deliberate act on an heirloom. That is the rare case. A
sluice here needs a level and the player simply types one in — and afterwards the
value is in the catalog, with a tile, without anyone having decided to make a
thing.

Left untaught, a player comes away believing the catalog is a manual archive they
must curate. It is mostly a record of what they have already typed.

**Goal**: set the sluice by hand → look at the board → the value you typed is standing there.

---

### fortress5 — Kinds of Name

**Learns**: a thing has a **kind** (subtype), and kinds are what wants accept.

Two values, both correct, both refused by the other's fitting: a level is not a
maker, a maker is not a level. Nothing here is about spelling — the values are
different *kinds of thing*, and every want declares which kinds it takes.

**Goal**: try the wrong value in the wrong place → read what each asks for → see the kinds on the tiles → match them.

**Why before seeding**: `fortress6` shows a list narrowing itself. Without this
stage the player sees the result and not the reason.

---

### fortress6 — The Gate Nobody Built

**Learns**: seeding a want from a thing.

The relay gate has no want behind it — Lira did not live to build this one. The
player has to, and the interesting part is where they start.

Starting from the **pinned thing** rather than an empty form: the type list is
filtered to types that can accept this kind of value, and the new want is born
with the value already in the right parameter. Wrong choices are not rejected
later; they are absent.

**Goal**: the gate has no want → Add Want *from the tile* → notice the narrowed list → deploy → the gate opens.

---

### fortress7 — The Ward With No Address

**Learns**: grouping. One name for many things.

Eight of Lira's works stand in the ward, each with its own named value already on
the board — visible because each is in use, exactly as `fortress2` established.
The ward ledger will not take them one at a time. It takes a **district**.

Saying these eight belong together draws the lines that join them, and gives chap
an address it can accept. It is the first time the player names a **relationship**
rather than a value.

**Goal**: the ledger rejects buildings one by one → group the eight → the lines appear → file the district order.

---

## Act 2 — Reading a want

The player has deployed wants since `stage7` and has never once looked at one on
the board. Act 3 is unreadable without this.

### fortress8 — Green and Red

**Learns**: a want tile is readable — it has a state, and the state is not the
same as "doing something useful".

A row of Lira's wants, some running, some stopped, one failed. The player learns
to tell them apart by looking, and — more importantly — that **green does not mean
working**. That single fact is what makes `fortress11` land.

**Goal**: find the failed one among the running ones → read why → restart it.

---

### fortress9 — What It Has Made

**Learns**: a want holds results — what it has actually produced.

Something is producing the wrong pressure and the only way to know is to look at
what it has been putting out, not at whether it is running.

**Goal**: read the produced values → find the one that is wrong → correct its input.

---

### fortress10 — Mouths

**Learns**: a want has **ends** — what it offers, and what it needs.

The Keymakers' wants have openings on them, and until now the player has read them
as decoration. This is the stage that says what they are: one side offers a value,
the other asks for one, and they are the only places a road can attach.

**Goal**: walk the workshop → read the offers and the needs → say, without connecting anything, which two could be joined.

---

## Act 3 — Roads

### fortress11 — Everything Green, Nothing Moving

**Learns**: connecting two wants.

One want produces exactly what the door needs. The door is two metres away. Both
are green. Nothing happens.

**This stage betrays the dungeon's lesson on purpose.** The dungeon trained the
player that when something does not work, you deploy a want. Here that changes
nothing, twice, before the real move becomes visible: the two are not related, and
nobody has said they are. Connecting them draws a road, in the same visual
language as the roads a thing draws to its wants.

**Goal**: both green → deploy something (no effect) → connect the two ends → the road appears → the door powers.

---

### fortress12 — The Wrong Floor

**Learns**: disconnecting and rewiring. A connection can be wrong without being an
error.

Several wants offer output; the lift takes one input. Connect the wrong one and
**nothing errors** — the lift runs, arrives, opens onto the wrong floor, and the
door behind you locks. Plausible and wrong teaches better than rejected.

**Goal**: connect (plausible, wrong) → arrive on the wrong floor → read the ends again → disconnect → connect the matching pair.

---

### fortress13 — The Value Nobody Makes

**Learns**: chains. Three wants in a row, deriving something none of them holds.

The assay door asks for a value that exists nowhere in the city. One want produces,
a second transforms, a third delivers.

**Goal**: connect A→B → connect B→C → the derived value reaches the door.

---

### fortress14 — The Name in the Middle

**Learns**: pinning something *you* produced. Act 1 and Act 3 on one board.

The value in the middle of that chain is worth having again, so the player pins it
— and this time the thing being named is not an heirloom but something they made.
A chain of roads with a named tile standing in the middle of it.

**Goal**: pin the intermediate value → it stands in the chain → use it to seed the next want.

---

### fortress15 — Rewiring the District

**Learns**: the two halves at scale. The finale.

The eight buildings from `fortress7`, each needing its own named value, all fed
from one supply. The player wires the first two by hand and the tedium is
deliberate — the same tedium `stage7` used to make wants worth having, now in this
world's material.

The answer is the dungeon's answer applied to what Act 1 built: deploy one want
whose destinations are **the group**. The names stop being labels and become the
list the intent fans out across. The district lights building by building, the
board fills with roads, and the Monolith's outer shell opens.

**Goal**: wire two by hand → notice the shape repeating → deploy one want across the group → the district lights → the shell opens.

> There is a shorter way to do this stage, and world 3 is where the player is
> given it. Doing it by hand first is what makes that shortcut mean anything.

---

## How a stage hears about mywant

The doors read mywant, and until they were told when to look they only ever
found out by being asked — a condition checked at the moment somebody tried a
door. Enough to unlock one, and nothing to narrate with. The moment worth
showing is when the player names a value and a tile appears with a road running
to the gate, and a game that learns it later cannot say so.

mywant streams `thing_changed` on /api/v1/events, so the game listens (see
`server/mywant_sse.go`). On each one it re-reads the doors of the stage in play
and records an ordinary event for any whose condition has **changed**:

| | |
|---|---|
| `city` / `answered` | a condition that was unmet is met |
| `city` / `forgot` | one that was met no longer is |

`city` because neither `you` nor `chap` did it. Both directions, because a name
can leave — which is not a non-event but the Empire's whole method, performed on
a name the player restored, and the beat `fortress2` exists for.

The watcher **opens the doors it watches**, and moves nobody and clears nothing.
Opening them is the point rather than an overreach: in the dungeon chap opens
doors because `you` has no privilege to, and a fortress gate is not like that —
it answers to its maker, and what was missing was the name. Making the player
name a value and then go and ask chap to try the door adds a step that teaches
nothing and dilutes the one thing this world is about. Naming is the act; the
gate opening is its consequence.

Which means an achievement for a door opening matches the **city's** event, not
chap's. Keying one to an act nobody performs any more is an achievement that
never fires and a goal that waits forever — and a goal for a beat that can no
longer happen (chap being refused, once the name is already in the city) needs
`unless:` so it stands down instead of asking for the impossible.

**Every stage from here uses this**: a board condition worth narrating gets a
`match: { actor: city, action: answered|forgot, target: <door> }` narration and
needs nothing else.

The baseline is taken on arrival, not on the first event to reach us — otherwise
the first change after entering a stage is spent working out what "before" was,
and the player's opening move goes unremarked.

---

## What has to exist

Every operation in Act 1 and Act 3 rests on affordances that already ship:

| Stage | Rests on |
|---|---|
| fortress1, 4 | naming a value explicitly, and implicitly by typing it into a want |
| fortress2 | pinning; the drawing rule (pinned / placed / referred to by a want) |
| fortress3 | a thing tile draws a road to every want naming it |
| fortress5 | subtypes: the catalog a value is filed under comes from the field it was typed into |
| fortress6 | Add Want from a thing: filtered type list, seeded parameter |
| fortress7 | thing groups (mywant calls them constellations), drawn as lines between members |
| fortress8–10 | want state, produced values, expose/import ends |
| fortress11–14 | want↔want connection, drawn as a road |
| fortress15 | one want addressing many named values |

The scenery is wants, because everything on a mywant canvas is. `rpg_fixture`
(`server/skills/rpg-fixture/`) is the one type they are all made of: a `maker`
field subtyped `person`, a `level` field subtyped `level`, an `about` list of
thing ids, and no agent at all. It does nothing on purpose — a fixture is read,
named, grouped, connected and stopped, and every one of those is an operation on
the want rather than something the want performs. A stage puts its fixtures on
the board with `mywant_setup.wants`.

## Three rules the first two stages settled

Written down because each was arrived at the hard way, and every stage from
here obeys them.

**A door is a door.** It says `requires_device` and nothing else — the same
sentence stage4's power_door says. A door that also weighed up the state of the
city was a second mechanism for something world 1 already had a shape for, and
it showed: doors only ever knew about the current stage, so a player standing at
one a room over (the boards put every stage on one canvas) could satisfy it and
watch nothing happen.

**An activator holds the condition, and says what is missing.** The device is
what reads the city, and it is running exactly while the city holds what it
watches for. Its card is the only thing in a position to tell the player what to
do, so every condition answers three questions and not one:

```go
satisfied() bool   // is it running
checks() []Check   // the parts of the question, each true or false on its own
need() string      // the one job outstanding right now, in the imperative
```

`checks()` and `need()` live in Go next to the condition (`server/device_cond.go`),
not in the skill that draws the card: the thing that owns the fact answers for
it. The device's JSON carries `on`, `checks` and `need`; `rpg-device/main.py`
relays them and `view/plugin.jsx` renders a tick list and a sentence, neither of
them knowing what kind of condition produced it. Adding a kind is one struct and
three methods — no Python, no JSX, no want-type schema change.

Two things `need()` has to get right, both learned from fortress1 and 2:

- name the job that is **actually outstanding**, not the one the condition
  mentions. A value the city has never heard of cannot be pinned; saying "pin
  it" sends somebody hunting for a button that does not exist. A chain condition
  names the *first* missing hop and not all of them;
- say it as an instruction, not as the rule behind it. "Lira only stays on the
  board while something uses it" is true and useless to somebody who has never
  pinned anything. The stage may add a `need_hint:` — the concrete how-to that
  only the stage knows ("the sluice in this gallery has a level field") — which
  is appended to the generated statement of what.

**The RPG server reads mywant and never writes on the player's behalf.** Naming
a value, pinning it, connecting two wants are mywant's own operations, done in
mywant — by the CLI, the Thing page, the canvas. An RPG verb wrapping the same
call taught a player something that works nowhere else. It follows that a stage
cannot require its own route: whatever route the player takes, the activator
notices. (`mywant_setup` does write, but it writes scenery on stage entry, not
the player's move.)

### The conditions, as they ended up

Every one is a fact about mywant, and every one is something the player can see
on the board:

```yaml
devices:
  registry_terminal:
    reads_thing: { subtype: person, value: "Lira" }               # named
  ward_registry:
    reads_thing: { subtype: person, value: "Lira", pinned: true } # and pinned
  level_register:
    reads_thing: { subtype: level }                     # any value of that kind
  relay_meter:
    reads_thing_used_by: { subtype: person, value: "Kesh", min_users: 2 }
  district_ledger:
    reads_group: { name: district_a, min_members: 8 }
  row_monitor:
    reads_want: { type: rpg_fixture, state: suspended, absent: true }
  pressure_assay:
    reads_want_state: { name: pressure_stack_b, field: level, equals: "9" }
  road_meter:
    reads_connection: { from: head_tank, to: district_main }
  chain_reader:
    reads_connection: { chain: [ore_face, smelter, assay_feed] }
  crosstalk_check:
    reads_connection: { from: spare_winch, to: lift_carriage, absent: true }
  district_meter:
    reads_connection: { from: district_supply, min_count: 8 }
```

The board interface (`server/rules_things.go`) is the whole of what the game
knows about the city, and it is deliberately small — `CountNamed`, `Pinned`,
`UsedBy`, `GroupMembers`, `CountWants`, `WantState`, `Connected`,
`ConnectionsFrom`. Each is answered by asking mywant (`server/board_mywant.go`)
rather than computed here, so the line a door waits for and the line the player
sees are the same line.

The door in front of each device says `requires_device`, the SSE watcher notices
when any activator's answer changes and narrates it, and the door opens because
its device came alive. The plumbing: the RPG server holds a mywant client
(`server/mywant_client.go`), subscribes to `/api/v1/events`
(`server/mywant_sse.go`), and receives lifecycle webhooks at
`/api/v1/hooks/mywant`.

## A stage's spine, written once

Every stage is the same shape — a line of rooms, a door between each pair, a
reader powering each door, and the player walking to the last room. Written
longhand that is a waypoint block with the adjacency spelled both ways, a door
block repeating the two room ids, an achievement per reader, an achievement per
room, and a `cleared_when` naming the last of them: thirty-odd lines in which
the only interesting words are the ids.

So a stage writes its rooms instead (`server/stage_rooms.go`):

```yaml
rooms:
  - { id: cistern_floor, label: "The Cistern Floor" }
  - { id: cistern_stair, label: "The Stair Out",
      door: { id: cistern_hatch, device: cistern_gauge } }
```

and the loader derives the waypoints, the doors, the achievements
`<device>_running` and `reached_<room>`, the initial position and the
`cleared_when`. Those two ids are part of the contract, because goals and
narrations refer to them. Anything the stage writes by hand wins, and a stage
whose shape this cannot describe — a loop, a room off to one side, a door that
starts open — writes `waypoints:` and `doors:` longhand as fortress1 and 2 still
do.

`cmd/stage-to-world` expands the same shorthand before laying a stage out, so a
stage using `rooms:` is a stage with a board.

---

## Worlds (plumbing)

The player never learns this; it is save-file machinery, and it is here because
neither of the two new worlds can exist without it.

`GameState` grows a world layer, and everything per-progress moves into it:
`CurrentWorld` plus `Worlds map[string]*WorldState`, where a `WorldState` holds
what `GameState` holds today (`CurrentStage`, `You`, `Chap`, `Stages`,
`Achievements`, `StageExit`). Stage YAMLs move to `server/stages/<world>/`, with a
flat `stages/*.yaml` still read as world `dungeon` so existing saves keep working.
Each world directory carries a `world.yaml` naming its title, first stage, unlock
achievement, and the mywant world it corresponds to.

Three endpoints, mirroring mywant's verbs: `GET /api/v1/worlds`,
`POST /api/v1/worlds/switch` (`{world, stage?}`), `POST /api/v1/worlds/reset`
(`{world}`).

Two notes worth keeping:

- **There is no save verb.** mywant's `world open` snapshots before switching
  because its worlds share one live canvas. Here each world's progress is a
  separate subtree the others never write to, so a switch cannot lose anything.
- **`unlocked_by` gates the offer, not the switch.** An explicit switch always
  works, like `debug/jump`, because demoing a world you have not earned is normal.

**Prerequisite**: `cmd/stage-to-world` currently flattens every stage into a single
mywant world called `skills-rpg`. It needs a `-world` filter and a default name of
`skills-rpg-<world>` before any world can be linked to a mywant world of its own.
