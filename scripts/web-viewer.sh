#!/bin/bash
# =============================================================================
# Web 视角观察工具 - 支持任意视角、分辨率、交互状态截图
# =============================================================================

set -e

# 获取项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# 颜色输出
info() { echo -e "→ $*"; }
success() { echo -e "✓ $*"; }
error() { echo -e "✗ $*" >&2; }

# 查找或创建 Python 虚拟环境
find_or_create_venv() {
    local venv_base="${XDG_CONFIG_HOME:-$HOME/.config}/agent-town"
    local venv_path="$venv_base/venv"
    
    if [ -f "$venv_path/bin/python" ]; then
        echo "$venv_path"
        return 0
    fi
    
    echo "→ 创建 Python 虚拟环境..." >&2
    mkdir -p "$venv_base"
    
    if command -v python3 &> /dev/null; then
        python3 -m venv "$venv_path"
    elif command -v python &> /dev/null; then
        python -m venv "$venv_path"
    else
        error "未找到 Python"
        return 1
    fi
    
    echo "$venv_path"
}

# 确保 Playwright 已安装
ensure_playwright() {
    local venv_path="$1"
    local pip="$venv_path/bin/pip"
    local python="$venv_path/bin/python"
    
    if ! $python -c "import playwright" 2>/dev/null; then
        info "安装 Playwright..."
        $pip install playwright
        $python -m playwright install chromium
    fi
}

# 显示帮助
show_help() {
    cat << 'EOF'
Web 视角观察工具 - 支持任意视角截图

用法:
    ./scripts/web-viewer.sh [选项]

常用选项:
    -o, --output <路径>      输出文件路径 (默认: /tmp/web_viewer.png)
    -u, --url <URL>          目标网址 (默认: http://localhost:8080)
    -W, --width <像素>       视口宽度 (默认: 1280)
    -H, --height <像素>      视口高度 (默认: 720)
    -f, --full-page          全页面截图
    -w, --wait <秒>          加载等待时间 (默认: 15)
    -d, --device <设备>      使用设备预设

设备预设:
    desktop, desktop-hd, desktop-4k
    ipad, ipad-pro
    iphone-se, iphone-14, iphone-14-pro
    pixel-7

交互选项:
    -a, --action <动作>      交互动作: login, click
    -t, --token <Token>      登录 Token (配合 login 动作)
    -c, --click <x,y>        点击坐标 (配合 click 动作)

序列截图:
    -S, --sequence <数量>    连续截图数量
    -i, --interval <秒>      截图间隔 (默认: 3)

其他:
    --list-devices           列出所有设备预设
    -h, --help              显示此帮助

示例:
    # 默认 1280x720 截图
    ./scripts/web-viewer.sh

    # 4K 分辨率
    ./scripts/web-viewer.sh --device desktop-4k

    # iPhone 14 视角
    ./scripts/web-viewer.sh --device iphone-14 -o /tmp/mobile.png

    # 登录后截图
    ./scripts/web-viewer.sh --action login --token "your-token-here"

    # 点击坐标后截图
    ./scripts/web-viewer.sh --action click --click 640,360 --wait 2

    # 连续截图 10 张，观察变化
    ./scripts/web-viewer.sh --sequence 10 --interval 5

    # 等待 Godot 完全加载后截图
    ./scripts/web-viewer.sh --wait 30
EOF
}

# 解析参数
PYTHON_ARGS=()

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -o|--output|-u|--url|-W|--width|-H|--height|-w|--wait|-d|--device|-a|--action|-t|--token|-c|--click|-s|--selector|-v|--selector-visible|-S|--sequence|-i|--interval)
            PYTHON_ARGS+=("$1" "$2")
            shift 2
            ;;
        -f|--full-page|--list-devices)
            PYTHON_ARGS+=("$1")
            shift
            ;;
        *)
            error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
done

# 确保虚拟环境
info "检查环境..."
VENV_PATH=$(find_or_create_venv)
ensure_playwright "$VENV_PATH"
success "环境就绪"

# 运行截图工具
info "启动 Web 观察器..."
"$VENV_PATH/bin/python" "$PROJECT_ROOT/scripts/web-viewer.py" "${PYTHON_ARGS[@]}"
