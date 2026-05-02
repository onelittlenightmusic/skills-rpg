#!/usr/bin/env python3
"""rpg-next-goal: fetch the player's next suggested goal."""
import json
import os
import sys
import urllib.error
import urllib.request


def main() -> None:
    base = os.environ.get("RPG_SERVER_URL", "http://localhost:7100").rstrip("/")
    url = f"{base}/api/v1/next-goal"
    print(json.dumps({"_progress": 50, "_message": "fetching"}, ensure_ascii=False), flush=True)
    try:
        with urllib.request.urlopen(url, timeout=15) as resp:
            data = json.loads(resp.read())
    except urllib.error.URLError as e:
        print(json.dumps({"error": f"cannot reach rpg-server: {e}"}, ensure_ascii=False), flush=True)
        sys.exit(1)
    print(json.dumps({"_progress": 100, "_message": "done"}, ensure_ascii=False), flush=True)
    print(json.dumps(data, ensure_ascii=False), flush=True)


if __name__ == "__main__":
    main()
