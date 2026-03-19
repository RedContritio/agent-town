#!/bin/bash
# Development helper commands

SERVER_PORT=8080
SERVER_LOG=/tmp/agent-town-server.log

case "$1" in
    status)
        if curl -s "http://localhost:$SERVER_PORT/api/v1/world/info" > /dev/null 2>&1; then
            echo "✓ Server running at http://localhost:$SERVER_PORT"
        else
            echo "✗ Server not running"
        fi
        ;;
    logs)
        if [ -f "$SERVER_LOG" ]; then
            tail -f "$SERVER_LOG"
        else
            echo "No log file found at $SERVER_LOG"
        fi
        ;;
    test)
        echo "→ Testing API..."
        curl -s "http://localhost:$SERVER_PORT/api/v1/world/info" 2>/dev/null | head -c 200 || echo "✗ Server not responding"
        echo ""
        ;;
    env)
        echo "=== Environment ==="
        echo "Go:     $(which go 2>/dev/null || echo 'not found') $(go version 2>/dev/null | awk '{print $3}')"
        echo "Godot:  $(which godot4 2>/dev/null || echo 'not found') $(godot4 --version 2>/dev/null | head -1)"
        echo "Web:    $(test -f server/cmd/server/web/index.html && echo 'found' || echo 'not built')"
        ;;
    *)
        echo "Usage: $0 {status|logs|test|env}"
        echo ""
        echo "  status  # Check server status"
        echo "  logs    # View server logs"
        echo "  test    # Test API endpoint"
        echo "  env     # Show environment info"
        ;;
esac
