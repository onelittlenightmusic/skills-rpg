# Skills RPG

> *An interactive RPG to learn [MCP](https://modelcontextprotocol.io/), [Agent Skills](https://docs.anthropic.com/en/docs/claude-code/skills), and [Wants](https://github.com/onelittlenightmusic/mywant) — by playing through them.*

**[日本語での説明はこちら (README-jp.md)](README-jp.md)**

---

## The Story

The **Monolith** controls all of society's infrastructure — administration, healthcare, distribution, communications. But it has grown old and corrupt. Citizens suffer under broken rules that no one can fix. The Empire hoards the Monolith's power and imprisons anyone who dares to challenge it.

**You** are an engineer. Watching citizens suffer, you set out to fix the Monolith. The Imperial Army captured you at the fortress gates and threw you into their dungeon.

When you woke up, a strange AI agent named **chap** was waiting. A single cable runs from its back — an **[MCP (Model Context Protocol)](https://modelcontextprotocol.io/)** connection cable that links it directly to the dungeon control system.

You don't know if you can trust it yet. But right now, you have no other choice.

*Escape the dungeon. Claim the Legacy. Save the people.*

---

## What You Learn

This game is designed to give you **hands-on experience with three escalating methods of AI agent intervention**:

```
MCP (direct command) → Agent Skills (automation) → Wants (intent definition & scaling)
```

> [MCP](https://modelcontextprotocol.io/) · [Agent Skills](https://docs.anthropic.com/en/docs/claude-code/skills) · [Wants / mywant](https://github.com/onelittlenightmusic/mywant)

Each stage teaches one concept. By the time you escape, you will have used all three layers in real code — not just read about them.

| # | Stage | Concept |
|---|-------|---------|
| 1 | The Locked Room | Ask chap directly via MCP |
| 2 | The Vault Door | Manual trial-and-error; feel the limits of repetition |
| 3 | The Forgotten Lab | Automate repetition with an Agent Skill |
| 4 | Dark Lab | Activate a device to satisfy a door precondition |
| 5 | Alarm Room | Deactivate a device to remove a blocker |
| 6 | Control Room | Correct sequence: deactivate → activate → open |
| 7 | The Want Factory | Scale across 5 doors with a deployed Want |
| 8 | The Want Factory: Parallel | Open 5 doors at once with a `parallelize` Want |
| 9 | Lights Out | Combine two Wants to change the world without MCP |

**Core insight:** Skills are *tools*. Wants are the *definition of intent* that orchestrates those tools. Define the intent once, and the scale is unlimited.

---

## Prerequisites

> **Note**: For Stages 7–9, you will need the configuration files located in the `examples/` directory of this repository. If you installed via Homebrew, please clone the [skills-rpg repository](https://github.com/onelittlenightmusic/skills-rpg) to access them.

| Requirement | Notes |
|-------------|-------|
| [Claude Code](https://claude.ai/code) | The CLI / IDE extension |
| Go 1.22+ | To build the RPG server |
| Python 3.9+ | Some skills use Python |
| [mywant](https://github.com/onelittlenightmusic/mywant) | Required for stages 7–9 |
| mywant-skills | Included in the mywant repo |

---

## Installation

### Option 1: Homebrew (Recommended)
```sh
brew install onelittlenightmusic/mywant/mywant-rpg
```

### Option 2: Build from source
```sh
git clone https://github.com/onelittlenightmusic/skills-rpg.git
cd skills-rpg
make build
# Add bin/ to your PATH
```

## Setup & Start Playing

### 1. Initialize Server & Skills
```sh
# Start the server
mywant-rpg server start

# Install skills (Gemini, Claude, MyWant, or Codex)
mywant-rpg install gemini
```

### 2. Configure Agent (MCP)
Add this to your agent's `settings.json` (e.g., `~/.gemini/settings.json`):

```json
"mcpServers": {
  "rpg": {
    "command": "mywant-rpg",
    "args": ["mcp", "serve"]
  }
}
> **Note:** After configuring the MCP server, restart your agent. Then run `/mcp list` to verify that the `rpg` server and tools like `rpg_observe` appear.

```

### 3. Start the Game
Open your agent and say:

> *"Let's play skills-rpg. Use `rpg_start` to get the context, then `rpg_observe` to look around."*
> （日本語: 「skills-rpgをプレイしよう。rpg_startでコンテキストを取得して、rpg_observeで周りを見回して。」）

---

## Language Setting
Default is English. To switch to Japanese:
```sh
curl -X PUT http://localhost:7100/api/v1/settings \
  -H "Content-Type: application/json" \
  -d '{"language":"ja"}'
```

---

## Architecture

```
Claude Code (you)
    │
    ├── MCP tools (rpg_control_system, rpg_observe, …)
    │       └── bin/rpg-mcp  ──────────────────────────────┐
    │                                                        │
    └── Agent Skills (/rpg-try-keys, /rpg-observe, …)       │
            └── skills/*.sh / *.py  ───────────────────────►│
                                                             │
                                               bin/rpg-server  :7100
                                               (game state, rules, narration)
                                                             │
                                               ~/.mywant-rpg/current.yaml
```

Wants (stages 7–9) are deployed via the `mywant-deploy` skill and run as background agents that call the same HTTP API.

## Settings API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/settings` | GET | Get current settings |
| `/api/v1/settings` | PUT | Update settings (e.g. `{"language":"ja"}`) |

Supported languages: `en` (default), `ja`

## Development

### Reloading stage YAMLs

```sh
curl -X POST http://localhost:7100/api/v1/reset
```

### Rebuild and restart

```sh
make restart
```

### Troubleshooting

| Symptom | Fix |
|---------|-----|
| `cannot reach rpg-server` | Ensure `./bin/rpg-server` is running on port 7100 |
| MCP server not connected | Restart Claude Code; check `claude mcp list` |
| Tools missing in chat | Run `/mcp reload` in Claude Code |
| Stage changes not reflected | Run `curl -X POST http://localhost:7100/api/v1/reset` |
| Stages 7–9 skill not found | Run `make install-claude` to link mywant-skills |
