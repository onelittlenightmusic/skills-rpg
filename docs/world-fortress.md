# World 2 — The Fortress City

> **Status: design.** Nothing here is implemented yet. When these stages ship,
> `STAGES.md` gets the usual per-stage entries and this file stays as the record
> of why they are shaped the way they are.

The dungeon (`stage1`–`stage9`) teaches **MCP → Skills → Wants**: how to ask, how
to automate the asking, how to define an intent that scales. It teaches through
doors, because a door is something the player already wants opened.

The fortress teaches the board. Every stage is **one operation on the want
canvas**, performed by the character standing on it, with a result the player can
see without asking anyone:

| | Operation | What appears on the board |
|---|---|---|
| 1 | **Pin** a thing | the value becomes a tile, at the character's feet |
| 2 | **Read** the tile | roads to every want that uses it, and how often |
| 3 | **Group** things | lines joining them; one name for the set |
| 4 | **Seed** a want from a thing | the type list narrows; the value is already in the field |
| 5 | **Connect** two wants | a road appears, and something starts moving |
| 6 | **Match** the right ends | what each end speaks decides what fits |
| 7 | **Chain**, name the middle | the derived value becomes a tile of its own |
| 8 | **Apply a kata** to the board | the form names what is missing, and fills it |

Nothing is learned by being told to run a command. Each stage is a place the
player is stuck, the operation is the way out, and the board shows what changed.

---

## Kata — the spine

The operations above are exactly what **kata (型)** are made of. A kata is a named
combination of **waza (所作)**, and from 黄帯 upward every kata is measured against
a **thing group** (`join: { kind: thing_group }`) — the group is what makes its
waza be about the same thing. Grouping is not a side topic; it is the connective
tissue of the whole ladder.

So the fortress does not teach kata as a subject. It **is** a kata practice, and
the player finds that out the way a student does: by noticing that chap needs less
telling than it did an hour ago.

Kata grant no features — every want type is usable from day one. What they grant
is 手数 (shorthand), 権限 (delegation) and 語彙 (vocabulary), which in this story
reads as chap getting better at understanding you. That is the reward the player
feels, and the record book (the kata page) is where they go to see why.

### The fortress kata

A set of its own, shipped with the world, named in the same register as the白帯
forms (駅・市・標・催・金):

| Kata | Waza | Yields | Stage |
|---|---|---|---|
| **留** (tome) — to fix in place | `thing` pinned | a value you can walk to | fortress1–2 |
| **群** (mure) — the flock | a thing group with members | one name for many | fortress3 |
| **種** (tane) — the seed | `want_type` joined to a pinned thing | a want born knowing its value | fortress4 |
| **道** (michi) — the road | a connection between two wants | what one makes reaches the other | fortress5–6 |
| **束** (taba) — the sheaf | 道 repeated across a group | one intent, every destination | fortress7–8 |

Each is held at 初伝 the first time and at 皆伝 with repetition, which is why
several stages revisit a form rather than each introducing a new one. `束` is a
`repeat` waza over `道`, so it cannot be reached without having drawn roads by
hand first — the ladder enforces the order the stages teach in.

**What chap gains, in story terms:**

| Form | 手数 / 権限 / 語彙 |
|---|---|
| 留 | chap can be told a name instead of a description |
| 群 | chap accepts "the district" as an address |
| 種 | chap proposes only mechanisms that can take the value you named |
| 道 | chap reports what is connected to what, unasked |
| 束 | chap wires a whole group on one instruction |

The last one is the world's thesis: **you stop giving instructions per object.**

### make Kata — a form you can apply

A kata that only reports progress is a scoreboard. The point of holding a form is
that you can *use* it, and the fortress ends by making that literal.

**Entry point.** The character action bubble — the one that opens when you press
Start standing on empty ground, where Call, Ride, Drop, Say and Map live — gains
**Kata**. That menu is the right home because it is the menu that acts on *you*:
applying a form is something the player does, not something done to a card.

**The gesture.**

1. Start on open ground → the bubble → **Kata**
2. A picker of the forms you **hold** — an unlearned form has nothing to lend, so
   only 初伝 and above are offered
3. Pick one, e.g. 束
4. The board answers with **what is missing**: this form wants a road from the
   supply to each member of the district, you have two, here are the other six —
   drawn on the board as proposals, not as facts
5. Accept, and they are made

**What the player understands at that moment**: a kata is not a badge for work
already done. It is a description of a shape, and the board can be compared
against it. Having practised 留・群・種・道 by hand is what makes the comparison
mean anything — the form knows what "the district" is because the player grouped
it, and knows what a road is because the player drew some.

**Backend.** Most of this already exists. Kata progress is derived server-side on
every read, and each waza already comes back with `satisfied`, `have`, `need`,
`matchedIDs` and a one-phrase `hint`. What is missing is the step from "this waza
is unsatisfied" to "here is the specific thing to make on *this* board":

```
POST /api/v1/kata/{id}/recommend   →  { proposals: [ ... ] }
```

Each proposal is executable rather than descriptive — a want to create with its
type and seed, a connection to draw between two named ends, a group to extend —
so the frontend can offer it and, on acceptance, carry it out. The matching work
(which pinned things, which existing wants, which group) is the same matching the
progress derivation already does; this endpoint returns the *complement* of it
instead of the count.

**Frontend.** A `Kata` item and picker stage in the character action bubble
(alongside `pick-call` / `pick-ride`), and the proposals rendered on the canvas in
the idiom the app already uses for recommendations — proposed, reviewable,
accepted or dismissed — rather than applied silently.

**Scope note**: this is the one part of the fortress that needs work on both
sides. Everything else in Act 1 and Act 2 rests on affordances that already ship.

---

## The Story

You are out of the dungeon. The fortress city runs on the Monolith — the lights
are on, machinery turns, and every mechanism reports healthy.

Nothing works.

The Empire did not sabotage the city. Breaking things invites repair. It did two
quieter things: it **took the names off everything**, so no instruction can find
its object, and it **cut the roads between the mechanisms**, so everything still
produces what it always produced and there is nowhere for it to go.

The result is a city that cannot be accused of being broken and cannot be used.

**Lira**, a Keymaker archivist, kept a ledger through all of it. She is the reason
the old names can be recovered — and the reason the player learns that a name is
not a sticker but a record of what a value has been used for. She is also the one
who recognises the forms: it is Lira who tells the player, the first time chap
answers a shorter instruction correctly, that this is what the Keymakers called a
kata.

The dungeon taught that defining an intent makes repetition scale. The fortress
is the other half: **an intent with nothing named to act on, and no road to act
along, scales to nothing.**

`fortress8` opens the Monolith's outer shell. The Legacy is inside — world 3.

---

## Act 1 — Putting things on the board

### fortress1 — The Plate at Your Feet

**Learning Theme**: pinning. A value that is not on the board is not in the world.
**Kata**: 留 initiated.

| Item | Details |
|---|---|
| Room | `gatehouse` |
| Blocker | `north_gate` — its mechanism asks for a destination and has none |
| Clear Condition | `first_thing_pinned` |
| Next Stage | fortress2 |

The gate's mechanism sits green and idle with an empty destination field. The
player is carrying the answer — a name scratched on a plate taken from the
dungeon — but it is text, and text is nowhere.

Naming it makes it a thing. **Pinning it puts it on the board at the character's
own cell**, which is where the lesson lands: the value stops being something the
player knows and becomes something the world contains, standing on the floor
beside them.

**Goal Steps**: observe → the gate names what it lacks → name the value → pin it → the tile appears underfoot → the gate resolves.

---

### fortress2 — Three Identical Valves

**Learning Theme**: reading a thing tile. The board answers "what uses this".
**Kata**: 留 to 皆伝 (a second and third value pinned along the way).

| Item | Details |
|---|---|
| Room | `pump_hall` |
| Blocker | `sluice_door` — opens from whichever valve actually feeds it |
| Clear Condition | `traced_by_board` |
| Next Stage | fortress3 |

Three mechanisms, identical to look at, one of them fed by the value from
`fortress1`. Asking chap gets a shrug: it can act, it cannot know which of three
indistinguishable things you meant.

The pinned thing answers it. A thing tile **draws a road to every want that names
it**, and carries its use count and the icons of the want types using it. The
player queries nothing — they walk to the tile and see which road leaves it.

**Goal Steps**: observe → chap cannot disambiguate → walk to the pinned thing → follow its road → operate that valve.

---

### fortress3 — The Ward With No Address

**Learning Theme**: grouping. One name for many things.
**Kata**: 群.

| Item | Details |
|---|---|
| Room | `ward_office` |
| Blocker | `ward_ledger` — will accept a work order for a district, not for a building |
| Clear Condition | `group_formed` |
| Next Stage | fortress4 |

Eight buildings, each now named and pinned, and a ledger that refuses every one of
them: it does not take addresses one at a time, it takes a **district**. Naming
each building again in a list is possible and pointedly tedious.

The move is to say that these eight belong together. On the board that draws the
lines that join them; to chap it becomes an address it can accept. This is the
first time the player names something that is not a value but a **relationship**.

It is also the stage everything above 白帯 quietly depends on — `join:
{ kind: thing_group }` is how every later kata knows its waza are about the same
thing — so it is placed before the first want is built rather than after.

**Goal Steps**: observe → the ledger rejects single buildings → group the pinned things → the lines appear → file the district order.

---

### fortress4 — The Gate That Has No Mechanism

**Learning Theme**: seeding a want from a thing. The name shapes what you can build.
**Kata**: 種.

| Item | Details |
|---|---|
| Room | `east_approach` |
| Blocker | `relay_gate` — needs a mechanism that does not exist yet |
| Clear Condition | `seeded_from_thing` |
| Next Stage | fortress5 |

Nothing here can open the gate, because the mechanism for it was never built. The
player has to make one, and the interesting part is where they start.

Starting from the **pinned thing** rather than an empty form does two visible
things: the type list is **filtered to the types that can accept this kind of
value**, and the new want is born with the value already in the right parameter.
Wrong choices are not rejected later — they are absent.

**Goal Steps**: observe → the gate has no mechanism → Add Want *from the pinned thing* → notice the narrowed list → deploy → the gate opens.

**Why here**: it closes Act 1 by answering what a name is worth — it is worth the
thing you build next.

---

## Act 2 — Roads between mechanisms

### fortress5 — Everything Green, Nothing Moving

**Learning Theme**: connection. A healthy want connected to nothing does nothing.
**Kata**: 道 initiated.

| Item | Details |
|---|---|
| Room | `keymaker_workshop` |
| Blocker | `pressure_door` — needs supply produced two metres away |
| Clear Condition | `first_road_drawn` |
| Next Stage | fortress6 |

The Keymakers' machinery, intact, still running. One mechanism produces exactly
what the door needs. The door is two metres away. Both report healthy. Nothing
happens.

**This stage betrays the dungeon's lesson on purpose.** The dungeon trained the
player that when something does not work, you deploy a want. Here that changes
nothing, twice, before the real move becomes visible: the two are not related and
nobody has said they are. Connecting them **draws a road**, in the same visual
language as the roads a thing draws to its wants — so the player already knows how
to read it.

**Goal Steps**: observe → both green → deploy something (no effect) → connect the two → the road appears → the door powers.

---

### fortress6 — The Wrong Floor

**Learning Theme**: a connection is between *ends*, not between wants. Disconnecting
is a normal move.
**Kata**: 道 to 皆伝.

| Item | Details |
|---|---|
| Room | `lift_lobby` |
| Blocker | `lift` — arrives wherever it was told to |
| Clear Condition | `rematched` |
| Next Stage | fortress7 |

Several mechanisms offer output; the lift takes one input. Connect the wrong one
and **nothing errors** — the lift runs, arrives, and opens onto the wrong floor,
where the door behind you locks. Plausible and wrong teaches better than rejected.

Getting out means reading what each end actually speaks, disconnecting, and
connecting the pair that matches.

**Goal Steps**: observe → connect (plausible, wrong) → arrive on the wrong floor → read the ends → disconnect → connect the matching pair.

---

### fortress7 — The Value Nobody Makes

**Learning Theme**: chains, and naming what comes out of the middle. Both acts on
one board.
**Kata**: 道 held; 留 applied to something the player produced.

| Item | Details |
|---|---|
| Room | `assay_room` |
| Blocker | `derived_door` — wants a value no single mechanism produces |
| Clear Condition | `chain_named` |
| Next Stage | fortress8 |

The door asks for something that exists nowhere in the city. It has to be derived:
one mechanism produces, a second transforms, a third delivers — three roads in a
row.

Then the move that ties the world together. The value in the middle of that chain
is worth having again, so the player **pins it**. A chain of roads with a named
tile standing in the middle of it, and from here that derived value can seed a
want like any other name — which is exactly what the last stage needs.

**Goal Steps**: observe → connect A→B → connect B→C → pin the intermediate value → the door opens, and the new tile stands in the chain.

---

### fortress8 — Rewiring the District

**Learning Theme**: what all of it is *for*. One intent, every named destination.
**Kata**: 束 — and with it the belt.

| Item | Details |
|---|---|
| Room | `district_ward` |
| Blocker | `shell_gate` — the Monolith's outer shell |
| Clear Condition | `city_relit` |
| Next Stage | — (opens world 3) |

The district from `fortress3`: eight buildings, each needing its own named value,
all fed from one supply. The player wires the first two by hand and the tedium is
deliberate — the same tedium `stage7` used to make wants worth having, now in this
world's material.

The answer is the dungeon's answer applied to what Act 1 built: **the form knows
this shape already.** Start on open ground, pick **Kata → 束**, and the board
answers with the six roads it is missing — because the district is a group the
player made, and a road is something the player has drawn by hand five times. The
proposals appear, the player accepts, and the district lights building by
building.

This is where kata stops being a progress bar. The player has spent seven stages
performing forms; here a form performs for them, and it can only do so because
everything it needs to match against was named by hand first.

束 is a `repeat` waza over 道, so the ladder only offers it to a player who has
drawn roads themselves. Reaching it is the belt, and the belt is what opens the
shell.

**Goal Steps**: wire two by hand → notice the shape repeating → Start → Kata → 束 → the missing roads are proposed → accept → the district lights → the shell opens.

---

## What has to exist

Most of it is already built in the GUI. The stages are arranged around affordances
that work today:

| Stage | Rests on |
|---|---|
| fortress1 | pinning a thing places it at the character's cell |
| fortress2 | a thing tile draws a road to every want naming it, and shows use count and want-type icons |
| fortress3 | thing groups, drawn on the board as lines joining their members |
| fortress4 | Add Want from a thing filters types by subtype and seeds the matching parameter |
| fortress5–7 | want↔want connection, drawn as a road |
| fortress8 | **make Kata** — see below; the one genuinely new feature |

Four things need adding.

**make Kata**, described in full above: a `Kata` item in the character action
bubble, a picker of held forms, `POST /api/v1/kata/{id}/recommend` returning
executable proposals, and those proposals rendered on the canvas. Both sides.

**A waza kind for connections.** Kata waza today are `want_type`, `thing` and
`repeat`. 道 and 束 need a fourth that is satisfied by a connection existing
between two wants. Everything else the fortress kata need — including the group
axis — is already expressible, since `join: { kind: thing_group }` is how the
existing belts bind their waza together.

**Door predicates for the new operations**, in the shape `requires_device` /
`blocked_by_device` already established:

```yaml
doors:
  north_gate:
    requires_thing_pinned:    { subtype: destination }
  ward_ledger:
    requires_thing_group:     { min_members: 8 }
  relay_gate:
    requires_want_seeded_from:{ thing_subtype: destination }
  pressure_door:
    requires_connection:      { from: supply, to: sink }
```

**A way to see kata advance in-game.** The record book exists as a page; what the
fortress needs is for a form reaching 初伝 to be a *moment* — Lira naming it, chap
answering a shorter instruction — rather than a number changing on a screen the
player may never open. This matters more once make Kata exists: a form the player
never noticed themselves earning is a tool they will never think to reach for.

The plumbing for all of this is in place: the RPG server holds a mywant client
(`server/mywant_client.go`) and receives lifecycle webhooks
(`/api/v1/hooks/mywant`), which is how a stage sees a want created, a thing
pinned, or a connection made.

---

## Worlds (plumbing)

The player never learns this; it is save-file machinery, and it is here only
because a second world cannot exist without it.

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
`skills-rpg-<world>` before either world can be linked to a mywant world of its own.
