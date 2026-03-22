#!/bin/bash
# Debug Environment Startup Script
# 只管理自己创建的进程，不影响其他服务

set -e

# 确定项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
PID_DIR="/tmp/agent-town-debug"

mkdir -p $PID_DIR

echo "========================================"
echo "  Agent Town Debug Environment"
echo "========================================"
echo ""

# 停止之前创建的进程
echo "→ Stopping previous debug processes..."
for pidfile in $PID_DIR/*.pid; do
    if [ -f "$pidfile" ]; then
        pid=$(cat "$pidfile" 2>/dev/null)
        if [ -n "$pid" ]; then
            kill $pid 2>/dev/null || true
            rm -f "$pidfile"
        fi
    fi
done
sleep 2
echo "✓ Cleaned"
echo ""

# 启动 debug_server
echo "→ Starting debug server..."
if [ ! -f "$PROJECT_ROOT/bin/debug-server" ]; then
    echo "→ Building debug server..."
    cd "$SKILL_DIR/server"
    /usr/local/go/bin/go build -o "$PROJECT_ROOT/bin/debug-server" .
fi
"$PROJECT_ROOT/bin/debug-server" > /tmp/debug-server.log 2>&1 &
echo $! > $PID_DIR/debug-server.pid
sleep 2

if curl -s http://localhost:8081/health > /dev/null; then
    echo "✓ Debug server: http://localhost:8081"
else
    echo "✗ Debug server failed to start"
    exit 1
fi
echo ""

# 启动主服务器 (如果不运行)
echo "→ Checking main server..."
if ! curl -s http://localhost:8080/api/v1/world/info > /dev/null 2>&1; then
    echo "→ Starting main server..."
    if [ ! -f "$PROJECT_ROOT/bin/server" ]; then
        echo "→ Building server..."
        cd "$PROJECT_ROOT"
        make server
    fi
    "$PROJECT_ROOT/bin/server" > /tmp/agent-town-server.log 2>&1 &
    echo $! > $PID_DIR/server.pid
    sleep 3
    if curl -s http://localhost:8080/api/v1/world/info > /dev/null; then
        echo "✓ Main server: http://localhost:8080"
    else
        echo "✗ Main server failed to start"
        exit 1
    fi
else
    echo "✓ Main server already running"
fi
echo ""

# 启动浏览器
echo "→ Launching browser..."
python3 "$SCRIPT_DIR/launch-browser.py" > /tmp/browser.log 2>&1 &
echo $! > $PID_DIR/browser.pid

echo "✓ Browser launched"
echo ""
echo "========================================"
echo "  Waiting for godot-web to connect..."
echo "========================================"

# 等待 godot 连接
for i in {1..30}; do
    if curl -s http://localhost:8081/health | grep -q '"godot_ready":true'; then
        echo ""
        echo "✓ Godot-web is ready!"
        echo ""
        echo "========================================"
        echo "  Debug Environment Ready!"
        echo "========================================"
        echo ""
        echo "Commands:"
        echo "  $SCRIPT_DIR/client.sh info"
        echo "  $SCRIPT_DIR/client.sh preset top"
        echo "  $SCRIPT_DIR/client.sh capture"
        echo ""
        echo "Stop with: $SCRIPT_DIR/debug-stop.sh"
        echo ""
        exit 0
    fi
    sleep 1
done

echo ""
echo "✗ Timeout waiting for godot-web"
echo "Check browser log: /tmp/browser.log"
echo "Check server log: /tmp/debug-server.log"
exit 1
