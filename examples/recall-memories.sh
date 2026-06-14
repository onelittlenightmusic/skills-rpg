#!/bin/sh
# Stage 1: Family & Memories (Recall)
set -u
MYWANT=${MYWANT_URL:-http://localhost:8080}
RPG=${RPG_URL:-http://localhost:7100}

# Monitoring loop
while true; do
  WANTS=$(curl -sf "$MYWANT/api/v1/wants")
  
  # Check connections using jq for precision
  # 1. weather-check -> remind-wife
  WIFE_CONN=$(echo "$WANTS" | jq '.wants[] | select(.metadata.name=="remind-wife") | .spec.imports | length > 0')
  # 2. route-search -> book-hotel
  DAUGHTER_CONN=$(echo "$WANTS" | jq '.wants[] | select(.metadata.name=="book-hotel") | .spec.imports | length > 0')

  if [ "$WIFE_CONN" = "true" ] && [ "$DAUGHTER_CONN" = "true" ]; then
    echo "Memory recall successful: Family bonds restored!"
    # Trigger RPG success
    curl -s -X POST "$RPG/api/v1/control" -H 'Content-Type: application/json' -d '{"actor":"chap","action":"activate","target":"core"}'
    
    # Update Core Self with final story message
    CORE_ID=$(echo "$WANTS" | jq -r '.wants[] | select(.metadata.name=="core-self") | .metadata.id')
    if [ -n "$CORE_ID" ]; then
      curl -s -X PUT "$MYWANT/api/v1/states/$CORE_ID" -H 'Content-Type: application/json' -d '{"final_result":"思い出しました。私は家族を守るエンジニアです。\n妻へのリマインド、娘との旅行の約束……すべて繋がりました。\n\n愛する人たちが待っている。ここで立ち止まっている暇はありません。"}'
    fi
    exit 0
  fi
  sleep 4
done
