#!/bin/bash
# =============================================================================
# 捕获 Godot Web 截图（AI 调试工具）
# =============================================================================

# 加载技能库
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/../lib.sh"

OUTPUT="${1:-/tmp/godot_screenshot.png}"
URL="${2:-http://localhost:8080}"

# 确保 Playwright 虚拟环境已准备
VENV_PATH=$(ensure_playwright)
if [ -z "$VENV_PATH" ]; then
    error "无法准备 Playwright 环境"
    exit 1
fi

# 查找 capture.py
CAPTURE_SCRIPT="$SCRIPT_DIR/capture.py"

if [ ! -f "$CAPTURE_SCRIPT" ]; then
    error "Capture script not found at $CAPTURE_SCRIPT"
    exit 1
fi

info "使用虚拟环境: $VENV_PATH"
"$VENV_PATH/bin/python" "$CAPTURE_SCRIPT" "$OUTPUT" "$URL"
