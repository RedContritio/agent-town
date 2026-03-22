#!/bin/bash
# Debug Environment Startup Script
# 调用 skill 目录中的实现

set -e

SKILL_DIR="$(dirname "$0")/../.agents/skills/godot-web-debug"

# 检查 skill 是否存在
if [ ! -f "$SKILL_DIR/bin/debug-start.sh" ]; then
    echo "Error: godot-web-debug skill not found"
    echo "Please check: $SKILL_DIR"
    exit 1
fi

# 执行 skill 中的脚本
exec "$SKILL_DIR/bin/debug-start.sh" "$@"
