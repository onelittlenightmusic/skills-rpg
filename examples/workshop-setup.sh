#!/bin/sh
# Workshop: Family Rescue — start script
# RPGサーバーが全ての初期化（want作成・webhook登録・オープニングナレーション）を担当します。
# このスクリプトは debug/jump を呼ぶだけで十分です。
#
# 使い方:
#   mywant-rpg server start --mywant-url http://localhost:8080
#   ./workshop-setup.sh

RPG=${RPG_URL:-http://localhost:7100}

echo "== Jump RPG server to workshop1 =="
curl -sf -X POST "$RPG/api/v1/debug/jump" \
  -H 'Content-Type: application/json' \
  -d '{"stage":"workshop1"}' | python3 -m json.tool

echo ""
echo "RPG server will automatically:"
echo "  1. Clean up existing workshop1 wants"
echo "  2. Create wife-target, weather-check, daughter-target, route-search"
echo "  3. Register lifecycle webhook (mywant → RPG)"
echo "  4. Run opening narration via mywant-gui"
echo ""
echo "Player then adds the missing wants (remind-wife, book-hotel) on the canvas."
echo "done."
