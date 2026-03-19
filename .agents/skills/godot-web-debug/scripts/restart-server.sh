#!/bin/bash
# 重启 Godot Web 服务器

PORT="${1:-8080}"
PID_FILE="/tmp/godot-web-server.pid"

echo "→ 停止现有服务器..."
if [ -f "$PID_FILE" ]; then
    PID=$(cat "$PID_FILE" 2>/dev/null) || true
    if [ -n "$PID" ] && kill -0 "$PID" 2>/dev/null; then
        kill "$PID" 2>/dev/null || true
        sleep 1
    fi
    rm -f "$PID_FILE"
fi
pkill -f "bin/server" 2>/dev/null || true
sleep 1

echo "→ 构建并启动服务器（端口 $PORT）..."
cd "$(dirname "$0")/../../../.."
export PATH=$PATH:/usr/local/go/bin:/snap/bin

make build >/dev/null 2>&1 || go build -o bin/server ./server/cmd/server

./bin/server &
NEW_PID=$!
echo $NEW_PID > "$PID_FILE"

sleep 2

if kill -0 "$NEW_PID" 2>/dev/null; then
    echo "✓ 服务器运行中 http://localhost:$PORT (PID: $NEW_PID)"
else
    echo "✗ 服务器启动失败"
    exit 1
fi
