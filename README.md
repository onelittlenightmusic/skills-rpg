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
