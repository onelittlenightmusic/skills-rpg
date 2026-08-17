# World 3 — The Monolith

> **Status: design.** Nothing here is implemented yet.
>
> World 2 (things and connections) is [docs/world-fortress.md](world-fortress.md),
> which also specifies the world-switching both worlds need.

The dungeon taught how to ask, how to automate the asking, and how to define an
intent that scales. The fortress taught the board: naming values, and the roads
between wants. Twenty-five stages of doing things by hand.

The Monolith is where the player finds out that the doing had a shape, that the
shape has a name, and that the shape can be applied.

**Kata (型).** And with it, the Legacy.

---

## The reveal

The Legacy has been the goal since the first line of `STORY.md`: a treasure hidden
deep in the Monolith, and *whoever obtains it gains the power to save many people*.
The Empire will not release it, because **those who are saved stop obeying**.

It is not a device.

The Legacy is the **practice** — the accumulated forms of every Keymaker who ever
worked on the Monolith, recorded and kept. Its power to save people is exactly the
power the player has been assembling by hand for two worlds: name a thing, draw a
road, and a broken rule stops binding anyone. The Empire hoarded it because a
person who holds the forms does not need permission to fix anything.

Which means taking the Legacy is not picking something up. **It is being able to
perform the forms.** The player already can — they have been doing it for
twenty-five stages without the word. World 3 gives them the word, shows them the
record, and hands them the one thing they have never had: the ability to point a
form at a board and have it answer.

---

## What the player has to end up understanding

Kata has more moving parts than either earlier concept, which is why it gets a
world rather than the two stages an earlier draft gave it:

| | |
|---|---|
| A form exists, and you already hold some | you have been practising without the word |
| A form is made of **waza (所作)** | the parts, and which of them you satisfy |
| Repetition deepens it — 初伝 → 皆伝 | holding a form once is not holding it well |
| Forms are grouped into **belts**, and belts promote | there is a ladder, and where you stand on it |
| A held form changes what chap needs told | 手数 / 権限 / 語彙, demonstrated rather than claimed |
| A form can be **applied** to a board | make Kata — the payoff |

The fifth of these is the one an earlier draft asserted in prose and never made
the player check. Assertion is not learning.

---

## Stages

### monolith1 — The Hall of Forms

**Learns**: kata exists, and you hold some already.

Inside the shell is a hall of records, and one of the records is **yours** — kept
by the Monolith itself, which has been watching every gate you opened and every
road you drew since the dungeon. The player has never seen a kata page. What they
see first is not a tutorial but a transcript of things they personally did.

**Goal**: enter → read the record → find 留 standing at 初伝, under your own name.

**Why first, and why it is a reading stage**: the player must recognise their own
history before the word can mean anything. A form introduced as a game mechanic is
a chore; a form introduced as *the name for what you already did* is a promotion.

---

### monolith2 — The Parts of a Form

**Learns**: a form is a combination of **waza**, and you can see which you satisfy.

One form on the wall is unfinished, with its parts listed. Some are ticked and some
are not, and the ticked ones are things the player did in the fortress.

This is the stage that makes make Kata legible later: a form is not a badge, it is
**a list of conditions**, and a board can be checked against a list.

**Goal**: open one form → read its waza → find which one you are missing → satisfy it here.

---

### monolith3 — Once Is Not Enough

**Learns**: 初伝 → 皆伝. Repetition deepens a form.

A door in the hall answers only to a form held at 皆伝. The player holds it at 初伝.
Nothing is wrong with what they did — they have simply only done it once, and the
Keymakers' word for that is 初伝.

**Goal**: find why the door refuses → perform the same form twice more → the door answers.

---

### monolith4 — The Belt

**Learns**: forms are grouped, and clearing enough of a group promotes you.

The hall is tiered. What is on the next tier cannot be reached by holding one form
well; it needs a *set* of forms, which is what a belt is. The player sees the
ladder they have been climbing without knowing it existed, and where they stand.

**Goal**: read the tier → find which forms of the belt you are missing → hold enough of them → the tier opens.

---

### monolith5 — Fewer Words

**Learns**: what a held form actually buys — 手数, 権限, 語彙 — by feeling it.

The player is asked to do something they did in the fortress, under time pressure,
with chap. The instruction they would have had to give then is long. The one they
give now is short, and chap does the right thing — because of what they hold.

Then the same task with a form they have *not* got, and the long instruction is
back.

**Goal**: do the task with a held form → do it again outside one → notice which one cost more words.

**Why a stage of its own**: this is the entire value proposition of kata, and an
earlier draft delivered it as a sentence in the story. The player has to be the one
who notices.

---

### monolith6 — Pointing a Form at a Board

**Learns**: **make Kata**. A form can be applied, and it answers with what is
missing.

The player stands in front of a district of the Monolith's own machinery, in the
state the fortress's last stage started in: things named, wants standing, and the
roads not drawn. They already know how to fix it by hand, because they did exactly
this in `fortress16`.

Instead: Start on open ground → the character bubble → **Kata** → pick 束.

The board answers with what the form is missing — the roads it wants, drawn as
proposals, not facts. Accept, and they are made.

**What lands**: a form is a description of a shape, a board can be compared against
it, and the difference is a list of things to do. It reads as magic only if you
have not spent twenty-five stages assembling the parts by hand; having done so, it
reads as *the point*.

**Goal**: recognise the shape from fortress16 → Kata → 束 → read the proposals → accept → the district wires itself.

---

### monolith7 — The Legacy

**Learns**: nothing new. This is where the world says what the player has become.

The vault the Empire built to keep the Legacy from anyone. Inside is the record —
every form of every Keymaker who worked on the Monolith, Lira's among them, her
name on the ones she left.

Taking it is not a pickup. It is the hall recognising the record the player has
been writing since the dungeon, and adding it to the rest.

**Goal**: open the vault → find Lira's forms → add your own record beside them.

---

### monolith8 — Fix It

**Learns**: nothing new. The reason all of it mattered.

The broken rule from the first line of the story — the one binding citizens, that
no one could fix. It is one board, one district, and the player fixes it. Not by
being given power over the Monolith, but by naming what needed naming and drawing
the roads that were cut.

The Empire's hold was never on the machine. It was on the vocabulary and the
connections, and both of those are now the player's.

**Goal**: the rule that started it → name → connect → apply a form → the rule stops binding anyone.

---

## make Kata — the feature

This is the only genuinely new capability in either world. Everything else in
worlds 2 and 3 rests on what already ships.

**Entry point.** The character action bubble — the one that opens when you press
Start standing on empty ground, where Call, Ride, Drop, Say and Map live — gains
**Kata**. That menu is the right home because it is the menu that acts on *you*:
applying a form is something the player does, not something done to a card.

**The gesture.**

1. Start on open ground → the bubble → **Kata**
2. A picker of the forms you **hold** — an unheld form has nothing to lend, so only
   初伝 and above are offered
3. Pick one
4. The board answers with **what is missing**, drawn as proposals rather than facts
5. Accept, and they are made

**Backend.** Most of this already exists. Kata progress is derived server-side on
every read, and each waza comes back with `satisfied`, `have`, `need`, `matchedIDs`
and a one-phrase `hint`. What is missing is the step from "this waza is
unsatisfied" to "here is the specific thing to make on *this* board":

```
POST /api/v1/kata/{id}/recommend   →  { proposals: [ ... ] }
```

Each proposal is executable rather than descriptive — a want to create with its
type and seed, a connection to draw between two named ends, a group to extend — so
the frontend can offer it and, on acceptance, carry it out. The matching work
(which pinned things, which existing wants, which group) is the same matching the
progress derivation already does; this endpoint returns the **complement** of it
instead of the count.

**Frontend.** A `Kata` item and picker stage in the character action bubble
(alongside `pick-call` / `pick-ride`), and the proposals rendered on the canvas in
the idiom the app already uses for recommendations — proposed, reviewable, accepted
or dismissed — rather than applied silently.

---

## The forms this world is built on

The fortress is practice; these are the names for it. A set of its own, in the
register of the 白帯 forms (駅・市・標・催・金):

| Kata | Waza | What it is | Practised in |
|---|---|---|---|
| **留** (tome) — to fix in place | `thing` pinned | a value that stays whether or not anything uses it | fortress3, 15 |
| **群** (mure) — the flock | a thing group with members | one name for many | fortress8 |
| **種** (tane) — the seed | `want_type` joined to a pinned thing | a want born knowing its value | fortress7 |
| **道** (michi) — the road | a connection between two wants | what one makes reaches the other | fortress12–14 |
| **束** (taba) — the sheaf | 道 repeated across a group | one intent, every destination | fortress16 |

束 is a `repeat` waza over 道, so the ladder cannot offer it to a player who has
not drawn roads by hand. The dependency between the forms is the dependency
between the stages, which is why the fortress is a prerequisite rather than a
recommendation.

**What the engine needs**: kata waza today are `want_type`, `thing` and `repeat`.
道 and 束 need a fourth kind, satisfied by a connection existing between two wants.
Everything else is already expressible — including the group axis, since
`join: { kind: thing_group }` is how the existing belts bind their waza together.
