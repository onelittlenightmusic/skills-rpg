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

> **canvas 配置に注意**: `metadata.labels` に `mywant.io/canvas-x` / `mywant.io/canvas-y` を指定しないと、canvas 原点 (0,0) 起点のフォールバック配置になり、ロボットキャラクター（CursorMan）から離れた位置にデプロイされる。デプロイ前に `mywant-gui i get`（または `mywant-gui cursor get`）でキャラクターの現在位置を確認し、その近くの座標を文字列で指定すること。詳細は mywant-gui skill の SKILL.md「YAMLでwantをデプロイする際のcanvas座標」を参照。

```yaml
wants:
  - metadata:
      name: my-rpg-load
      type: rpg_load
      labels:
        mywant.io/canvas-x: "12"
        mywant.io/canvas-y: "8"
    spec:
      params:
        slot: autosave
```