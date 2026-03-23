---
name: project-build
description: 构建和运行 Agent-Town 项目。安装依赖（Go 1.22+、Godot 4.5）、构建服务端/CLI/Web、启动和停止服务器。当用户要求构建、运行、安装依赖、启动服务器时触发。
---

# Agent-Town 构建指南

## 快速命令

```bash
# 安装所有依赖
./scripts/setup.sh

# 构建所有组件
make build

# 启动服务器
make run

# 停止服务器
make stop
```

## 依赖要求

| 组件 | 版本 | 说明 |
|------|------|------|
| Go | 1.22+ | 服务端和 CLI |
| Godot | 4.5 | Web 前端导出 |
| Godot Export Templates | 4.5 | Web 导出必需 |

## 构建流程

### 1. 安装依赖

```bash
# 完整安装（交互式）
./scripts/setup.sh

# 或分步安装
bash scripts/install-go.sh
bash scripts/install-godot.sh
```

### 2. 构建项目

```bash
# 构建全部（server + cli + web）
make build

# 单独构建
make server   # 仅服务端
make cli      # 仅 CLI
make web      # 仅 Web 前端（需要 godot4）
```

### 3. 运行

```bash
# 构建并运行
make run

# 访问 http://localhost:8080
```

## Go 依赖问题

如果遇到 `modernc.org/sqlite@v1.47.0 requires go >= 1.25.0` 错误：

```bash
# 方案 1：安装 Go 1.25+（如果可用）
bash scripts/install-go.sh

# 方案 2：降级 sqlite 版本
go get modernc.org/sqlite@v1.37.0
go mod tidy
```

## 调试

```bash
# 完整调试环境（推荐）
# 启动 debug-server + 主服务器 + Godot
make debug

# 仅启动 Godot（用于手动调试）
make godot-run

# 停止调试环境
make debug-stop

# 如果需要 Web 导出（用于部署）
make web
```

调试环境架构：
- Godot 直接运行（不导出），通过 HTTP 连接主服务器
- Debug Server (8081) 提供截图、相机控制等功能
- 支持热重载、断点调试等完整 Godot 功能

## 常见问题

| 问题 | 解决方案 |
|------|----------|
| `godot4: command not found` | 使用完整路径 `/snap/bin/godot4` 或添加到 PATH |
| `go: command not found` | 确保 `/usr/local/go/bin` 在 PATH 中 |
| 端口 8080 被占用 | `make stop` 或 `pkill -f bin/server` |
| Godot 无法连接 | 检查图形环境 `echo $DISPLAY`，或尝试 `godot4 --headless` |
| Web 导出不工作 | 检查 Godot 导出模板是否安装 |

## 验证

构建成功标志：
- `bin/server` 二进制文件存在
- `bin/cli` 二进制文件存在（如果构建了 CLI）
- `server/cmd/server/web/index.html` 存在

服务器运行验证：
```bash
curl http://localhost:8080/api/v1/world/info
# 应返回 JSON 世界信息
```
