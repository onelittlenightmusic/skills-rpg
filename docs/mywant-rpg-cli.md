# mywant-rpg CLI Reference

`mywant-rpg` is a CLI plugin for [MyWant](https://github.com/onelittlenightmusic/MyWant) that provides a command-line interface to the skills-rpg game server.

## Installation

```sh
brew install onelittlenightmusic/mywant/mywant-rpg
```

This installs `mywant` (the core CLI) as a dependency automatically.

## Invocation

You can invoke the plugin in two ways:

```sh
mywant rpg <command>     # via MyWant plugin dispatch (recommended)
mywant-rpg <command>     # directly
```

## Global Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--server-url` | `http://localhost:7100` | rpg-server base URL |
| `--json` | false | Output raw JSON |

The server URL can also be set via the `RPG_SERVER_URL` environment variable.

---

## Commands

### `start`

Show your role, the rules of the game, and the current game state. **Call this first at the beginning of every session.**

```sh
mywant rpg start
```

---

### `goal`

Get the next suggested goal, including a hint and recommended skill.

```sh
mywant rpg goal
```

**Example output:**
```json
{
  "hint": "rpg_control_system actor=chap action=open target=door1",
  "text": "Open the door to the next room",
  "required_skill": "rpg-control"
}
```

---

### `observe [path]`

Read the current game state. Without a path, returns the full state. With a dot-path, returns only that subtree.

```sh
mywant rpg observe               # full state
mywant rpg observe you           # your current position
mywant rpg observe stages.stage1.doors.door1  # specific door state
```

**Example output:**
```json
{
  "value": {
    "position": "control_room"
  }
}
```

---

### `control <action> [target]`

Perform a game action. What the actions are, who may perform each one and what
has to be true for them to succeed is [actions.md](actions.md); this is only how
to send one.

```sh
mywant rpg control <action> [target] [--actor chap|you] [--arg key=value ...]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--actor` | `chap` | Who is acting: `chap` (the agent) or `you` (the player) |
| `--arg` | | Extra argument as `key=value`, repeatable |

The actor matters: the two halves of the game are split between them, and
either one asked to do the other's work is refused. chap opens doors and works
devices; you walks, picks things up and travels between stages.

```sh
# chap's half — the default
mywant rpg control open door1
mywant rpg control deactivate alarm_system
mywant rpg control state main_generator

# you's half
mywant rpg control move control_room --actor you
mywant rpg control pickup key_bronze --actor you

# Travel between stages — the ordinary way, which leaves both stages standing
mywant rpg control advance --actor you
mywant rpg control return  --actor you

# Naming the key to try, so a wrong one is a wrong one rather than an
# automatic pick from what chap happens to be carrying
mywant rpg control open vault_door --arg key=key_bronze
```

---

### `saves`

List all existing save slots with metadata.

```sh
mywant rpg saves
```

**Example output:**
```json
[
  { "slot": "quicksave", "name": "Quick Save", "updated_at": "2026-05-07T08:30:00Z" },
  { "slot": "before-stage3", "name": "Before stage 3", "updated_at": "2026-05-06T12:00:00Z" }
]
```

---

### `save <slot>`

Save the current game state to a named slot. Reserved slot names: `autosave`, `quicksave`.

```sh
mywant rpg save quicksave
mywant rpg save before-stage3 --name "Before stage 3"
```

| Flag | Description |
|------|-------------|
| `--name` | Human-readable label for the slot |

---

### `load <slot>`

Restore game state from a save slot. Returns the next goal after loading.

```sh
mywant rpg load quicksave
mywant rpg load before-stage3
```

---

### `rm <slot>`

Delete a save slot. Alias: `delete`.

```sh
mywant rpg rm quicksave
mywant rpg delete before-stage3
```

---

### `server`

Manage the rpg-server process. Useful when running the server locally.

#### `server start`

Start rpg-server as a background process. Logs are written to `~/.mywant/rpg-server.log`.

```sh
mywant rpg server start
mywant rpg server start --port 7200
mywant rpg server start --reset          # discard saved state and restart fresh
mywant rpg server start --bin /path/to/rpg-server
```

| Flag | Default | Description |
|------|---------|-------------|
| `--bin` | auto-detected | Path to the `rpg-server` binary |
| `--port` | `7100` | Port to listen on |
| `--data-dir` | `~/.mywant-rpg` | Directory for save data |
| `--stages-dir` | `stages/` | Directory containing stage YAML files |
| `--reset` | false | Discard current state and start fresh |

The binary is auto-detected in this order:
1. `--bin` flag
2. `rpg-server` in `PATH`
3. `~/work/skills-rpg/bin/rpg-server`
4. `/usr/local/bin/rpg-server`

#### `server stop`

Stop the rpg-server process that was started by this plugin.

```sh
mywant rpg server stop
```

#### `server status`

Check whether the managed rpg-server process is running.

```sh
mywant rpg server status
```

> **Note:** This only tracks processes started via `mywant rpg server start`. If rpg-server was started another way (e.g. `make run-server`), the status will show `stopped` even though the server is reachable.

---

### `debug jump <stage>`

Restart any stage from scratch. **Restores that stage to its pristine YAML
definition — doors shut, devices off, items back where they started — and
clears all achievements and inventory.** For testing only.

```sh
mywant rpg debug jump stage1
mywant rpg debug jump stage4
```

> **Not a way to travel.** Used to move between stages, this makes every trip
> out and back re-arm the alarms and forget what you were carrying. Use
> `control advance` / `control return` (see [actions.md](actions.md)), which
> leave both stages exactly as they stand.

---

### `install <target>` / `uninstall <target>`

Copy the bundled skills into an agent's skill directory, so it can call them.

```sh
mywant rpg install claude     # ~/.claude/skills
mywant rpg install gemini
mywant rpg install codex
mywant rpg install mywant     # ~/.mywant/custom-types — the want types the canvas draws
```

`uninstall` takes the same four targets and removes what was copied.

---

### `mcp serve`

Run the MCP server on stdio, exposing the game as agent tools (`rpg_start`,
`rpg_observe`, `rpg_control_system`, …). This is what an agent runtime launches;
you do not normally run it by hand.

```sh
mywant rpg mcp serve
```

Unlike `control`, the MCP tool takes an actor argument, so an agent reaches both
halves of the game through it.

---

## Typical Session Workflow

```sh
# 1. Start the server (if not already running)
mywant rpg server start

# 2. Begin the session
mywant rpg start

# 3. Check what to do next
mywant rpg goal

# 4. Observe surroundings
mywant rpg observe you

# 5. Take action — chap opens, you walks
mywant rpg control open door1
mywant rpg control move entrance --actor you

# 6. Save progress
mywant rpg save quicksave

# 7. Continue...
mywant rpg goal
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `RPG_SERVER_URL` | Override the default rpg-server URL (`http://localhost:7100`) |
