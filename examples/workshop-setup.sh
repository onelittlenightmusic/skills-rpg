#!/bin/sh
# Workshop: Family Rescue — canvas setup
# 妻への傘リマインドチェーン + 娘の箱根旅行チェーン
set -eu

MYWANT=${MYWANT_URL:-http://localhost:8080}
RPG=${RPG_URL:-http://localhost:7100}

api() {
  curl -sf -X "$1" "$MYWANT$2" -H 'Content-Type: application/json' ${3:+-d "$3"}
}

echo "== cleanup: remove existing game=workshop1 wants"
EXISTING=$(curl -sf "$MYWANT/api/v1/wants" | jq -r '.wants[]? // empty | select(.metadata.labels.game == "workshop1") | .metadata.id')
if [ -n "$EXISTING" ]; then
  IDS=$(printf '%s\n' "$EXISTING" | jq -R . | jq -sc .)
  curl -sf -X DELETE "$MYWANT/api/v1/wants" -H 'Content-Type: application/json' -d "{\"ids\":$IDS}" > /dev/null
  sleep 1
fi

# ── Wife target (上段中央) ────────────────────────────────────────────────────
echo "== create wife-target"
WIFE_RESP=$(api POST /api/v1/wants '{
  "metadata": {
    "name": "wife-target",
    "type": "custom_target",
    "labels": {
      "mywant.io/canvas-x": "8",
      "mywant.io/canvas-y": "-2",
      "game": "workshop1",
      "rpg-label": "妻のゴール"
    }
  },
  "spec": {
    "params": {
      "target_value": "wife_umbrella_reminder_sent",
      "label": "妻への傘リマインド"
    }
  }
}')
WIFE_ID=$(echo "$WIFE_RESP" | jq -r '.want_ids[0]')
echo "   wife-target: $WIFE_ID"

# ── Daughter target (下段中央) ───────────────────────────────────────────────
echo "== create daughter-target"
DAUGHTER_RESP=$(api POST /api/v1/wants '{
  "metadata": {
    "name": "daughter-target",
    "type": "custom_target",
    "labels": {
      "mywant.io/canvas-x": "8",
      "mywant.io/canvas-y": "6",
      "game": "workshop1",
      "rpg-label": "娘のゴール"
    }
  },
  "spec": {
    "params": {
      "target_value": "daughter_hakone_trip_booked",
      "label": "娘の箱根旅行"
    }
  }
}')
DAUGHTER_ID=$(echo "$DAUGHTER_RESP" | jq -r '.want_ids[0]')
echo "   daughter-target: $DAUGHTER_ID"

# ── Helper: create child want under a parent ─────────────────────────────────
create_child() {
  # $1=name $2=type $3=role $4=x $5=y $6=rpg-label $7=parent-id $8=parent-name $9=params_json
  api POST /api/v1/wants "{
    \"metadata\": {
      \"name\": \"$1\",
      \"type\": \"$2\",
      \"labels\": {
        \"child-role\": \"$3\",
        \"mywant.io/canvas-x\": \"$4\",
        \"mywant.io/canvas-y\": \"$5\",
        \"game\": \"workshop1\",
        \"rpg-label\": \"$6\"
      },
      \"ownerReferences\": [{
        \"apiVersion\": \"mywant/v1\", \"kind\": \"Want\",
        \"name\": \"$8\", \"id\": \"$7\",
        \"controller\": true, \"blockOwnerDeletion\": true
      }]
    },
    \"spec\": { \"params\": $9 }
  }" | jq -r '.want_ids[0]'
}

# ── Wife's chain: weather-check → remind-wife ────────────────────────────────
echo "== create wife chain (weather-check, remind-wife)"
W_CHECK_ID=$(create_child \
  weather-check weather monitor 5 -5 "今日の天気" \
  "$WIFE_ID" "wife-target" \
  '{"weather_city": "Tokyo"}')

W_REMIND_ID=$(create_child \
  remind-wife reminder doer 11 -5 "妻への連絡" \
  "$WIFE_ID" "wife-target" \
  '{"message": "雨が降りそうだよ。傘を忘れずにね。", "duration_from_now": "1 hour", "require_reaction": false}')

echo "   weather-check: $W_CHECK_ID"
echo "   remind-wife:   $W_REMIND_ID"

# ── Daughter's chain: route-search → book-hotel ──────────────────────────────
echo "== create daughter chain (route-search, book-hotel)"
D_ROUTE_ID=$(create_child \
  route-search transit monitor 5 9 "旅行の経路" \
  "$DAUGHTER_ID" "daughter-target" \
  '{"from": "Tokyo", "to": "Hakone"}')

D_BOOK_ID=$(create_child \
  book-hotel hotel doer 11 9 "宿泊予約" \
  "$DAUGHTER_ID" "daughter-target" \
  '{"hotel_type": "standard", "location": "Hakone"}')

echo "   route-search: $D_ROUTE_ID"
echo "   book-hotel:   $D_BOOK_ID"

# ── Jump RPG server to workshop1 ─────────────────────────────────────────────
echo "== jump rpg-server to workshop1"
curl -sf -X POST "$RPG/api/v1/debug/jump" \
  -H 'Content-Type: application/json' \
  -d '{"stage":"workshop1"}' | jq -c '{ok, current_stage}'

# ── Opening narration ─────────────────────────────────────────────────────────
echo "== opening narration"
mywant-gui cursor set 8 2
sleep 0.5
mywant-gui say "家族がお前を必要としている。" --duration 3
sleep 3.5
mywant-gui say "2本の連鎖を繋げ。結果がFinal Resultに現れる。" --duration 4

echo ""
echo "== Canvas layout:"
echo "   wife-target    (8,-2)  ← 妻への傘リマインド"
echo "     weather-check (5,-5)  monitor"
echo "     remind-wife  (11,-5)  doer"
echo "   daughter-target (8, 6)  ← 娘の箱根旅行"
echo "     route-search  (5, 9)  monitor"
echo "     book-hotel   (11, 9)  doer"
echo "== done."
