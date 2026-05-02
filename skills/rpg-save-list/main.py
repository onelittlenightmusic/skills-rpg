#!/usr/bin/env python3
"""rpg-save-list: list save slots."""
import json
import os
import sys
import urllib.error
import urllib.request


def main() -> None:
    base = os.environ.get("RPG_SERVER_URL", "http://localhost:7100").rstrip("/")
    url = f"{base}/api/v1/saves"
    try:
        with urllib.request.urlopen(url, timeout=15) as resp:
            data = json.loads(resp.read())
    except urllib.error.URLError as e:
        print(json.dumps({"error": f"cannot reach rpg-server: {e}", "slots": []}, ensure_ascii=False), flush=True)
        sys.exit(1)
    print(json.dumps(data, ensure_ascii=False), flush=True)


if __name__ == "__main__":
    main()
