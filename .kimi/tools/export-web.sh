#!/bin/bash
# Export Godot Web frontend

export PATH="$PATH:/snap/bin:/usr/local/go/bin"
cd /home/redcontritio/agent-town/godot-web
godot4 --headless --export-release "Web" ../server/cmd/server/web/index.html
