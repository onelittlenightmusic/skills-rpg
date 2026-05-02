---
name: rpg-control
description: rpg-server に対してゲームアクション (open / move / pickup / observe など) を `chap` として実行する。`you` から拒否されたアクションを代行する用途で使う。
compatibility:
  python: ">=3.9"
metadata:
  type-name: rpg_control
  category: rpg
  final-result-field: summary
---

`${CLAUDE_SKILL_DIR}/main.py` に JSON 引数 `{"action": ..., "target": ..., "args": ...}` を渡すと、rpg-server の `/api/v1/control` を `actor=chap` で呼び出した結果を JSON で返す。

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
| `action` | string | ✓ | — | ゲームアクション名（observe / move / pickup / open など） |
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
