#!/bin/bash
set -e

# Start the server in the background
./bin/mywant-rpg server start > /tmp/rpg-smoke.log 2>&1 &
PID=$!
trap 'kill $PID' EXIT

# Wait for server to start
sleep 3

# Check if process is running
if ! ps -p $PID > /dev/null; then
    echo "Server failed to start. Logs:"
    cat /tmp/rpg-smoke.log
    exit 1
fi

# Make a request to verify it's working
if curl -s http://localhost:7100/api/v1/start | grep -q "agent_role"; then
    echo "Smoke test passed!"
else
    echo "Server did not respond correctly. Logs:"
    cat /tmp/rpg-smoke.log
    exit 1
fi
