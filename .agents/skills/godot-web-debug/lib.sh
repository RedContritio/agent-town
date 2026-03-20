#!/bin/bash
# =============================================================================
# Godot Web Debug 技能 - 工具库
# 为 AI 调试工具提供共享函数（独立于项目 scripts/lib.sh）
# =============================================================================

# 获取此库文件所在目录
SKILL_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 获取技能根目录
get_skill_root() {
    echo "$SKILL_LIB_DIR"
}

# 获取项目根目录（技能位于 .agents/skills/godot-web-debug/）
get_project_root() {
    cd "$SKILL_LIB_DIR/../../../.." && pwd
}

# 打印带颜色的状态信息
info() { echo -e "→ $*"; }
success() { echo -e "✓ $*"; }
error() { echo -e "✗ $*" >&2; }
warn() { echo -e "⚠ $*"; }

# 查找可执行文件
find_executable() {
    local cmd_name="$1"
    local extra_paths="${2:-}"
    local search_paths=""
    
    if [ -n "$extra_paths" ]; then
        search_paths="$extra_paths:$PATH"
    else
        search_paths="$PATH"
    fi
    
    IFS=':' read -ra PATHS <<< "$search_paths"
    for p in "${PATHS[@]}"; do
        if [ -x "$p/$cmd_name" ]; then
            echo "$p/$cmd_name"
            return 0
        fi
    done
    
    if command -v "$cmd_name" &> /dev/null; then
        command -v "$cmd_name"
        return 0
    fi
    
    return 1
}

# 查找项目 scripts/lib.sh 并加载（如果存在）
load_project_lib() {
    local project_lib="$(get_project_root)/scripts/lib.sh"
    if [ -f "$project_lib" ]; then
        source "$project_lib"
        return 0
    fi
    return 1
}

# 查找或创建 Python 虚拟环境（用于 Playwright）
# 优先：~/.config/agent-town/venv
find_or_create_venv() {
    local venv_base="${XDG_CONFIG_HOME:-$HOME/.config}/agent-town"
    local venv_path="$venv_base/venv"
    
    if [ -f "$venv_path/bin/python" ]; then
        echo "$venv_path"
        return 0
    fi
    
    echo "→ Creating Python virtual environment at $venv_path..." >&2
    mkdir -p "$venv_base"
    
    if command -v python3 &> /dev/null; then
        python3 -m venv "$venv_path"
    elif command -v python &> /dev/null; then
        python -m venv "$venv_path"
    else
        error "Python not found"
        return 1
    fi
    
    if [ -f "$venv_path/bin/python" ]; then
        echo "$venv_path"
        return 0
    fi
    
    return 1
}

# 确保 Playwright 已安装
ensure_playwright() {
    local venv_path
    venv_path=$(find_or_create_venv)
    
    if [ -z "$venv_path" ]; then
        return 1
    fi
    
    local pip="$venv_path/bin/pip"
    local python="$venv_path/bin/python"
    
    if ! $python -c "import playwright" 2>/dev/null; then
        info "Installing Playwright..."
        $pip install playwright
        $python -m playwright install chromium
    fi
    
    echo "$venv_path"
}
