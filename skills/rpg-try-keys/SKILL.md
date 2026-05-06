---
name: rpg-try-keys
description: chap が持つ全ての鍵を指定したドアに順番に試し、開く鍵を自動で見つけて開錠する。
compatibility:
  python: ">=3.9"
metadata:
  type-name: rpg_try_keys
  category: rpg
  final-result-field: summary
---

`${CLAUDE_SKILL_DIR}/main.py` に JSON 引数 `{"target": <door_id>}` を渡すと、chap のインベントリにある全ての鍵を順番に試して正しい鍵を見つけ、ドアを開ける。

```bash
python3 "${CLAUDE_SKILL_DIR}/main.py" $ARGUMENTS
```

## 実行特性

| 項目 | 値 | 説明 |
|---|---|---|
| 実行モデル | `foreground` | 全鍵を試して結果を返し終了する |

## パラメータ

| フィールド | 型 | 必須 | デフォルト | 説明 |
|---|---|---|---|---|
| `target` | string | ✓ | — | 開けたいドアの ID |

## 出力フィールド

| フィールド名 | 型 | 説明 |
|---|---|---|
| `ok` | bool | 開錠できたか |
| `reason` | string | 失敗時の理由 |
| `changes` | object | 状態変化の差分 |
| `achievements_unlocked` | array | 新たに解錠された achievement |
| `next_goal` | object | 次に推奨されるゴール |

## 使用例

```bash
python3 "${CLAUDE_SKILL_DIR}/main.py" '{"target":"lab_door"}'
```

## mywant want としてデプロイする場合

`requires` と `finalResultField` は want type 定義に既定値が設定されているため **省略可能**です。
最小構成で動作します。

```yaml
wants:
  - metadata:
      name: my-rpg-try-keys
      type: rpg_try_keys
    spec:
      params:
        target: lab_door
```