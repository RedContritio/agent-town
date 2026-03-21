#!/bin/bash
# 验证 Agent-Town 构建是否成功

set -e

echo "=== 验证 Agent-Town 构建 ==="
echo ""

ERRORS=0

# 检查 Go
echo -n "Go: "
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo "✓ $GO_VERSION"
else
    echo "✗ 未安装"
    ERRORS=$((ERRORS+1))
fi

# 检查 Godot
echo -n "Godot: "
if command -v godot4 &> /dev/null || [ -x "/snap/bin/godot4" ]; then
    GODOT_VERSION=$(/snap/bin/godot4 --version 2>/dev/null | head -1 || echo "unknown")
    echo "✓ $GODOT_VERSION"
else
    echo "✗ 未安装"
    ERRORS=$((ERRORS+1))
fi

# 检查 server 二进制
echo -n "Server: "
if [ -x "bin/server" ]; then
    echo "✓ bin/server"
else
    echo "✗ 未构建（运行 make server）"
    ERRORS=$((ERRORS+1))
fi

# 检查 web 导出
echo -n "Web: "
if [ -f "server/cmd/server/web/index.html" ]; then
    echo "✓ server/cmd/server/web/index.html"
else
    echo "✗ 未导出（运行 make web）"
    ERRORS=$((ERRORS+1))
fi

# 检查服务器运行状态
echo -n "Server running: "
if curl -s http://localhost:8080/api/v1/world/info &>/dev/null; then
    echo "✓ 是"
else
    echo "✗ 否（运行 make run）"
    ERRORS=$((ERRORS+1))
fi

echo ""
if [ $ERRORS -eq 0 ]; then
    echo "✓ 所有检查通过"
    exit 0
else
    echo "✗ $ERRORS 项检查失败"
    exit 1
fi
