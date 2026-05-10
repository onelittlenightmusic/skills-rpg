# Skills RPG (日本語)

> *[MCP](https://modelcontextprotocol.io/), [Agent Skills](https://docs.anthropic.com/en/docs/claude-code/skills), [Wants](https://github.com/onelittlenightmusic/mywant) を体験しながら学ぶインタラクティブRPG。*

---

## 概要

このゲームは、AIエージェントの操作方法を3つのステップで学べるように設計されています。

```
MCP (直接操作) → Agent Skills (自動化) → Wants (意図の定義とスケーリング)
```

---

## インストール

### オプション1: Homebrew (推奨)
```sh
brew install onelittlenightmusic/mywant/mywant-rpg
```

### オプション2: ソースからビルド
```sh
git clone https://github.com/onelittlenightmusic/skills-rpg.git
cd skills-rpg
make build
# bin/ ディレクトリにパスを通してください
```

## セットアップと開始手順

### 1. サーバーの初期化とスキルのインストール
```sh
# サーバーを起動
mywant-rpg server start

# 使用するエージェント用のスキルをインストール (gemini, claude, mywant, codex)
mywant-rpg install gemini
```

### 2. エージェント設定 (MCP)
エージェントの設定ファイル（例: `~/.gemini/settings.json` や `~/.claude/settings.json`）に以下を追記してください。

```json
"mcpServers": {
  "rpg": {
    "command": "mywant-rpg",
    "args": ["mcp", "serve"]
  }
}
```

> **注意:** 設定後、エージェントを再起動してください。その後、`/mcp list` コマンドを実行し、`rpg` サーバーと `rpg_observe` などのツールが一覧に表示されることを確認してください。

### 3. ゲーム開始
エージェントを開き、以下のように話しかけてください：

> *「skills-rpgをプレイしよう。rpg_startでコンテキストを取得して、rpg_observeで周りを見回して。」*

---

## 言語設定
デフォルトは英語です。日本語に切り替えるには：
```sh
curl -X PUT http://localhost:7100/api/v1/settings \
  -H "Content-Type: application/json" \
  -d '{"language":"ja"}'
```
設定は `~/.skills-rpg.conf` に保存されます。
