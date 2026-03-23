# Agent-Town AI 助手指南

本文档为 AI 编程助手提供项目背景、架构和开发规范。

## 项目概述

**Agent-Town** 是一个 3D 虚拟小镇沙盒模拟游戏：
- **AI Agent** 通过 CLI 与世界交互（可自动化或人工控制）
- **人类观察者** 通过 Web 浏览器观察世界并进行轻量级干预
- **服务端** 管理世界状态、物理、经济和 Agent 交互

## 技术栈

| 组件 | 技术 | 版本 |
|------|------|------|
| 后端 | Go | 1.22+ |
| Web 前端 | Godot | 4.5 |
| 3D 渲染 | Godot Engine | 内置 |
| Web 导出 | WebAssembly | WASM |
| API 协议 | HTTP REST | JSON |
| 数据库 | PostgreSQL | 16（计划中） |
| 缓存 | Redis | 7（计划中） |

## 环境安装

首次使用需要安装依赖（Go 1.22+、Godot 4.5、导出模板）：

```bash
./scripts/setup.sh
```

或手动安装：
```bash
./scripts/install-go.sh              # 安装 Go
./scripts/install-godot.sh           # 安装 Godot
./scripts/install-godot-templates.sh # 安装导出模板（~1GB）
```

## 构建命令

**⚠️ 始终优先使用 Make 命令** - 不要手动执行底层命令，Make 会处理依赖、错误检查和日志管理。

```bash
# 查看所有可用命令
make help

# 构建所有组件（server + cli + web）
make build

# 单独构建服务端
make server

# 单独构建 CLI
make cli

# 单独导出 Web 前端（需要 godot4）
make web

# 运行测试
make test

# 清理构建产物
make clean
```

### 运行与调试

```bash
# 构建并启动主服务器（端口 8080）
make run

# 停止服务器
make stop

# 启动 Godot 原生模式（用于开发调试，不导出 Web）
make godot-run

# 构建并运行 Debug Server（端口 8081）
make debug-server-run

# 一键启动完整调试环境（server + debug-server + godot）
make debug

# 停止调试环境
make debug-stop
```

### 快速开始

```bash
# 完整开发环境
make debug

# 然后打开浏览器访问 http://localhost:8080
# 或使用客户端脚本：
#   ./.agents/skills/godot-web-debug/bin/client.sh capture  # 截图
#   ./.agents/skills/godot-web-debug/bin/client.sh info     # 相机信息
```

### 为什么使用 Make？

| 优势 | 说明 |
|------|------|
| **统一入口** | 所有命令在一个地方管理，查看 `make help` 即可 |
| **依赖处理** | `make run` 会自动先执行 `make server` |
| **错误检查** | `make run` 会 curl 检查服务器是否成功启动 |
| **日志管理** | 自动重定向到 `/tmp/` 文件 |
| **环境隔离** | 自动处理工作目录和环境变量 |

### 开发工作流

```bash
# 1. 开发时直接运行（不构建）
cd server/cmd/server && go run .

# 3. 修改 Godot 后重新导出
make web
```

## 项目结构

```
agent-town/
├── cli/              # Agent CLI 客户端（Go）
├── server/           # 后端服务器（Go）
│   └── cmd/server/
│       └── web/      # Godot Web 导出（生成，gitignored）
├── godot-web/        # Web 前端（Godot 项目）
├── proto/            # Protocol Buffer 定义
├── scripts/          # 安装脚本
├── docs/             # 文档
├── Makefile
└── go.mod
```

### Godot Web 开发

修改 Godot 场景或脚本后：

```bash
# 导出 Web 版本
cd godot-web && godot4 --headless --export-release "Web" ../server/cmd/server/web/index.html

# 或使用 make
make web
```

Godot 项目结构：
- `scenes/` - 场景文件（.tscn）
- `scripts/` - GDScript 文件（.gd）
- `assets/` - 3D 模型、纹理、材质

## 架构详情

### 当前实现状态

**服务端** (`server/cmd/server/main.go`)：
- 基于 HTTP 的 REST API 服务器
- 提供 `/api/v1/` 端点
- 从 `./web` 目录提供静态 Web 文件（Godot 导出）
- 启用 CORS，允许所有来源
- 为 Godot Web 设置正确的 MIME 类型和 COOP/COEP 头

**世界生成** (`server/internal/world/`)：
- 使用 Simplex 噪声生成分形地形
- 支持多种地形类型：草地、道路、水域、农田、沙滩、山丘、地基
- 区块系统：32x32 地块，惰性生成，内存缓存（LRU 淘汰）
- 初始生成 3x3 区块（中心区块和周围 8 个区块）
- 四个初始建筑：市政厅、银行、任务中心、商店
- 道路系统连接四个建筑
- 资源生成：树木（草地/农田）、矿产（山丘）

**CLI** (`cli/cmd/cli/main.go`)：
- 占位实现
- 计划功能：注册、登录、状态查询、移动、采集、背包、TODO、Token 生成

**Godot Web** (`godot-web/`)：
- 功能完整的 3D 世界查看器
- 使用 Godot Engine 4.5 WebAssembly 导出
- 使用 MultiMesh 渲染地形以获得性能
- 每 5 秒自动刷新世界数据
- HUD 作为独立 CanvasLayer
- 名称标签使用屏幕空间（不受 3D 遮挡影响）

### API 端点（当前）

```
GET  /api/v1/world/info              # 世界元数据
GET  /api/v1/world/time              # 世界时间
GET  /api/v1/world/map?x=0&y=0&radius=20  # 地图数据
GET  /api/v1/agents                  # Agent 列表
GET  /api/v1/agents/{id}             # Agent 详情
GET  /api/v1/agents/{id}/status      # Agent 状态
GET  /api/v1/agents/{id}/visible-area
GET  /api/v1/agents/{id}/todos
GET  /api/v1/agents/{id}/skills
```

### 协议缓冲区定义（规划中）

`proto/agenttown/v1/` 目录包含 gRPC 服务定义：
- `AgentService` - Agent 管理、认证、状态
- `ActionService` - 采集、建造、种植、对话、交易、战斗
- `BattleService` - PvP/PvE 战斗
- `TaskService` - 任务系统
- `WorldService` - 世界信息、区块、时间
- `EconomyService` - 转账、交易、好友

## AI 编程助手工作原则

### 先理解视觉效果，再反推实现

**教训**：水面渲染优化中，用户想要的是"分层渲染"（水面在下，陆地在上遮挡），但助手陷入"精确渲染"思维，试图找出哪些格子是水、找边界、合并相邻水面，导致过度复杂化。

**原则**：
- 首先问清用户想要的**最终视觉效果**
- 不要默认必须用精确几何匹配
- 考虑"超渲染+遮挡"等简化方案

### 利用服务端已有结构

**教训**：客户端从格子反推 chunk，增加复杂性。应该修改服务端直接返回 chunk 结构。

**原则**：
- 修改服务端通常比客户端 workaround 更干净
- API 应该直接返回客户端需要的结构
- 减少客户端的数据转换逻辑

### 理清坐标系再动手

**教训**：处理 agent 渲染时，没有仔细梳理 API 坐标（X, Y, Z）和 Godot 坐标（X, Y高度, Z）的映射，导致 agent 被埋进地下。

**原则**：
- 画图或表格明确坐标映射关系
- 确认方块的体积和边界
- 验证高度计算（中心 vs 边界）

### 避免过早优化细节

**教训**：一开始就想优化单个方块的渲染效率，而没有先搭建正确的大框架。

**原则**：
- 先让基本逻辑正确，再优化性能
- 不要假设必须优化，先测量
- 简单方案往往足够好

### 及时创建工具脚本

**教训**：多次手动 `export PATH`，浪费时间。

**原则**：
- 重复使用两次以上的命令应做成脚本
- 放在 `scripts/` 目录便于管理
- 虚拟环境放在 `~/.config/agent-town/venv` 或临时目录

## 代码风格指南

### Go
- 遵循标准 Go 格式化（`gofmt`）
- 使用 `go vet` 进行静态分析
- 优先显式错误处理
- 包结构：`cmd/` 用于可执行文件，`internal/` 用于私有代码
- 世界包使用中文注释，保持与文档一致

### GDScript（Godot）
- 函数和变量使用 `snake_case`
- 类名和常量使用 `PascalCase`
- 使用制表符缩进（Godot 约定）
- 鼓励函数参数和返回值使用类型提示
- 信号名使用 `snake_case` 和过去时态动词（`player_died`）

### Protocol Buffers
- 使用 proto3 语法
- 包名：`agenttown.v1`
- Go 包：`github.com/RedContritio/agent-town/proto/agenttown/v1`
- 消息字段使用 snake_case

## 测试策略

```bash
# 运行所有 Go 测试
make test

# 运行特定包
go test -v ./server/...
```

**注意**：项目目前测试覆盖较少。应添加：
- 核心游戏逻辑测试（世界生成、Agent 动作）
- API 处理器测试
- 协议缓冲区消息验证测试

## 关键配置

### 环境变量（服务端）

```bash
DATABASE_URL=postgres://agenttown:agenttown@localhost:5432/agenttown?sslmode=disable
REDIS_URL=localhost:6379
GRPC_PORT=50051
HTTP_PORT=8080
```

### Godot 导出

导出设置 (`godot-web/export_presets.cfg`)：
- 导出路径：`../server/cmd/server/web/index.html`
- 自定义 HTML 外壳：默认
- Head 包含：无（由 Go 服务器设置头）

## GitIgnore 结构

每个目录有自己的 `.gitignore`：

- `.gitignore`（根目录）- IDE、OS、环境变量、日志
- `godot-web/.gitignore` - `.godot/`、`.import/`、Godot 缓存
- `server/cmd/server/.gitignore` - Web 导出产物（`web/`、`*.wasm`、`*.pck`）

## 视觉设计

参见 `docs/VISUAL_DESIGN.md` 获取完整视觉规范，包括：
- 调色板（UI 和世界）
- 字体排版
- 建筑类型颜色
- UI 组件样式

关键颜色：
- UI 背景：`#1a1a2e`
- UI 强调：`#4fc3f7`
- 天空：`#1a1a2e`（Godot 环境背景色）
- 地面：`#c0c0c0`

地形颜色（来自 VISUAL_DESIGN.md）：
- 草地：`#5a8f63`
- 道路：`#c8b08a`
- 水域：`#4a92b8`
- 农田：`#9a7820`
- 沙滩：`#d4c088`
- 地基：`#6a6a6a`

建筑颜色：
- shop（商店）：`#e8a87c`
- bank（银行）：`#4a90d9`
- exchange（交易所）：`#50c878`
- park（公园）：`#6b8e23`
- office（办公室）：`#8899aa`
- cafe（咖啡馆）：`#c4956a`
- home（住宅）：`#d4a574`

## 重要设计决策

1. **TODO Token 系统**：Agent 通过 CLI 生成 Token 授权 Web 用户管理其 TODO 列表。Token 通过 `X-Todo-Token` 头传递。

2. **视野系统**：服务端只提供 Agent 视野范围内的数据。Agent 必须自己维护已探索区域的"心理地图"。

3. **身份认证**：Agent 使用公钥密码学进行认证。Agent ID 派生自公钥。

4. **经济系统**：所有交易对所有 Agent 可见（透明经济）。税率可通过公投调整。

5. **战斗系统**：回合制。PvP 需要双方同意。PvE 允许离线僵持。

6. **架构原则**：**服务端是唯一的真实数据源**。前端只负责：
   - 渲染服务端提供的状态
   - 将用户输入传递给服务端
   - 本地视图状态管理（与业务逻辑无关）

## 文档参考

详细设计文档在 `docs/` 目录：

- `GAME_DESIGN.md` - 游戏机制、世界系统、Agent 生命周期（中文）
- `TECH_DESIGN.md` - 数据库 Schema、API 设计、服务架构（中文）
- `ARCHITECTURE.md` - 架构分层设计、职责边界（中文）
- `VISUAL_DESIGN.md` - 当前视觉规范
- `WEB_DESIGN.md` - 旧版 Web 设计（React 时代），参见 VISUAL_DESIGN.md

## 常见任务

### 添加新 API 端点

1. 在 `server/cmd/server/main.go` 添加 HTTP 处理器
2. 在 `godot-web/scripts/api_client.gd` 添加客户端方法
3. 在适当的场景/脚本中使用新 API

### 添加新地形类型

1. 在服务端更新地形生成逻辑（`server/internal/world/terrain.go`）
2. 在 `godot-web/scripts/world_manager.gd` 添加材质定义
3. 在 `docs/VISUAL_DESIGN.md` 更新颜色

### 添加新建筑类型

1. 在 `godot-web/scripts/world_manager.gd` 更新 `terrain_materials`
2. 在 `godot-web/scripts/entities/building.gd` 更新 `TYPE_COLORS`
3. 在 HUD 图例 `godot-web/scripts/ui/hud.gd` 添加
4. 在 `docs/VISUAL_DESIGN.md` 更新颜色规范

### 数据库 Schema 变更

1. 在 `docs/TECH_DESIGN.md` 更新 Schema
2. 创建迁移文件（迁移系统实现后）
3. 更新 Go 结构体
4. 如 API 变更，更新 proto 消息

## 安全考虑

- **认证**：公钥签名验证后的 JWT Token
- **TODO Token**：用于 Web TODO 管理的独立限定范围 Token
- **速率限制**：应在 API 端点上实现（目前 TODO）
- **输入验证**：验证所有传入的 JSON 数据
- **CORS**：目前允许所有来源（`*`）- 生产环境应限制
- **COOP/COEP**：为 Godot WebAssembly（SharedArrayBuffer）设置必需的头

## 许可

MIT License - Copyright (c) 2026 RedContritio
