#!/bin/bash
# 重启服务器并捕获截图

OUTPUT="${1:-/tmp/godot_screenshot.png}"
SCRIPT_DIR="$(dirname "$0")"

echo "=== Godot Web 调试：重启并截图 ==="
echo ""

# 重启服务器
"$SCRIPT_DIR/restart-server.sh"
if [ $? -ne 0 ]; then
    echo "✗ 服务器重启失败"
    exit 1
fi

echo ""

# 捕获截图
/tmp/playwright-venv/bin/python "$SCRIPT_DIR/capture.py" "$OUTPUT"

echo ""
echo "=== 完成 ==="
echo "截图：$OUTPUT"
echo "查看方式：ReadMediaFile(\"$OUTPUT\")"
