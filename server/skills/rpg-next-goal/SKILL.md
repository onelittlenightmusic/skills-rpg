---
name: rpg-next-goal
description: 主人公 you が次にやるべきこと（text / hint / required_skill）を rpg-server から取得して JSON で返す。プレイヤーへのナレーションや次の want を決めるのに使う。
compatibility:
  python: ">=3.9"
metadata:
  type-name: rpg_next_goal
  category: rpg
  final-result-field: text
---

```bash
python3 "${CLAUDE_SKILL_DIR}/main.py"
```

## 実行特性

| 項目 | 値 | 説明 |
|---|---|---|
| 実行モデル | `background` | next_goal を継続的にサンプルする |

## 出力フィールド

| フィールド名 | 型 | JSONパス | 永続化 | 説明 |
|---|---|---|---|---|
| `text` | string | `text` | true | 次にやるべきことの本文 |
| `hint` | string | `hint` | true | 達成のためのヒント |
| `required_skill` | string | `required_skill` | true | 推奨される Skill 名 |
| `cleared` | bool | `cleared` | true | 全ステージクリア済みか |

## エラー時

```json
{ "error": "cannot reach rpg-server: ..." }
```

## mywant want としてデプロイする場合

`requires` と `finalResultField` は want type 定義に既定値が設定されているため **省略可能**です。
最小構成で動作します。

> **canvas 配置に注意**: `metadata.labels` に `mywant.io/canvas-x` / `mywant.io/canvas-y` を指定しないと、canvas 原点 (0,0) 起点のフォールバック配置になり、ロボットキャラクター（CursorMan）から離れた位置にデプロイされる。デプロイ前に `mywant-gui i get`（または `mywant-gui cursor get`）でキャラクターの現在位置を確認し、その近くの座標を文字列で指定すること。詳細は mywant-gui skill の SKILL.md「YAMLでwantをデプロイする際のcanvas座標」を参照。

```yaml
wants:
  - metadata:
      name: my-rpg-next-goal
      type: rpg_next_goal
      labels:
        mywant.io/canvas-x: "12"
        mywant.io/canvas-y: "8"
    spec:
      params:
        {}
```