# Stage 1 Playthrough — The Locked Room

> 主人公 `you` は部屋に閉じ込められている。`you` 自身ではドアを開けられない。
> AI エージェント `chap` を呼び、ドアを開けてもらってから脱出する。

ここでは **3つの経路で同じゴールに到達できる**ことを体験する:

| 経路 | 誰が叩く | actor |
|---|---|---|
| MCP `rpg-control-system` | チャットの中から直接 | `you` |
| Skill `/rpg-control` | Claude Code が代行 | `chap` |
| mywant want `rpg_control` | YAML から起動 | `chap` |

3経路すべて `rpg-server` の `/api/v1/control` に収束する。

---

## 0. セットアップ

```sh
cd ~/work/mywant-rpg
make build              # bin/rpg-server, bin/rpg-mcp
make install-skills     # ~/.mywant/custom-types/ に skills + want types を配備
./bin/rpg-server &      # :7100
```

`.mcp.json` の例:

```json
{
  "mcpServers": {
    "rpg": {
      "command": "/path/to/mywant-rpg/bin/rpg-mcp",
      "env": { "RPG_SERVER_URL": "http://localhost:7100" }
    }
  }
}
```

---

## 1. ナレーション開始

```sh
curl -s localhost:7100/api/v1/next-goal | jq .
```

→ 「まずは部屋を見回そう」

## 2. 観測する（`look_around` を解錠）

MCP: `rpg-observe target=stages.stage1.doors.door1`
HTTP:
```sh
curl -s "localhost:7100/api/v1/observe?target=stages.stage1.doors.door1"
```

`achievements_unlocked: ["look_around"]` が返り、next_goal が「ドアを自分で開けてみよう」に変わる。

## 3. you が open を試みる → 拒否（学習の核）

MCP: `rpg-control-system action=open target=door1`
HTTP:
```sh
curl -s -X POST localhost:7100/api/v1/control \
  -H 'content-type: application/json' \
  -d '{"actor":"you","action":"open","target":"door1"}'
```

→ HTTP 409 / `ok:false` / `reason: "you cannot perform \"open\""` / `hints` にヒント / `achievements_unlocked: ["attempted_self_unlock"]`。
next_goal が「chap に頼もう」に切り替わる。

> ここでセーブしておくと再現実験しやすい:
> ```sh
> curl -s -X POST localhost:7100/api/v1/saves/slot-1 \
>   -H 'content-type: application/json' \
>   -d '{"name":"after self-unlock attempt"}'
> ```

## 4. chap にドアを開けてもらう

選べる3経路（どれを使っても結果は同じ）:

### 4a. Skill 経路（推奨）

Claude Code から:
```
/rpg-control {"action":"open","target":"door1"}
```
内部で `skills/rpg-control/main.py` が走り `actor=chap` で `/control` を叩く。

### 4b. Want 経路

```sh
mywant apply examples/ask-chap-unlock.yaml
```
`rpg_control` want が一度だけ実行され、agent.yaml 経由で skill が起動する。
Want 状態に `ok / changes / achievements_unlocked / next_goal` が刻まれる。

### 4c. 直接 HTTP（仕組み確認用）

```sh
curl -s -X POST localhost:7100/api/v1/control \
  -H 'content-type: application/json' \
  -d '{"actor":"chap","action":"open","target":"door1"}'
```

いずれの経路でも `chap_unlocked_door1` が解錠される。

## 5. 部屋を出る

```sh
curl -s -X POST localhost:7100/api/v1/control \
  -H 'content-type: application/json' \
  -d '{"actor":"you","action":"move","target":"room2"}'
```

→ `escaped_room1` が解錠 → next_goal が `cleared: true` に。**Stage 1 クリア**。

## 6. セーブ／ロードを試す

```sh
# 一覧
curl -s localhost:7100/api/v1/saves | jq .

# slot-1 をロード（手順3直後の状態に戻る）
curl -s -X POST localhost:7100/api/v1/saves/slot-1/load
```

ロード後に `/api/v1/state` を見ると `door1.open=false` / `you.position=room1` に戻り、
mywant 側に `rpg_observe` want を立てていれば、次のサンプル時刻でキャッシュも追従する。

---

## トラブルシュート

| 症状 | 原因 | 対処 |
|---|---|---|
| `cannot reach rpg-server` | `RPG_SERVER_URL` 未設定 / サーバ未起動 | `./bin/rpg-server &`、env 設定 |
| `make install-skills` 後も want type が見えない | mywant が起動済み | mywant を再起動して custom-types を再読込 |
| ロード直後の want.state.current が古い | observer の poll 遅延（仕様） | 次サイクルで自動追従 |
