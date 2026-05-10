---
name: rpg-switch
description: デバイスをactivate（on=true）またはdeactivate（on=false）する。switch want cardのwebhookから呼ばれる。
compatibility:
  python: ">=3.9"
metadata:
  type-name: rpg_switch
  category: rpg
  final-result-field: summary
---

`${CLAUDE_SKILL_DIR}/main.py` に JSON 引数 `{"target": <device_id>, "on": true|false}` を渡す。
`on=true` なら `activate`、`on=false` なら `deactivate` を rpg-server に送る。

```bash
python3 "${CLAUDE_SKILL_DIR}/main.py" $ARGUMENTS
```

## パラメータ

| フィールド | 型 | 必須 | 説明 |
|---|---|---|---|
| `target` | string | ✓ | 操作するデバイスの ID |
| `on` | boolean | ✓ | true=activate / false=deactivate |

## 使用例

```bash
python3 "${CLAUDE_SKILL_DIR}/main.py" '{"target":"generator","on":true}'
python3 "${CLAUDE_SKILL_DIR}/main.py" '{"target":"alarm","on":false}'
```
