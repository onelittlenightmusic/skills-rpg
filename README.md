# Skills RPG

> *An interactive RPG to learn [MCP](https://modelcontextprotocol.io/), [Agent Skills](https://docs.anthropic.com/en/docs/claude-code/skills), and [Wants](https://github.com/onelittlenightmusic/mywant) — by playing through them.*

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

| Requirement | Notes |
|-------------|-------|
| [Claude Code](https://claude.ai/code) | The CLI / IDE extension |
| Go 1.22+ | To build the RPG server |
| Python 3.9+ | Some skills use Python |
| [mywant](https://github.com/onelittlenightmusic/mywant) | Required for stages 7–9 |
| mywant-skills | Included in the mywant repo |

---

## Setup (English)

### 1. Clone and build

```sh
git clone https://github.com/onelittlenightmusic/skills-rpg.git
cd skills-rpg
make build
```

### 2. Start the RPG server

```sh
./bin/rpg-server &
curl -sf http://localhost:7100/healthz   # → {"ok":true}
```

The server saves game state to `~/.mywant-rpg/current.yaml`.  
To reset back to stage 1 at any time:

```sh
curl -X POST http://localhost:7100/api/v1/reset
```

### 3. Install Skills into Claude Code

```sh
make install-skills
```

This registers the `skills/` directory so the following slash commands become available in Claude Code:

| Skill | What it does |
|-------|-------------|
| `/rpg-observe` | Read the current game state and narration |
| `/rpg-control` | Send a control action as chap |
| `/rpg-try-keys` | Try all keys on a door and unlock it |
| `/rpg-next-goal` | Show the current objective |
| `/rpg-save` / `/rpg-load` | Save-slot management |

> Run `/mcp reload` in Claude Code after installing.

### 4. Connect the MCP server

The project includes `.claude/settings.json` with the MCP configuration.  
After building, **restart Claude Code** — the following MCP tools become available automatically:

| Tool | Purpose |
|------|---------|
| `rpg_observe` | Read game state by dot-path |
| `rpg_control_system` | Perform an action as chap or you |
| `rpg_next_goal` | Show next objective |
| `rpg_start` | Get full onboarding context |
| `rpg_save` / `rpg_load` / `rpg_save_list` / `rpg_save_delete` | Saves |
| `rpg_debug_jump_stage` | Jump to any stage (debug) |

To add the MCP server globally (all projects):

```sh
claude mcp add rpg --scope user -- "$PWD/bin/rpg-mcp"
```

### 5. Install mywant-skills (required for stages 7–9)

Stages 7–9 require the `mywant-deploy` skill from [mywant-skills](https://github.com/onelittlenightmusic/mywant).
Follow the installation guide in that repository, then run:

```sh
make install-claude   # links mywant-skills into ~/.claude/skills/
```

### 6. Start playing

Open Claude Code and say:

> *"Let's play skills-rpg. Use rpg_start to get the context, then rpg_observe to look around."*

Follow chap's guidance stage by stage.

### Language setting

By default the game runs in English. To switch to Japanese:

```sh
curl -X PUT http://localhost:7100/api/v1/settings \
  -H "Content-Type: application/json" \
  -d '{"language":"ja"}'
```

Setting is saved to `~/.skills-rpg.conf`.

---

## セットアップ (日本語)

### 1. クローンとビルド

```sh
git clone https://github.com/onelittlenightmusic/skills-rpg.git
cd skills-rpg
make build
```

### 2. RPGサーバーを起動

```sh
./bin/rpg-server &
curl -sf http://localhost:7100/healthz   # → {"ok":true}
```

ゲーム状態は `~/.mywant-rpg/current.yaml` に保存されます。  
いつでもステージ1に戻すには：

```sh
curl -X POST http://localhost:7100/api/v1/reset
```

### 3. スキルをClaude Codeにインストール

```sh
make install-skills
```

インストール後、Claude Code で `/mcp reload` を実行してください。

### 4. MCPサーバーに接続

`.claude/settings.json` にMCP設定が含まれています。  
ビルド後に **Claude Code を再起動** すると、MCPツールが自動的に有効になります。

### 5. mywant-skillsをインストール（ステージ7〜9に必要）

[mywant](https://github.com/onelittlenightmusic/mywant) リポジトリの手順に従ってmywant-skillsをインストールした後：

```sh
make install-claude
```

### 6. 日本語に切り替え

```sh
curl -X PUT http://localhost:7100/api/v1/settings \
  -H "Content-Type: application/json" \
  -d '{"language":"ja"}'
```

設定は `~/.skills-rpg.conf` に保存されます。

### 7. ゲーム開始

Claude Code を開いて次のように話しかけてください：

> *「skills-rpgをプレイしよう。rpg_startでコンテキストを取得して、rpg_observeで周りを見回して。」*

chapのガイダンスに従ってステージを進めてください。

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
