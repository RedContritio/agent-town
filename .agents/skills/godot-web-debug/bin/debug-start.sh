#!/bin/bash
# Debug Environment Startup Script (Godot Native Mode)
# 启动 Godot 编辑器/运行时直接连接调试，无需 Web 导出

set -e

# 确定项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../../.." && pwd)"
PID_DIR="/tmp/agent-town-debug"

mkdir -p $PID_DIR

echo "========================================"
echo "  Agent Town Debug Environment"
echo "  (Godot Native Mode)"
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
# 同时检查并停止可能残留的 godot 进程
pkill -f "godot4.*godot-web" 2>/dev/null || true
sleep 2
echo "✓ Cleaned"
echo ""

# 启动 debug_server
echo "→ Starting debug server..."
if [ ! -f "$PROJECT_ROOT/bin/debug-server" ]; then
    echo "→ Building debug server..."
    cd "$PROJECT_ROOT/.agents/skills/godot-web-debug/server"
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

# 启动 Godot
echo "→ Starting Godot..."
cd "$PROJECT_ROOT/godot-web"
godot4 . > /tmp/godot.log 2>&1 &
echo $! > $PID_DIR/godot.pid

echo "✓ Godot started"
echo ""
echo "========================================"
echo "  Waiting for Godot to connect..."
echo "========================================"

# 等待 godot 连接
for i in {1..30}; do
    if curl -s http://localhost:8081/health | grep -q '"godot_ready":true'; then
        echo ""
        echo "✓ Godot is ready!"
        echo ""
        echo "========================================"
        echo "  Debug Environment Ready!"
        echo "========================================"
        echo ""
        echo "Commands:"
        echo "  $SCRIPT_DIR/client.sh info                    # 获取相机信息"
        echo "  $SCRIPT_DIR/client.sh preset top              # 设置预设视角"
        echo "  $SCRIPT_DIR/client.sh direct -X -12 -Y 3 -Z -20 -x -12 -y 1.5 -z -12  # 平视视角"
        echo "  $SCRIPT_DIR/client.sh capture                 # 截图"
        echo ""
        echo "Stop with: $SCRIPT_DIR/debug-stop.sh"
        echo ""
        exit 0
    fi
    sleep 1
done

echo ""
echo "✗ Timeout waiting for Godot"
echo "Check Godot log: /tmp/godot.log"
echo "Check debug server log: /tmp/debug-server.log"
exit 1
