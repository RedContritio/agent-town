#!/bin/bash
# =============================================================================
# 重启 Godot Web 服务器（AI 调试工具）
# =============================================================================

# 加载技能库
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../lib.sh"

PROJECT_ROOT=$(get_project_root)

PORT="${1:-8080}"
PID_FILE="/tmp/godot-web-server.pid"

info "停止现有服务器..."
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

info "构建并启动服务器（端口 $PORT）..."
cd "$PROJECT_ROOT"

# 动态查找 go
GO=$(find_executable "go" "/usr/local/go/bin")
if [ -z "$GO" ]; then
    error "Go not found. Please install Go 1.22+"
    exit 1
fi

# 优先使用 make，否则直接编译
if [ -f "Makefile" ]; then
    make build >/dev/null 2>&1 || "$GO" build -o bin/server ./server/cmd/server
else
    "$GO" build -o bin/server ./server/cmd/server
fi

./bin/server &
NEW_PID=$!
echo $NEW_PID > "$PID_FILE"

sleep 2

if kill -0 "$NEW_PID" 2>/dev/null; then
    success "服务器运行中 http://localhost:$PORT (PID: $NEW_PID)"
else
    error "服务器启动失败"
    exit 1
fi
