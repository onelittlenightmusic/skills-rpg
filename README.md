# Skill RPG

An RPG-style learning game for getting comfortable with **mywant** and **Claude Code Agent Skills**.

You (`you`) are trapped. Your AI agent (`chap`) can act on your behalf via Skills, MCP tools, or
mywant want types. Three routes converge on the same game-state HTTP API; that convergence is the
lesson.

## Components

| Component | Role |
|---|---|
| `cmd/rpg-server` | Authoritative game-state HTTP server (state, control, saves) |
| `cmd/rpg-mcp` | Thin stdio MCP wrapper around rpg-server (`actor=you`) |
| `skills/` | Machine Readable Skills — invoke rpg-server with `actor=chap` |
| `generated-want-types/` | Want type YAML auto-generated from skills |
| `stages/` | Stage definitions loaded by rpg-server on first run |
| `examples/` | mywant YAML examples chaining buttons / observers / actions |

## Quick start

```sh
make build
./bin/rpg-server &              # default :7100
curl localhost:7100/api/v1/next-goal
```

See `docs/playthrough-stage1.md` for the intended first run.

## Prerequisites: mywant-skills

Stage 7 and 8 use `mywant` wants to automate door-opening. The `/mywant-deploy`
skill (from [mywant-skills](https://github.com/onelittlenightmusic/mywant)) is
required. Install it first:

```sh
# See ~/.mywant/custom-types/mywant-skills/README.md for details
```

Once mywant-skills is installed at `~/.mywant/custom-types/mywant-skills/`,
`make install-claude` (below) will automatically link them into `~/.claude/skills/`.

## Installing Skills into Claude Code

The `skills/` directory contains Agent Skills (e.g. `rpg-try-keys`). To make them
available as `/skill-name` slash commands in Claude Code, register the `skills/`
directory as a project:

```sh
python3 - <<'EOF'
import json, pathlib

claude_json = pathlib.Path.home() / ".claude.json"
skills_dir  = str(pathlib.Path(__file__).resolve().parent / "skills")  # adjust if needed

with open(claude_json) as f:
    d = json.load(f)

if skills_dir not in d.get("projects", {}):
    d.setdefault("projects", {})[skills_dir] = {}
    with open(claude_json, "w") as f:
        json.dump(d, f, indent=2)
    print("registered:", skills_dir)
else:
    print("already registered")
EOF
```

Or run the one-liner directly (replace the path if your clone is elsewhere):

```sh
python3 -c "
import json, pathlib
p = pathlib.Path.home() / '.claude.json'
d = json.load(open(p))
s = '$PWD/skills'
d.setdefault('projects', {})[s] = {}
json.dump(d, open(p,'w'), indent=2)
print('registered', s)
"
```

After registration the following skills become available in any Claude Code session:

| Skill | Description |
|---|---|
| `/rpg-try-keys` | Try all of chap's keys on a door and unlock it |
| `/rpg-control` | Send a control action as chap |
| `/rpg-observe` | Observe a subtree of game state |
| `/rpg-next-goal` | Show the next suggested goal |
| `/rpg-save` / `/rpg-load` | Save-slot management |

> **Note:** No restart is required — skills are re-read on each invocation.

## Connecting Claude Code to the rpg MCP server

`bin/rpg-mcp` is a stdio MCP server that wraps `rpg-server`. The project already
includes `.claude/settings.json` with the MCP configuration, so no manual registration
is needed for project-scope use.

### 1. Build and start `rpg-server`

```sh
make build
./bin/rpg-server &
curl -sf http://localhost:7100/healthz       # → {"ok":true}
```

### 2. Restart Claude Code

Restart Claude Code so the MCP server (configured in `.claude/settings.json`) is
loaded. After restart the following tools become available (prefixed `mcp__rpg__`):

| Tool | Purpose |
|---|---|
| `rpg_next_goal` | Show what to do next |
| `rpg_observe` | Read a subtree of game state by dot-path |
| `rpg_control_system` | Perform an action **as `chap`** (the AI agent) |
| `rpg_save` / `rpg_load` / `rpg_save_list` / `rpg_save_delete` | Save-slot management |

Try in chat: *"Use rpg_control_system to open door1."*

### Optional: user-scope registration

To make the server available across all projects instead:

```sh
claude mcp add rpg \
  --scope user \
  --env RPG_SERVER_URL=http://localhost:7100 \
  -- "$PWD/bin/rpg-mcp"
```

### Verify

```sh
claude mcp list
# rpg: /Users/.../skill-rpg/bin/rpg-mcp  - ✓ Connected
```

### Removing the server

```sh
claude mcp remove rpg --scope user
```

### Troubleshooting

| Symptom | Fix |
|---|---|
| `cannot reach rpg-server` from a tool | Make sure `./bin/rpg-server` is running and `RPG_SERVER_URL` matches its port |
| `claude mcp list` shows the server but not "Connected" | Check stderr in Claude Code logs; rebuild with `make build` |
| Tools missing in chat | Restart Claude Code after `claude mcp add` |
