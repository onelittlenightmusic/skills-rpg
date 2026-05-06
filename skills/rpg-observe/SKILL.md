---
name: rpg-observe
description: rpg-server のゲーム状態の指定部分（path）を継続的に観測して JSON で返す。`you` の現在位置や特定ドア・アイテムの状態を mywant want のキャッシュとして取り込むのに使う。
compatibility:
  python: ">=3.9"
metadata:
  type-name: rpg_observe
  category: rpg
  final-result-field: value
---

`${CLAUDE_SKILL_DIR}/main.py` に JSON 引数 `{"target": "stages.stage1.doors.door1"}` を渡すと、当該パスの値を JSON で返す。`target` 省略で全状態。

```bash
python3 "${CLAUDE_SKILL_DIR}/main.py" $ARGUMENTS
```

## 実行特性

| 項目 | 値 | 説明 |
|---|---|---|
| 実行モデル | `background` | 周期的に状態をサンプリングして want に反映する |

## パラメータ

| フィールド | 型 | 必須 | デフォルト | 説明 |
|---|---|---|---|---|
| `target` | string | — | "" | 観測対象の dot-path（空 = 全状態） |

## 出力フィールド

| フィールド名 | 型 | JSONパス | 永続化 | 説明 |
|---|---|---|---|---|
| `value` | object | `value` | true | 指定パスのサブツリー |
| `target` | string | `target` | true | 観測した path |
| `achievements_unlocked` | object | `achievements_unlocked` | true | observe により解錠された achievement |
| `next_goal` | object | `next_goal` | true | observe 後の next_goal |

## 使用例

### Watch door1

```bash
python3 "${CLAUDE_SKILL_DIR}/main.py" '{"target":"stages.stage1.doors.door1"}'
```

出力:
```json
{
  "target": "stages.stage1.doors.door1",
  "value": {"between": ["room1","room2"], "open": false, "locked": true},
  "achievements_unlocked": ["look_around"],
  "next_goal": {"text": "..."}
}
```

## エラー時

```json
{ "error": "cannot reach rpg-server: ..." }
```

## mywant want としてデプロイする場合

`requires` と `finalResultField` は want type 定義に既定値が設定されているため **省略可能**です。
最小構成で動作します。

```yaml
wants:
  - metadata:
      name: my-rpg-observe
      type: rpg_observe
    spec:
      params:
        target: stages.stage1
```