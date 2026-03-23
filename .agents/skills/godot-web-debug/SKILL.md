# Godot Debug Skill

用于调试 Godot 项目的工具集，提供截图、相机控制等功能。Godot 直接运行，通过 WebSocket 与调试服务器通信。

## 架构

```
┌─────────────┐      HTTP      ┌─────────────┐     WebSocket    ┌─────────────┐
│   Client    │ ─────────────→ │ Debug Server│ ←──────────────→ │ Godot Editor │
│  (client.sh)│                │   (:8081)   │                  │ (Native Run) │
└─────────────┘                └─────────────┘                  └─────────────┘
                                       ↑
                                       │ HTTP (:8080)
                                       ↓
                               ┌─────────────┐
                               │  Main Server│
                               │   (Go)      │
                               └─────────────┘
```

## 安装

### 1. 安装依赖

```bash
# Go 1.22+
/usr/local/go/bin/go version

# Godot 4.5
godot4 --version
```

### 2. 构建

```bash
# 进入 skill 目录
cd .agents/skills/godot-web-debug

# 构建 debug server
cd server
/usr/local/go/bin/go build -o ../bin/debug-server .
cd ..

# 确保脚本可执行
chmod +x bin/*.sh bin/*.py
```

## 快速开始

### 启动完整调试环境

```bash
# 使用启动脚本（推荐）
./bin/debug-start.sh

# 或手动启动各组件
./bin/debug-server &          # 1. 启动 debug server (端口 8081)
make run                      # 2. 启动主服务器 (端口 8080)
cd godot-web && godot4 .      # 3. 启动 Godot
```

### 停止环境

```bash
./bin/debug-stop.sh
```

## 脚本说明

### 服务器端

| 脚本 | 用途 | 端口 |
|------|------|------|
| `bin/debug-server` | WebSocket/HTTP 桥接服务器 | 8081 |
| `bin/client.sh` | HTTP 客户端 (Bash) | - |
| `bin/client.py` | HTTP 客户端 (Python，功能更完整) | - |

### 工作流脚本

| 脚本 | 用途 |
|------|------|
| `bin/debug-start.sh` | 一键启动完整环境（debug-server + main-server + Godot） |
| `bin/debug-stop.sh` | 停止所有 debug 相关进程 |

## 使用示例

### 1. 截图

```bash
# 先启动环境
./bin/debug-start.sh

# 等待 Godot 连接完成（约 5-10 秒）

# 发送截图命令
./bin/client.sh capture

# 或使用 Python 客户端
./bin/client.py capture
```

### 2. 获取相机信息

```bash
./bin/client.sh info
```

输出示例：
```json
{
  "target": [0.0, 1.5, 0.0],
  "distance": 25.0,
  "azimuth_deg": -45.0,
  "polar_deg": 60.0
}
```

### 3. 设置预设视角

```bash
./bin/client.sh preset top     # 俯视
./bin/client.sh preset side    # 侧面
./bin/client.sh preset north   # 北面
```

### 4. 自定义相机位置

轨道模式（第三人称视角）：
```bash
./bin/client.sh set -x 0 -z 0 -d 30 -a 0 -p 45
```

直接设置模式（用于平视视角，绕过轨道控制器限制）：
```bash
./bin/client.sh direct -X -12 -Y 3 -Z -20 -x -12 -y 1.5 -z -12
```
参数说明：
- `-X, -Y, -Z`: 相机位置（大写）
- `-x, -y, -z`: 看向位置（小写）

### 5. 平视视角截图示例

拍摄 Gov Hall 四个方向的平视视角（相机高度与墙齐平）：

```bash
# 从北边看（看向南）
./bin/client.sh direct -X -12 -Y 3 -Z -20 -x -12 -y 1.5 -z -12
./bin/client.sh capture

# 从南边看（看向北）
./bin/client.sh direct -X -12 -Y 3 -Z -4 -x -12 -y 1.5 -z -12
./bin/client.sh capture

# 从东边看（看向西）
./bin/client.sh direct -X -4 -Y 3 -Z -12 -x -12 -y 1.5 -z -12
./bin/client.sh capture

# 从西边看（看向东）
./bin/client.sh direct -X -20 -Y 3 -Z -12 -x -12 -y 1.5 -z -12
./bin/client.sh capture
```

## API 端点

Debug Server 提供以下 HTTP API：

| 端点 | 方法 | 描述 |
|------|------|------|
| `/health` | GET | 健康检查，返回 Godot 连接状态 |
| `/api/camera` | POST | 设置相机参数（轨道模式） |
| `/api/camera/direct` | POST | 直接设置相机位置（用于平视视角） |
| `/api/preset` | POST | 设置预设视角 |
| `/api/capture` | POST | 截图 |
| `/api/info` | GET | 获取相机信息 |
| `/godot` | WebSocket | Godot 客户端连接 |

## 故障排查

### Godot 未连接

**检查步骤**：

```bash
# 1. 检查 debug-server 是否运行
curl http://localhost:8081/health

# 2. 检查 Godot 是否运行
ps aux | grep godot

# 3. 检查 Godot 控制台日志
tail -f /tmp/godot.log
```

### 端口占用

```bash
# 检查 8081 端口
lsof -i :8081

# 释放端口
pkill -f debug-server
```

## 文件结构

```
.agents/skills/godot-web-debug/
├── SKILL.md              # 本文件
├── bin/
│   ├── debug-start.sh    # 启动完整环境（Godot 原生模式）
│   ├── debug-stop.sh     # 停止环境
│   ├── client.sh         # HTTP 客户端 (Bash)
│   ├── client.py         # HTTP 客户端 (Python，功能更完整)
│   └── debug-server      # 编译后的服务器 (二进制)
├── server/
│   ├── main.go           # Debug server 源码
│   └── go.mod            # Go 依赖
└── scripts/              # 其他调试脚本
    └── capture.py
```

## 运行模式说明

本 skill 使用 **Godot 原生运行模式**，特点：

| 特性 | 说明 |
|------|------|
| 启动速度 | 快（无需 Web 导出） |
| 调试工具 | 完整（断点、Profiler） |
| 热重载 | 支持 |
| 截图功能 | 需要 X11 显示环境 |
| 适用场景 | 日常开发调试、自动化测试 |

### 无头模式（Headless）

在没有显示环境的服务器上，可以使用 Xvfb 虚拟显示：

```bash
# 启动虚拟显示
Xvfb :99 -screen 0 1280x720x24 -ac &
export DISPLAY=:99

# 启动 Godot
cd godot-web && godot4 . &
```

## 注意事项

1. **Godot 版本**：需要 Godot 4.5+
2. **图形环境**：
   - 本地运行：需要 X11/Wayland 显示环境
   - 服务器运行：使用 Xvfb 虚拟显示（详见上方说明）
3. **进程管理**：`debug-stop.sh` 只停止由 `debug-start.sh` 创建的进程
4. **日志位置**：
   - Godot 日志：`/tmp/godot.log`
   - Debug server 日志：`/tmp/debug-server.log`
5. **截图保存**：截图保存在 Godot 主机的 `/tmp/` 目录下

## 相关文档

- `../../AGENTS.md` - 项目整体架构
- `../../../docs/VISUAL_DESIGN.md` - 视觉设计规范
