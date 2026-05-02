#!/usr/bin/env bash
# stage1 の最短プレイスルーをエンドツーエンドで検証するスモークテスト。
# - rpg-server を一時データディレクトリで起動
# - 観測 → you の open 失敗 → chap の open 成功 → 移動 でクリアまで進む
# - セーブとロードも確認
set -euo pipefail

PORT="${RPG_TEST_PORT:-17100}"
URL="http://localhost:${PORT}"
DATA="$(mktemp -d)"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

cleanup() {
  if [[ -n "${SERVER_PID:-}" ]]; then kill "$SERVER_PID" 2>/dev/null || true; wait 2>/dev/null || true; fi
  rm -rf "$DATA"
}
trap cleanup EXIT

cd "$ROOT"
go build -o bin/rpg-server ./cmd/rpg-server >/dev/null

RPG_DATA_DIR="$DATA" ./bin/rpg-server --stages-dir stages --port "$PORT" >"$DATA/server.log" 2>&1 &
SERVER_PID=$!
for _ in $(seq 1 30); do
  if curl -sf "$URL/healthz" >/dev/null 2>&1; then break; fi
  sleep 0.1
done

step() { echo; echo "### $1"; }
post() { curl -sf -X POST "$URL$1" -H 'content-type: application/json' -d "$2"; echo; }
get()  { curl -sf "$URL$1"; echo; }
fail() { echo "FAIL: $1" >&2; exit 1; }

step "0. initial next-goal (look_around)"
get /api/v1/next-goal | tee "$DATA/0.json"
grep -q "見回そう" "$DATA/0.json" || fail "expected 'look around' goal"

step "1. observe — should unlock look_around and shift goal"
get "/api/v1/observe?target=stages.stage1.doors.door1" >"$DATA/1.json"
grep -q "look_around" "$DATA/1.json" || fail "expected look_around in achievements_unlocked"

step "2. you tries open door1 — must be rejected (409)"
code=$(curl -s -o "$DATA/2.json" -w '%{http_code}' \
  -X POST "$URL/api/v1/control" -H 'content-type: application/json' \
  -d '{"actor":"you","action":"open","target":"door1"}')
[[ "$code" == "409" ]] || fail "expected 409 for you-open, got $code"
grep -q "attempted_self_unlock" "$DATA/2.json" || fail "expected attempted_self_unlock unlock"

step "3. save current progress to slot-1"
post /api/v1/saves/slot-1 '{"name":"after self-unlock attempt"}' >"$DATA/3.json"

step "4. chap opens door1"
post /api/v1/control '{"actor":"chap","action":"open","target":"door1"}' >"$DATA/4.json"
grep -q "chap_unlocked_door1" "$DATA/4.json" || fail "expected chap_unlocked_door1"

step "5. you moves to room2 — should clear stage1"
post /api/v1/control '{"actor":"you","action":"move","target":"room2"}' >"$DATA/5.json"
grep -q "escaped_room1" "$DATA/5.json" || fail "expected escaped_room1"
grep -q "cleared" "$DATA/5.json" || fail "expected cleared next_goal"

step "6. load slot-1 — should restore pre-unlock state"
post /api/v1/saves/slot-1/load 'null' >"$DATA/6.json"
get /api/v1/state >"$DATA/6_state.json"
python3 - <<PY "$DATA/6_state.json"
import json,sys
s=json.load(open(sys.argv[1]))
door=s['stages']['stage1']['doors']['door1']
assert door['open'] is False, f"expected door closed after load, got {door}"
assert s['you']['position']=='room1', f"expected position room1, got {s['you']['position']}"
print("load restored pre-unlock state OK")
PY

step "7. list saves"
get /api/v1/saves >"$DATA/7.json"
grep -q "slot-1" "$DATA/7.json" || fail "expected slot-1 in list"

echo
echo "ALL OK"
