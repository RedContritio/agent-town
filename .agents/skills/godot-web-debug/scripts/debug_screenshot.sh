#!/bin/bash
# 多角度调试截图工具
#
# 用法：
#   debug_screenshot.sh <output.png> [command] [params]
#
# 命令：
#   zoom <delta>     - 滚轮缩放（负数缩小，正数放大）
#   drag <dx> <dy>   - 拖拽旋转相机
#   key <key>        - 按键（如 1, 2, 3, +, -）
#   click <x> <y>    - 点击
#   scroll <x> <y> <delta> - 滚动到位置
#
# 示例：
#   debug_screenshot.sh /tmp/shot.png                    # 默认截图
#   debug_screenshot.sh /tmp/zoom_out.png zoom -200     # 缩小
#   debug_screenshot.sh /tmp/rotate.png drag 200 100    # 拖拽旋转
#   debug_screenshot.sh /tmp/building.png key 1          # 按1键

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

python3 "$SCRIPT_DIR/capture_debug.py" "$@"