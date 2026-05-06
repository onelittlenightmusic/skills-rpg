---
name: rpg-save-list
description: rpg-server に保存されているセーブスロット一覧（slot 名・更新時刻・サマリ）を取得して JSON で返す。
compatibility:
  python: ">=3.9"
metadata:
  type-name: rpg_save_list
  category: rpg
  final-result-field: slots
---

```bash
python3 "${CLAUDE_SKILL_DIR}/main.py"
```

## 実行特性

| 項目 | 値 | 説明 |
|---|---|---|
| 実行モデル | `background` | スロット一覧を継続的にサンプルする |

## 出力フィールド

| フィールド名 | 型 | JSONパス | 永続化 | 説明 |
|---|---|---|---|---|
| `slots` | object | `slots` | true | スロットメタ情報の配列 |

## エラー時

```json
{ "error": "...", "slots": [] }
```

## mywant want としてデプロイする場合

```yaml
wants:
  - metadata:
      name: save-list
      type: rpg_save_list
    spec:
      params: {}
      requires:
        - rpg_save_list
      finalResultField: slots
```
