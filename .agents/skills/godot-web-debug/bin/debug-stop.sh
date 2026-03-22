#!/bin/bash
# Debug Environment Stop Script
# 只停止由 debug-start.sh 创建的进程

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

echo "✓ Debug environment stopped"
echo "Note: Main server and other services are not affected"
