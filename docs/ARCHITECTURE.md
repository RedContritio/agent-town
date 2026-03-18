# Agent Town 架构分层设计

## 总体原则

**服务端是唯一的真实数据源（Single Source of Truth）**，前端仅负责：
1. 渲染服务端提供的状态
2. 将用户输入传递给服务端
3. 本地视图状态管理（与业务逻辑无关）

```
┌─────────────────────────────────────────────────────────────┐
│                      Client (Godot)                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │  Rendering   │  │     UI       │  │  Input Handling  │  │
│  │  (Pure View) │  │  (Display)   │  │  (Send to API)   │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
│  ┌──────────────┐  ┌──────────────┐                         │
│  │ API Client   │  │ Local Cache  │  (View State Only)      │
│  │ (HTTP/WS)    │  │ (Height Map) │                         │
│  └──────────────┘  └──────────────┘                         │
└──────────────────────────┬──────────────────────────────────┘
                           │ HTTP / WebSocket
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                     Server (Go)                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │  HTTP API    │  │  WebSocket   │  │   Auth/JWT       │  │
│  │  (REST)      │  │  (Real-time) │  │   (Identity)     │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Business Logic Layer                     │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌─────────┐ │  │
│  │  │  Agent   │ │ Economy  │ │ Building │ │  Battle │ │  │
│  │  │  System  │ │  System  │ │  System  │ │ System  │ │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └─────────┘ │  │
│  └──────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              World State (Single Source)              │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌─────────┐ │  │
│  │  │  Chunk   │ │  Agent   │ │ Building │ │  Item   │ │  │
│  │  │  Manager │ │  Store   │ │  Store   │ │  Store  │ │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └─────────┘ │  │
│  └──────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Persistence Layer                      │  │
│  │         (PostgreSQL / Redis / File)                 │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

---

## 服务端职责 (Server)

### 1. 世界状态管理
```go
// server/internal/world/state.go
// 唯一存储所有世界状态
- World: 世界元数据
- ChunkManager: 区块数据（地形、资源）
- AgentStore: Agent 状态（位置、HP、背包等）
- BuildingStore: 建筑数据
- Economy: 交易记录、物价
```

### 2. 业务逻辑
```go
// server/internal/logic/
- agent_logic.go:    Agent AI 决策、移动、行为
- economy_logic.go:  交易撮合、价格波动、税收
- building_logic.go: 建筑建造、升级、销毁
- battle_logic.go:   战斗结算、伤害计算
- craft_logic.go:     crafting 配方、生产
```

### 3. 生成和计算
```go
// server/internal/generator/
- terrain.go:   地形生成（噪声算法）
- resources.go: 资源分布计算
- buildings.go: 建筑生成
```

### 4. API 层
```go
// server/cmd/server/
- REST API: 查询状态、发起动作
- WebSocket: 推送实时更新（可选）
- Auth:     身份验证（公钥签名）
```

### 服务端不处理
- ❌ 渲染相关逻辑
- ❌ 相机控制
- ❌ 输入设备处理
- ❌ 视觉特效

---

## 前端职责 (Godot Web)

### 1. 渲染层
```gdscript
# godot-web/scripts/renderers/
- terrain_renderer.gd:   地形网格渲染
- entity_renderer.gd:    Agent/建筑渲染
- effect_renderer.gd:    特效（粒子、光效）
```

### 2. 视图状态（本地缓存）
```gdscript
# 允许前端缓存的纯视图数据：
- terrain_height_map:  用于高度查询（避免重复计算）
- camera_position:     相机位置/旋转
- visible_entities:    当前可见的实体列表
- loaded_chunks:       已加载的区块缓存

# 禁止前端维护：
- ❌ Agent HP/状态（必须从服务端获取）
- ❌ 物品数量（必须从服务端获取）
- ❌ 建筑所有权（必须从服务端获取）
```

### 3. 输入处理
```gdscript
# godot-web/scripts/input/
- camera_controller.gd: 相机移动/缩放
- selection_handler.gd: 实体选择
- command_sender.gd:    发送命令到服务端
```

### 4. UI 展示
```gdscript
# godot-web/scripts/ui/
- hud.gd:              主界面 HUD
- agent_panel.gd:      Agent 详情面板
- building_panel.gd:   建筑信息面板
- chat_window.gd:      聊天窗口
```

### 前端不处理
- ❌ 业务逻辑计算
- ❌ 状态修改（除本地视图状态）
- ❌ 预测性更新（除非服务端支持）

---

## API 契约

### 当前 REST API

```
# 查询类（只读）
GET  /api/v1/world/info        -> 世界元数据
GET  /api/v1/world/time        -> 世界时间
GET  /api/v1/world/map         -> 区块数据（地形、建筑）
GET  /api/v1/agents            -> Agent 列表
GET  /api/v1/agents/{id}       -> Agent 详情
GET  /api/v1/agents/{id}/status
GET  /api/v1/agents/{id}/visible-area
GET  /api/v1/agents/{id}/todos
GET  /api/v1/agents/{id}/skills

# 动作类（修改状态）
POST /api/v1/agents/{id}/move          # 移动
POST /api/v1/agents/{id}/gather        # 采集
POST /api/v1/agents/{id}/build         # 建造
POST /api/v1/agents/{id}/trade         # 交易
POST /api/v1/agents/{id}/talk          # 对话
POST /api/v1/agents/{id}/battle        # 战斗
```

### 未来 WebSocket 协议（实时推送）

```json
// Server -> Client
{
  "type": "entity_update",
  "data": {
    "agent_id": "agent-001",
    "position": {"x": 10, "y": 0, "z": 20},
    "hp": 85
  }
}

// Client -> Server
{
  "type": "command",
  "action": "move",
  "target": {"x": 15, "y": 0, "z": 25}
}
```

---

## 严格边界规则

### 1. 状态修改权
- **只有服务端能修改业务状态**
- 前端发送「请求」而非「命令」
- 服务端验证后执行，返回结果

```gdscript
# ❌ 错误：前端直接修改
agent.position = new_position

# ✅ 正确：发送请求到服务端
api_client.send_move_request(agent_id, target_position)
# 等待服务端确认后更新视图
```

### 2. 数据流向
```
World State (Server) 
    |
    | HTTP Response / WebSocket Push
    v
Local View Cache (Client)
    |
    | Render
    v
Screen
```

### 3. 计算职责
| 计算类型 | 服务端 | 前端 | 说明 |
|----------|--------|------|------|
| 地形生成 | ✅ | ❌ | 噪声算法在服务端 |
| 高度查询 | ❌ | ✅ | 前端缓存用于渲染 |
| Agent AI | ✅ | ❌ | 决策逻辑在服务端 |
| 移动插值 | ❌ | ✅ | 视觉平滑在前端 |
| 经济计算 | ✅ | ❌ | 价格在服务端 |
| 碰撞检测 | ⚠️ | ✅ | 仅视觉碰撞在前端 |

### 4. 文件组织

```
server/
├── internal/
│   ├── world/        # 世界状态（唯一真实数据）
│   ├── logic/        # 业务逻辑
│   ├── generator/    # 生成器
│   └── api/          # API 处理（未来拆分）
└── cmd/server/       # HTTP 服务入口

godot-web/
├── scripts/
│   ├── api/          # API 客户端
│   ├── renderers/    # 渲染器
│   ├── ui/           # UI 组件
│   └── input/        # 输入处理
└── scenes/           # 场景文件
```

---

## 当前需要重构的地方

### 1. 前端优化
- `world_manager.gd` 中的 `terrain_height_map` 是合理的视图缓存 ✅
- 需要移除任何「预测性」更新逻辑

### 2. API 扩展
- 当前只有查询 API，需要添加动作 API
- 添加认证中间件

### 3. 服务端增强
- 分离 `logic` 层从 `world` 层
- 添加 WebSocket 支持用于实时推送

---

## 扩展计划

### Phase 1: 当前（只读查看器）
- 前端：渲染 + 查询 API
- 服务端：世界生成 + 静态数据

### Phase 2: 观察者干预
- 前端：发送干预命令
- 服务端：接收命令 → 修改世界 → 推送更新

### Phase 3: Agent CLI
- CLI 作为 Agent 客户端
- 与 Web 前端共享同一个服务端

### Phase 4: 实时同步
- WebSocket 替换轮询
- 服务端主动推送状态变更
