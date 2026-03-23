#!/bin/bash
# Debug Environment Stop Script
# 停止由 debug-start.sh 创建的所有进程

PID_DIR="/tmp/agent-town-debug"

echo "Stopping debug environment..."

# 停止所有由 debug-start.sh 创建的进程
for pidfile in $PID_DIR/*.pid; do
    if [ -f "$pidfile" ]; then
        name=$(basename "$pidfile" .pid)
        pid=$(cat "$pidfile" 2>/dev/null)
        if [ -n "$pid" ]; then
            echo "  Stopping PID $pid ($name)"
            kill $pid 2>/dev/null || true
        fi
        rm -f "$pidfile"
    fi
done

# 同时尝试停止 Godot 进程（如果 PID 文件丢失）
if pgrep -f "godot4.*godot-web" > /dev/null; then
    echo "  Stopping Godot processes"
    pkill -f "godot4.*godot-web" 2>/dev/null || true
fi

echo "✓ Debug environment stopped"
echo "Note: Main server (if not started by debug-start.sh) continues running"
