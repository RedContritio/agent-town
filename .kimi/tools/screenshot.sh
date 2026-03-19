#!/bin/bash
# Take screenshot of Godot Web

export PATH="$PATH:/snap/bin:/usr/local/go/bin:/home/redcontritio/.kimi/tools/venv/bin"

OUTPUT="${1:-/tmp/godot_screenshot.png}"
URL="${2:-http://localhost:8080}"

python3 /home/redcontritio/agent-town/.agents/skills/godot-web-debug/scripts/capture.py "$OUTPUT" "$URL"
