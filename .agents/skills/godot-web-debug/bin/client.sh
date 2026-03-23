#!/bin/bash
#
# Debug Client - 使用 curl 调用 debug_server HTTP API
#
# 用法:
#   ./client.sh info                    # 获取相机信息
#   ./client.sh set [选项]              # 设置相机参数（轨道模式）
#   ./client.sh direct [选项]           # 直接设置相机位置（用于平视视角）
#   ./client.sh preset <name>           # 使用预设视角
#   ./client.sh capture [选项]          # 截图
#
# 示例:
#   ./client.sh info
#   ./client.sh set -x 0 -z 0 -d 30 -a 45 -p 60
#   ./client.sh direct -X -12 -Y 3 -Z -20 -x -12 -y 1.5 -z -12
#   ./client.sh preset top
#   ./client.sh capture

SERVER_URL="http://localhost:8081"

# 显示用法
usage() {
    echo "Debug Client - 控制 Godot 相机"
    echo ""
    echo "用法:"
    echo "  $0 info                 # 获取相机信息"
    echo "  $0 set [选项]           # 设置相机参数（轨道模式）"
    echo "  $0 direct [选项]        # 直接设置相机位置（用于平视视角）"
    echo "  $0 preset <name>        # 使用预设视角 (top|side|north|south|east|west)"
    echo "  $0 capture [选项]       # 截图 (保存到 Godot 主机的 /tmp/)"
    echo ""
    echo "set 选项:"
    echo "  -x X    目标位置 X (默认: 0)"
    echo "  -y Y    目标位置 Y (默认: 1.5)"
    echo "  -z Z    目标位置 Z (默认: 0)"
    echo "  -d D    相机距离 (默认: 25)"
    echo "  -a A    水平角度 (默认: -45)"
    echo "  -p P    垂直角度 (默认: 60)"
    echo ""
    echo "direct 选项 (直接设置相机位置，绕过轨道控制器):"
    echo "  -X X    相机位置 X (默认: 0)"
    echo "  -Y Y    相机位置 Y/高度 (默认: 3)"
    echo "  -Z Z    相机位置 Z (默认: 0)"
    echo "  -x X    看向位置 X (默认: 0)"
    echo "  -y Y    看向位置 Y (默认: 1.5)"
    echo "  -z Z    看向位置 Z (默认: 0)"
    echo ""
    echo "capture 选项:"
    echo "  -W W    宽度 (默认: 1280)"
    echo "  -H H    高度 (默认: 720)"
    echo ""
    echo "示例:"
    echo "  $0 preset top"
    echo "  $0 set -x -12 -z -12 -d 15 -a 0 -p 45"
    echo "  $0 direct -X -12 -Y 3 -Z -20 -x -12 -y 1.5 -z -12  # 平视 Gov Hall 北面"
    echo "  $0 capture -W 1920 -H 1080"
    exit 1
}

# 检查 godot 是否 ready
check_godot() {
    local status=$(curl -s "$SERVER_URL/health" | grep -o '"godot_ready":true' || echo "")
    if [ -z "$status" ]; then
        echo "Error: Godot not ready"
        echo "Please ensure Godot is running and connected to debug_server"
        exit 1
    fi
}

# 获取相机信息
cmd_info() {
    check_godot
    echo "Camera Info:"
    curl -s "$SERVER_URL/api/info" | python3 -m json.tool 2>/dev/null || curl -s "$SERVER_URL/api/info"
    echo ""
}

# 设置相机
cmd_set() {
    local x=0 y=1.5 z=0 distance=25 azimuth=-45 polar=60
    
    while getopts "x:y:z:d:a:p:h" opt; do
        case $opt in
            x) x=$OPTARG ;;
            y) y=$OPTARG ;;
            z) z=$OPTARG ;;
            d) distance=$OPTARG ;;
            a) azimuth=$OPTARG ;;
            p) polar=$OPTARG ;;
            h) usage ;;
            *) usage ;;
        esac
    done
    
    check_godot
    
    local payload=$(cat <<EOF
{
    "target": [$x, $y, $z],
    "distance": $distance,
    "azimuth": $azimuth,
    "polar": $polar
}
EOF
)
    
    echo "Setting camera: target=[$x, $y, $z], distance=$distance, azimuth=$azimuth, polar=$polar"
    curl -s -X POST -H "Content-Type: application/json" -d "$payload" "$SERVER_URL/api/camera"
    echo ""
}

# 使用预设
cmd_preset() {
    if [ -z "$1" ]; then
        echo "Error: preset name required"
        echo "Available: top, side, north, south, east, west"
        exit 1
    fi
    
    check_godot
    
    echo "Applying preset: $1"
    curl -s -X POST -H "Content-Type: application/json" -d "{\"name\":\"$1\"}" "$SERVER_URL/api/preset"
    echo ""
}

# 直接设置相机位置（绕过轨道控制器）
cmd_direct() {
    local pos_x=0 pos_y=3 pos_z=0 look_x=0 look_y=1.5 look_z=0
    
    while getopts "X:Y:Z:x:y:z:h" opt; do
        case $opt in
            X) pos_x=$OPTARG ;;
            Y) pos_y=$OPTARG ;;
            Z) pos_z=$OPTARG ;;
            x) look_x=$OPTARG ;;
            y) look_y=$OPTARG ;;
            z) look_z=$OPTARG ;;
            h) usage ;;
            *) usage ;;
        esac
    done
    
    check_godot
    
    local payload=$(cat <<EOF
{
    "position": [$pos_x, $pos_y, $pos_z],
    "look_at": [$look_x, $look_y, $look_z]
}
EOF
)
    
    echo "Setting camera directly: position=[$pos_x, $pos_y, $pos_z] -> look_at=[$look_x, $look_y, $look_z]"
    curl -s -X POST -H "Content-Type: application/json" -d "$payload" "$SERVER_URL/api/camera/direct"
    echo ""
}

# 截图
cmd_capture() {
    local width=1280 height=720
    
    while getopts "W:H:h" opt; do
        case $opt in
            W) width=$OPTARG ;;
            H) height=$OPTARG ;;
            h) usage ;;
            *) usage ;;
        esac
    done
    
    check_godot
    
    echo "Capturing screenshot (${width}x${height})..."
    local result=$(curl -s -X POST -H "Content-Type: application/json" -d "{\"width\":$width,\"height\":$height}" "$SERVER_URL/api/capture")
    
    # 解析 filepath
    local filepath=$(echo "$result" | python3 -c "import sys,json; print(json.load(sys.stdin).get('filepath',''))" 2>/dev/null)
    
    if [ -n "$filepath" ]; then
        echo "Screenshot saved: $filepath"
        echo "Size: ${width}x${height}"
    else
        echo "Response: $result"
    fi
}

# 主命令
case "$1" in
    info)
        cmd_info
        ;;
    set)
        shift
        cmd_set "$@"
        ;;
    direct)
        shift
        cmd_direct "$@"
        ;;
    preset)
        shift
        cmd_preset "$@"
        ;;
    capture)
        shift
        cmd_capture "$@"
        ;;
    *)
        usage
        ;;
esac
