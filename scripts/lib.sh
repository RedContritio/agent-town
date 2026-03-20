#!/bin/bash
# =============================================================================
# Agent-Town 脚本工具库
# 提供跨脚本共享的通用函数和路径解析
# =============================================================================

# 获取脚本所在目录的绝对路径
# 用法: SCRIPT_DIR=$(get_script_dir)
get_script_dir() {
    cd "$(dirname "${BASH_SOURCE[0]}")" && pwd
}

# 获取项目根目录（脚本位于 scripts/ 目录下，父目录即为项目根目录）
# 用法: PROJECT_ROOT=$(get_project_root)
get_project_root() {
    cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd
}

# 查找可执行文件，尝试多个可能的来源
# 用法: find_executable "go" "/usr/local/go/bin:/snap/bin"
find_executable() {
    local cmd_name="$1"
    local extra_paths="${2:-}"
    local search_paths=""
    
    # 构建搜索路径（优先级：extra_paths > PATH）
    if [ -n "$extra_paths" ]; then
        search_paths="$extra_paths:$PATH"
    else
        search_paths="$PATH"
    fi
    
    # 在 PATH 中搜索
    IFS=':' read -ra PATHS <<< "$search_paths"
    for p in "${PATHS[@]}"; do
        if [ -x "$p/$cmd_name" ]; then
            echo "$p/$cmd_name"
            return 0
        fi
    done
    
    # 尝试 which
    if command -v "$cmd_name" &> /dev/null; then
        command -v "$cmd_name"
        return 0
    fi
    
    return 1
}

# 查找 Go 可执行文件
# 用法: GO=$(find_go)
find_go() {
    find_executable "go" "/usr/local/go/bin"
}

# 查找 Godot4 可执行文件
# 用法: GODOT=$(find_godot)
find_godot() {
    # 尝试多个可能的名称和位置
    for cmd in godot4 godot /snap/bin/godot4; do
        if command -v "$cmd" &> /dev/null; then
            command -v "$cmd"
            return 0
        fi
    done
    return 1
}


# 检查命令是否存在，不存在则提示安装
# 用法: require_command "go" "https://golang.org/doc/install"
require_command() {
    local cmd="$1"
    local install_url="${2:-}"
    
    if ! command -v "$cmd" &> /dev/null; then
        echo "✗ Required command '$cmd' not found" >&2
        if [ -n "$install_url" ]; then
            echo "  Install: $install_url" >&2
        fi
        return 1
    fi
    return 0
}

# 打印带颜色的状态信息
info() { echo -e "→ $*"; }
success() { echo -e "✓ $*"; }
error() { echo -e "✗ $*" >&2; }
warn() { echo -e "⚠ $*"; }
