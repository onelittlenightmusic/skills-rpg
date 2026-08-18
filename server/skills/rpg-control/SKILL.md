---
name: rpg-control
description: rpg-server に対してゲームアクション (open / activate / move / pickup / advance / return / inspect / name_thing / pin_thing など) を実行する。既定は `chap`（`you` から拒否されたアクションの代行）。`actor` を `you` にすれば移動・ステージ移動・調査・名付け・pin も呼べる。
compatibility:
  python: ">=3.9"
metadata:
  type-name: rpg_control
  category: rpg
  final-result-field: summary
---

`${CLAUDE_SKILL_DIR}/main.py` に JSON 引数 `{"actor": ..., "action": ..., "target": ..., "args": ...}` を渡すと、rpg-server の `/api/v1/control` を呼び出した結果を JSON で返す。

アクターは2人で、できることが分かれている（chap は扉とデバイス、you は移動・取得・ステージ移動・調査・名付け・pin）。誰が何を実行できるか、前提条件は何かは [docs/actions.md](https://github.com/onelittlenightmusic/skills-rpg/blob/main/docs/actions.md) を参照。

```bash
python3 "${CLAUDE_SKILL_DIR}/main.py" $ARGUMENTS
```

## 実行特性

| 項目 | 値 | 説明 |
|---|---|---|
| 実行モデル | `foreground` | 1回実行して結果を返し終了する |

## パラメータ

| フィールド | 型 | 必須 | デフォルト | 説明 |
|---|---|---|---|---|
| `actor` | string | — | `chap` | 実行者。`chap`（扉・デバイス）か `you`（移動・取得・advance / return） |
| `action` | string | ✓ | — | ゲームアクション名（observe / move / pickup / open / advance / return / inspect / name_thing / pin_thing など） |
| `target` | string | — | — | 対象（door id / item id / waypoint id など） |
| `args` | object | — | — | アクション固有の追加引数 |

## 出力フィールド

| フィールド名 | 型 | JSONパス | 永続化 | 説明 |
|---|---|---|---|---|
| `ok` | bool | `ok` | true | アクションが受理されたか |
| `reason` | string | `reason` | true | 拒否時の理由（受理時は空） |
| `hints` | object | `hints` | true | 拒否時のヒント（chap 経由を促す等） |
| `changes` | object | `changes` | true | 状態変化の差分 |
| `achievements_unlocked` | object | `achievements_unlocked` | true | 今回新たに解錠された achievement |
| `next_goal` | object | `next_goal` | true | 直後に推奨される次のゴール |
| `summary` | string | `reason` | true | 結果サマリ（reason をそのまま流用） |

## 使用例

### Open a locked door

```bash
python3 "${CLAUDE_SKILL_DIR}/main.py" '{"action":"open","target":"door1"}'
```

### Move `you`, and go to the next stage

```bash
python3 "${CLAUDE_SKILL_DIR}/main.py" '{"actor":"you","action":"move","target":"control_room"}'
python3 "${CLAUDE_SKILL_DIR}/main.py" '{"actor":"you","action":"advance"}'
```

### 城塞都市（world: fortress）の3つ

盤面を読む・値に名前をつける・盤面に留める。どれも `you` の動作で、chap は
代行できない——世界への介入ではなく、プレイヤーが世界について「言う」ことだから。
聞かれたら説明はできる。

```bash
# 何を待っているのか、カードを読む（door なら waiting_for / subtype / satisfied が返る）
python3 "${CLAUDE_SKILL_DIR}/main.py" '{"actor":"you","action":"inspect","target":"north_gate"}'

# 持ち物に書かれている値を読む
python3 "${CLAUDE_SKILL_DIR}/main.py" '{"actor":"you","action":"inspect","target":"lira_mark"}'

# その値に名前をつけて、都市に伝える（target の持ち物から値を読む）
python3 "${CLAUDE_SKILL_DIR}/main.py" '{"actor":"you","action":"name_thing","target":"lira_mark","args":{"subtype":"maker"}}'

# 持ち物ではなく直接、値を指定して名付ける
python3 "${CLAUDE_SKILL_DIR}/main.py" '{"actor":"you","action":"name_thing","args":{"subtype":"level","value":"3.5"}}'

# 名付けた値を盤面に留める（参照が無くなっても消えなくなる）
python3 "${CLAUDE_SKILL_DIR}/main.py" '{"actor":"you","action":"pin_thing","target":"lira_mark","args":{"subtype":"maker"}}'
```

ステージ間の移動は `advance` / `return` を使う。どちらも両ステージをそのまま残す
（扉もデバイスも持ち物も achievements も）。`debug jump` は YAML の初期状態へ戻す
テスト用リスタートなので、移動には使わない。

出力:
```json
{
  "ok": true,
  "actor": "chap",
  "action": "open",
  "target": "door1",
  "changes": {"doors.door1.open": true, "doors.door1.locked": false},
  "achievements_unlocked": ["chap_unlocked_door1"],
  "next_goal": {"text": "ドアが開いた。部屋を出よう"}
}
```

## エラー時

```json
{ "error": "cannot reach rpg-server at ...", "ok": false }
```

## mywant want としてデプロイする場合

`requires` と `finalResultField` は want type 定義に既定値が設定されているため **省略可能**です。
最小構成で動作します。

> **canvas 配置に注意**: `metadata.labels` に `mywant.io/canvas-x` / `mywant.io/canvas-y` を指定しないと、canvas 原点 (0,0) 起点のフォールバック配置になり、ロボットキャラクター（CursorMan）から離れた位置にデプロイされる。デプロイ前に `mywant-gui i get`（または `mywant-gui cursor get`）でキャラクターの現在位置を確認し、その近くの座標を文字列で指定すること。詳細は mywant-gui skill の SKILL.md「YAMLでwantをデプロイする際のcanvas座標」を参照。

```yaml
wants:
  - metadata:
      name: my-rpg-control
      type: rpg_control
      labels:
        mywant.io/canvas-x: "12"
        mywant.io/canvas-y: "8"
    spec:
      params:
        action: open
        target: door1
```