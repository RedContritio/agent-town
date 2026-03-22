# Godot Web Debug Skill

用于调试 Godot Web 导出的工具集，提供截图、相机控制等功能。

## 架构

```
┌─────────────┐      HTTP      ┌─────────────┐     WebSocket    ┌─────────────┐
│   Client    │ ─────────────→ │ Debug Server│ ←──────────────→ │  Godot Web  │
│  (client.sh)│                │   (:8081)   │                  │  (浏览器)    │
└─────────────┘                └─────────────┘                  └─────────────┘
```

## 安装

### 1. 安装依赖

```bash
# Go 1.22+
/usr/local/go/bin/go version

# Playwright
pip3 install playwright
python3 -m playwright install chromium
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
./bin/debug-server &          # 1. 启动 debug server
make run                      # 2. 启动主服务器
./bin/launch-browser.py       # 3. 启动浏览器
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

### 浏览器端

| 脚本 | 用途 |
|------|------|
| `bin/launch-browser.py` | 启动 Playwright 浏览器，自动连接到 debug server |

### 工作流脚本

| 脚本 | 用途 |
|------|------|
| `bin/debug-start.sh` | 一键启动完整环境（server + browser） |
| `bin/debug-stop.sh` | 停止所有 debug 相关进程 |

## 使用示例

### 1. 截图

```bash
# 先启动环境
./bin/debug-start.sh

# 等待 Godot 连接完成（约 15 秒）

# 发送截图命令
./bin/client.sh capture

# 或使用 Python 客户端
./server/client.py capture
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

```bash
./bin/client.sh camera '{"target":[0,0,0],"distance":30,"azimuth":0,"polar":45}'
```

## API 端点

Debug Server 提供以下 HTTP API：

| 端点 | 方法 | 描述 |
|------|------|------|
| `/health` | GET | 健康检查，返回 godot 连接状态 |
| `/api/camera` | POST | 设置相机参数 |
| `/api/preset` | POST | 设置预设视角 |
| `/api/capture` | POST | 截图 |
| `/api/info` | GET | 获取相机信息 |
| `/godot` | WebSocket | Godot 浏览器客户端连接 |

## 故障排查

### 无法捕获 Godot 日志

**现象**：`launch-browser.py` 看不到 `[DebugController]` 日志

**原因**：`time.sleep()` 阻塞了 Playwright 事件循环

**解决**：使用 `page.wait_for_timeout()` 代替 `time.sleep()`

```python
# 错误
time.sleep(15)

# 正确
page.wait_for_timeout(15000)
```

### Godot 未连接

**检查步骤**：

```bash
# 1. 检查 debug-server 是否运行
curl http://localhost:8081/health

# 2. 检查浏览器是否运行
ps aux | grep launch-browser

# 3. 检查 Godot 控制台日志
# 查看 launch-browser.py 的输出是否有 "Initializing..."
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
│   ├── debug-start.sh    # 启动完整环境
│   ├── debug-stop.sh     # 停止环境
│   ├── launch-browser.py # 启动 Playwright 浏览器
│   ├── client.sh         # HTTP 客户端 (Bash)
│   └── debug-server      # 编译后的服务器 (二进制)
├── server/
│   ├── main.go           # Debug server 源码
│   └── go.mod            # Go 依赖
└── scripts/              # 其他调试脚本
    ├── capture.py
    └── ...
```

## 注意事项

1. **事件循环**：使用 Playwright 时，避免使用 `time.sleep()`，改用 `page.wait_for_timeout()`
2. **进程管理**：`debug-stop.sh` 只停止由 `debug-start.sh` 创建的进程
3. **Godot 初始化**：Godot Web 需要约 10-15 秒完成初始化

## 相关文档

- `../../AGENTS.md` - 项目整体架构
- `../../../docs/VISUAL_DESIGN.md` - 视觉设计规范
