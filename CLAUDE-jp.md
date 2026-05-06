# skill-rpg CLAUDE.md

## ドキュメント

### STAGES.md
各ステージの学習テーマ・扉・デバイス・鍵・クリア条件・ゴール手順を記録する。
ステージを追加・変更したら `STAGES.md` も必ず更新すること。

---

## 開発メモ

### ステージYAML変更後はサーバーリセットが必要

`stages/*.yaml` を編集した後にサーバーを再起動しても、`~/.mywant-rpg/current.yaml` が存在する限りサーバーはそちらを読み込む（ステージYAMLは読み直されない）。

新しいステージ内容をサーバーに反映するには、再起動後に必ずリセットを実行すること：

```bash
curl -s -X POST http://localhost:7100/api/v1/reset
```

または MCP ツール経由でも可。

**将来対策案**: サーバー起動時にステージYAMLのmtimeをcurrent.yamlと比較し、新しければ自動リブートする。
