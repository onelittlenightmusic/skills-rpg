---
name: rpg-load
description: rpg-server の状態を指定スロットの内容で復元する。直後の next_goal を含めて返すので、ロード後に何をすべきか即時に分かる。
compatibility:
  python: ">=3.9"
metadata:
  type-name: rpg_load
  category: rpg
  final-result-field: ok
---

```bash
python3 "${CLAUDE_SKILL_DIR}/main.py" $ARGUMENTS
```

## 実行特性

| 項目 | 値 | 説明 |
|---|---|---|
| 実行モデル | `foreground` | 1回復元して終了 |

## パラメータ

| フィールド | 型 | 必須 | デフォルト | 説明 |
|---|---|---|---|---|
| `slot` | string | ✓ | — | 復元元スロット名 |

## 出力フィールド

| フィールド名 | 型 | JSONパス | 永続化 | 説明 |
|---|---|---|---|---|
| `ok` | bool | `ok` | true | 復元に成功したか |
| `slot` | string | `slot` | true | 復元したスロット |
| `next_goal` | object | `next_goal` | true | 復元後の next_goal |

## エラー時

```json
{ "error": "...", "ok": false }
```

## mywant want としてデプロイする場合

`requires` と `finalResultField` は want type 定義に既定値が設定されているため **省略可能**です。
最小構成で動作します。

```yaml
wants:
  - metadata:
      name: my-rpg-load
      type: rpg_load
    spec:
      params:
        slot: autosave
```