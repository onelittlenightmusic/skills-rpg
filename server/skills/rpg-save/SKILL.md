---
name: rpg-save
description: rpg-server の現在状態を指定スロットに保存する。スロット名は英数字・ハイフン・アンダースコア（最大32文字）。`autosave` `quicksave` も使える。
compatibility:
  python: ">=3.9"
metadata:
  type-name: rpg_save
  category: rpg
  final-result-field: ok
---

```bash
python3 "${CLAUDE_SKILL_DIR}/main.py" $ARGUMENTS
```

## 実行特性

| 項目 | 値 | 説明 |
|---|---|---|
| 実行モデル | `foreground` | 1回保存して終了 |

## パラメータ

| フィールド | 型 | 必須 | デフォルト | 説明 |
|---|---|---|---|---|
| `slot` | string | ✓ | — | スロット名 |
| `name` | string | — | — | 表示用ラベル |

## 出力フィールド

| フィールド名 | 型 | JSONパス | 永続化 | 説明 |
|---|---|---|---|---|
| `ok` | bool | `ok` | true | 保存に成功したか |
| `meta` | object | `meta` | true | 保存後のメタ情報 |

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
      name: my-rpg-save
      type: rpg_save
    spec:
      params:
        slot: autosave
        name: ""
```