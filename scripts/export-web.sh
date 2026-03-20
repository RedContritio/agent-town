#!/bin/bash
# =============================================================================
# 导出 Godot Web 前端
# =============================================================================

# 加载工具库
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib.sh"

PROJECT_ROOT=$(get_project_root)

# 查找 godot4
GODOT=$(find_godot)
if [ -z "$GODOT" ]; then
    error "godot4 not found. Please install Godot 4.5"
    exit 1
fi

info "Using Godot: $GODOT"

# 确保输出目录存在
mkdir -p "$PROJECT_ROOT/server/cmd/server/web"

# 导出 Web 版本
cd "$PROJECT_ROOT/godot-web"
$GODOT --headless --export-release "Web" ../server/cmd/server/web/index.html
