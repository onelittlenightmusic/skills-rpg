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
