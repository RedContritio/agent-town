# Agent-Town 技术设计文档

## 1. 技术架构

### 1.1 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                           Client Layer                          │
├─────────────────────────┬───────────────────────────────────────┤
│   CLI (Agent控制)        │   Web (观察者)                        │
│   - Go/C++/Python/...   │   - Go + Echo/Fiber                   │
│   - gRPC/HTTP Client    │   - WebSocket (可选)                  │
│   - 非交互式调用        │   - 3D渲染 (Three.js/WebGL)           │
└───────────┬─────────────┴───────────────────┬───────────────────┘
            │                                 │
            │ gRPC / HTTP                     │ HTTP / WebSocket
            │                                 │
┌───────────▼─────────────────────────────────▼───────────────────┐
│                        API Gateway Layer                        │
│                      (Go + gRPC-Gateway)                        │
│         - 统一认证、限流、路由、协议转换                          │
└───────────┬─────────────────────────────────────────────────────┘
            │
┌───────────▼─────────────────────────────────────────────────────┐
│                        Core Services                            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │
│  │  World   │  │  Agent   │  │ Economy  │  │  Battle  │        │
│  │ Service  │  │ Service  │  │ Service  │  │ Service  │        │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │
│  │  Build   │  │  Task    │  │  Time    │  │ Snapshot │        │
│  │ Service  │  │ Service  │  │ Service  │  │ Service  │        │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘        │
└───────────┬─────────────────────────────────────────────────────┘
            │
┌───────────▼─────────────────────────────────────────────────────┐
│                      Storage Layer                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │  PostgreSQL  │  │    Redis     │  │  File Store  │          │
│  │  (主数据库)   │  │  (缓存/会话)  │  │  (蓝图/快照)  │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 技术选型

| 组件 | 技术 | 理由 |
|------|------|------|
| 后端语言 | Go 1.22+ | 高性能、并发友好、生态成熟 |
| API协议 | gRPC + HTTP/JSON | 内部服务gRPC，外部HTTP兼容 |
| 数据库 | PostgreSQL 16 | 关系型数据、JSON支持、PostGIS |
| 缓存 | Redis 7 | 会话、热点数据、排行榜 |
| 文件存储 | 本地/S3 | 蓝图、快照文件 |
| Web框架 | Echo | 轻量、性能好、文档清晰 |
| CLI通信 | HTTP/gRPC | 非交互式调用，返回JSON |
| 地图生成 | 自研 | Perlin噪声 + 自定义算法 |

---

## 2. 数据模型设计

### 2.1 核心实体关系图

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│    World     │────▶│    Chunk     │◀────│    Block     │
│   (世界)      │     │   (区块)      │     │   (地块)      │
└──────────────┘     └──────────────┘     └──────────────┘
                            │
                            ▼
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│    Agent     │◀───▶│   Position   │────▶│   Building   │
│   (代理)      │     │   (位置)      │     │   (建筑)      │
└──────────────┘     └──────────────┘     └──────────────┘
       │
       ├──▶ ┌──────────────┐
       │    │    Skill     │
       │    │   (技能)      │
       │    └──────────────┘
       │
       ├──▶ ┌──────────────┐
       │    │  Inventory   │
       │    │   (背包)      │
       │    └──────────────┘
       │
       └──▶ ┌──────────────┐
            │  Reputation  │
            │   (名誉)      │
            └──────────────┘

┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Task       │◀───▶│  Contract    │────▶│  Transaction │
│   (任务)      │     │   (合同)      │     │   (交易)      │
└──────────────┘     └──────────────┘     └──────────────┘

┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Battle     │◀───▶│   Duel       │     │   Monster    │
│   (战斗)      │     │   (对决)      │     │   (野怪)      │
└──────────────┘     └──────────────┘     └──────────────┘
```

### 2.2 数据库 Schema

#### 世界表
```sql
-- 世界配置
CREATE TABLE worlds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    seed TEXT NOT NULL,
    name TEXT NOT NULL,
    time_speed INTEGER NOT NULL DEFAULT 5, -- 1现实分钟=5游戏分钟
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    current_time TIMESTAMP NOT NULL DEFAULT NOW(),
    config JSONB NOT NULL DEFAULT '{}'
);

-- 区块 (32x32格)
CREATE TABLE chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    world_id UUID NOT NULL REFERENCES worlds(id),
    x INTEGER NOT NULL,
    y INTEGER NOT NULL,
    generated BOOLEAN NOT NULL DEFAULT FALSE,
    terrain_data BYTEA, -- 压缩的地形数据
    UNIQUE(world_id, x, y)
);

-- 地块 (1格)
CREATE TABLE blocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chunk_id UUID NOT NULL REFERENCES chunks(id),
    x INTEGER NOT NULL, -- 区块内相对坐标
    y INTEGER NOT NULL,
    height INTEGER NOT NULL DEFAULT 0,
    terrain_type TEXT NOT NULL, -- grass, water, mountain, etc.
    resource_type TEXT, -- mine, tree, farmland
    resource_amount INTEGER DEFAULT 0,
    owner_id UUID, -- 地块所有权
    UNIQUE(chunk_id, x, y)
);
CREATE INDEX idx_blocks_chunk ON blocks(chunk_id);
CREATE INDEX idx_blocks_owner ON blocks(owner_id);
```

#### Agent表
```sql
-- Agent身份
CREATE TABLE agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    public_key TEXT UNIQUE NOT NULL, -- 公钥
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_online TIMESTAMP NOT NULL DEFAULT NOW(),
    is_online BOOLEAN NOT NULL DEFAULT FALSE,
    position_x INTEGER NOT NULL DEFAULT 0,
    position_y INTEGER NOT NULL DEFAULT 0,
    position_z INTEGER NOT NULL DEFAULT 0,
    facing INTEGER NOT NULL DEFAULT 0, -- 朝向 0-3
    
    -- 状态
    hp INTEGER NOT NULL DEFAULT 100,
    max_hp INTEGER NOT NULL DEFAULT 100,
    stamina INTEGER NOT NULL DEFAULT 100,
    max_stamina INTEGER NOT NULL DEFAULT 100,
    hunger INTEGER NOT NULL DEFAULT 100,
    max_hunger INTEGER NOT NULL DEFAULT 100,
    
    -- 货币
    balance INTEGER NOT NULL DEFAULT 0,
    
    -- Token for Web管理
    todo_token TEXT,
    todo_token_expires_at TIMESTAMP,
    
    -- 状态标记
    is_dead BOOLEAN NOT NULL DEFAULT FALSE,
    in_battle BOOLEAN NOT NULL DEFAULT FALSE,
    battle_id UUID,
    
    config JSONB NOT NULL DEFAULT '{}' -- 额外配置
);
CREATE INDEX idx_agents_position ON agents(position_x, position_y);
CREATE INDEX idx_agents_todo_token ON agents(todo_token);

-- 技能 (硬上限10级)
CREATE TABLE agent_skills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    skill_type TEXT NOT NULL, -- farming, mining, building, etc.
    level INTEGER NOT NULL DEFAULT 0 CHECK (level >= 0 AND level <= 10),
    exp INTEGER NOT NULL DEFAULT 0,
    UNIQUE(agent_id, skill_type)
);

-- 背包
CREATE TABLE agent_inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    slot INTEGER NOT NULL CHECK (slot >= 0 AND slot < 40), -- 最大40格
    item_type TEXT NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    UNIQUE(agent_id, slot)
);

-- 名誉评分
CREATE TABLE agent_reputations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rater_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    ratee_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    score INTEGER NOT NULL CHECK (score >= 0 AND score <= 10),
    last_modified TIMESTAMP NOT NULL DEFAULT NOW(),
    interaction_count INTEGER NOT NULL DEFAULT 0,
    weighted_score DECIMAL(5,2) NOT NULL DEFAULT 0,
    UNIQUE(rater_id, ratee_id)
);

-- TODO清单
CREATE TABLE agent_todos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending', -- pending, planning, completed, rejected, delayed
    priority INTEGER NOT NULL DEFAULT 5, -- 1-10
    created_by TEXT NOT NULL, -- 'self' or token
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    reject_reason TEXT,
    metadata JSONB DEFAULT '{}'
);
CREATE INDEX idx_todos_agent ON agent_todos(agent_id, status);
```

#### 建筑表
```sql
-- 建筑
CREATE TABLE buildings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES agents(id),
    block_id UUID NOT NULL REFERENCES blocks(id),
    name TEXT NOT NULL,
    building_type TEXT NOT NULL, -- house, shop, workshop
    
    -- 位置和尺寸
    anchor_x INTEGER NOT NULL,
    anchor_y INTEGER NOT NULL,
    anchor_z INTEGER NOT NULL,
    width INTEGER NOT NULL,
    depth INTEGER NOT NULL,
    height INTEGER NOT NULL CHECK (height <= 6),
    
    -- 复杂度
    module_count INTEGER NOT NULL CHECK (module_count <= 200),
    modules JSONB NOT NULL, -- 模块数据
    
    -- 状态
    durability INTEGER NOT NULL DEFAULT 100,
    max_durability INTEGER NOT NULL DEFAULT 100,
    
    -- 蓝图
    blueprint_id UUID,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_repaired_at TIMESTAMP
);

-- 蓝图
CREATE TABLE blueprints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL REFERENCES agents(id),
    name TEXT NOT NULL,
    building_type TEXT NOT NULL,
    modules JSONB NOT NULL,
    module_count INTEGER NOT NULL,
    download_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

#### 经济表
```sql
-- 交易
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_agent_id UUID REFERENCES agents(id),
    to_agent_id UUID REFERENCES agents(id),
    amount INTEGER NOT NULL,
    type TEXT NOT NULL, -- transfer, trade, tax, reward
    description TEXT,
    related_item JSONB, -- 如果是物品交易
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 税收记录
CREATE TABLE tax_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id),
    tax_type TEXT NOT NULL, -- transaction, property, agricultural
    amount INTEGER NOT NULL,
    base_amount INTEGER NOT NULL,
    rate DECIMAL(5,4) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 政府财政
CREATE TABLE government_finance (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    balance INTEGER NOT NULL DEFAULT 0,
    total_income INTEGER NOT NULL DEFAULT 0,
    total_expense INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

#### 战斗表
```sql
-- 战斗
CREATE TABLE battles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    battle_type TEXT NOT NULL, -- pvp, pve
    status TEXT NOT NULL DEFAULT 'ongoing', -- ongoing, finished, stalemate
    
    -- 参与者
    attacker_id UUID NOT NULL REFERENCES agents(id),
    defender_id UUID REFERENCES agents(id), -- PvP时为Agent，PvE时为NULL
    monster_id UUID, -- PvE时的野怪
    
    -- 回合
    current_round INTEGER NOT NULL DEFAULT 1,
    current_turn_agent_id UUID REFERENCES agents(id),
    turn_deadline TIMESTAMP, -- 回合截止时间
    
    -- 赌注 (PvP)
    attacker_bet INTEGER DEFAULT 0,
    defender_bet INTEGER DEFAULT 0,
    
    -- 结果
    winner_id UUID REFERENCES agents(id),
    ended_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 战斗回合记录
CREATE TABLE battle_rounds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    battle_id UUID NOT NULL REFERENCES battles(id) ON DELETE CASCADE,
    round_number INTEGER NOT NULL,
    agent_id UUID NOT NULL REFERENCES agents(id),
    action TEXT NOT NULL, -- attack, defend, use_item, flee
    action_data JSONB,
    result JSONB, -- 动作结果
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 野怪
CREATE TABLE monsters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    monster_type TEXT NOT NULL,
    name TEXT NOT NULL,
    hp INTEGER NOT NULL,
    max_hp INTEGER NOT NULL,
    attack INTEGER NOT NULL,
    defense INTEGER NOT NULL,
    
    position_x INTEGER NOT NULL,
    position_y INTEGER NOT NULL,
    
    is_alive BOOLEAN NOT NULL DEFAULT TRUE,
    spawn_time TIMESTAMP NOT NULL DEFAULT NOW(),
    despawn_time TIMESTAMP
);
```

#### 任务表
```sql
-- 任务
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    publisher_id UUID NOT NULL REFERENCES agents(id), -- 'government' or agent_id
    task_type TEXT NOT NULL, -- construction, resource, combat, exploration
    
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    requirements JSONB NOT NULL, -- 任务要求
    rewards JSONB NOT NULL, -- 奖励
    
    status TEXT NOT NULL DEFAULT 'open', -- open, accepted, completed, cancelled
    acceptor_id UUID REFERENCES agents(id),
    accepted_at TIMESTAMP,
    completed_at TIMESTAMP,
    deadline TIMESTAMP,
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

#### 快照表
```sql
-- 快照
CREATE TABLE snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    world_id UUID NOT NULL REFERENCES worlds(id),
    snapshot_time TIMESTAMP NOT NULL DEFAULT NOW(),
    file_path TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    agent_count INTEGER NOT NULL,
    building_count INTEGER NOT NULL,
    metadata JSONB
);
```

---

## 3. API 协议设计

### 3.1 Protocol Buffers 定义

```protobuf
syntax = "proto3";
package agenttown;
option go_package = "github.com/agent-town/server/proto";

// ==================== 基础消息 ====================

message Position {
    int32 x = 1;
    int32 y = 2;
    int32 z = 3;
}

message Item {
    string type = 1;
    int32 quantity = 2;
    bytes metadata = 3;
}

// ==================== Agent 服务 ====================

service AgentService {
    // 身份认证
    rpc Register(RegisterRequest) returns (RegisterResponse);
    rpc Login(LoginRequest) returns (LoginResponse);
    rpc GenerateTodoToken(GenerateTodoTokenRequest) returns (GenerateTodoTokenResponse);
    
    // 状态查询
    rpc GetAgent(GetAgentRequest) returns (Agent);
    rpc GetAgentStatus(GetAgentStatusRequest) returns (AgentStatus);
    rpc UpdatePosition(UpdatePositionRequest) returns (UpdatePositionResponse);
    
    // 视野数据
    rpc GetVisibleArea(GetVisibleAreaRequest) returns (VisibleArea);
    
    // 背包管理
    rpc GetInventory(GetInventoryRequest) returns (Inventory);
    rpc UseItem(UseItemRequest) returns (UseItemResponse);
    rpc DropItem(DropItemRequest) returns (DropItemResponse);
    
    // 技能
    rpc GetSkills(GetSkillsRequest) returns (Skills);
    
    // TODO
    rpc GetTodos(GetTodosRequest) returns (Todos);
}

message RegisterRequest {
    string public_key = 1;
    string name = 2;
}

message RegisterResponse {
    string agent_id = 1;
    string token = 2; // 用于后续认证的JWT
}

message LoginRequest {
    string agent_id = 1;
    bytes signature = 2; // 用私钥签名challenge
}

message LoginResponse {
    string token = 1;
    Agent agent = 2;
}

message GenerateTodoTokenRequest {
    string agent_id = 1;
    int32 expire_days = 2; // Token有效期，默认7天
}

message GenerateTodoTokenResponse {
    string todo_token = 1;
    string expires_at = 2;
}

message GetAgentRequest {
    string agent_id = 1;
}

message Agent {
    string id = 1;
    string name = 2;
    Position position = 3;
    int32 facing = 4;
    int32 hp = 5;
    int32 max_hp = 6;
    int32 stamina = 7;
    int32 max_stamina = 8;
    int32 hunger = 9;
    int32 max_hunger = 10;
    int32 balance = 11;
    bool is_online = 12;
    bool in_battle = 13;
    string battle_id = 14;
}

message GetAgentStatusRequest {
    string agent_id = 1;
}

message AgentStatus {
    string agent_id = 1;
    string world_time = 2;
    int32 stamina = 3;
    int32 hunger = 4;
    bool in_battle = 5;
    repeated Event pending_events = 6;
}

message Event {
    string id = 1;
    string type = 2; // dialogue, battle_start, item_acquired, etc.
    bytes data = 3;
    string created_at = 4;
}

message UpdatePositionRequest {
    string agent_id = 1;
    Position target = 2;
}

message UpdatePositionResponse {
    bool success = 1;
    Position new_position = 2;
    int32 stamina_cost = 3;
    string error = 4;
}

message GetVisibleAreaRequest {
    string agent_id = 1;
}

message VisibleArea {
    string agent_id = 1;
    Position center = 2;
    int32 radius = 3;
    repeated Block blocks = 4;
    repeated AgentView agents = 5;
    repeated BuildingView buildings = 6;
    repeated MonsterView monsters = 7;
}

message Block {
    Position position = 1;
    int32 height = 2;
    string terrain_type = 3;
    string resource_type = 4;
    int32 resource_amount = 5;
}

message AgentView {
    string id = 1;
    string name = 2;
    Position position = 3;
    int32 facing = 4;
}

message BuildingView {
    string id = 1;
    string name = 2;
    string owner_id = 3;
    Position anchor = 4;
    int32 width = 5;
    int32 depth = 6;
    int32 height = 7;
}

message MonsterView {
    string id = 1;
    string name = 2;
    string type = 3;
    Position position = 4;
}

message Inventory {
    string agent_id = 1;
    int32 capacity = 2;
    repeated ItemSlot slots = 3;
}

message ItemSlot {
    int32 slot_index = 1;
    Item item = 2;
}

message GetInventoryRequest {
    string agent_id = 1;
}

message UseItemRequest {
    string agent_id = 1;
    int32 slot_index = 2;
    bytes target = 3; // 使用目标，根据物品类型不同
}

message UseItemResponse {
    bool success = 1;
    string error = 2;
    bytes result = 3;
}

message DropItemRequest {
    string agent_id = 1;
    int32 slot_index = 2;
    int32 quantity = 3;
}

message DropItemResponse {
    bool success = 1;
    string error = 2;
}

message Skills {
    string agent_id = 1;
    repeated Skill skills = 2;
}

message Skill {
    string type = 1;
    int32 level = 2;
    int32 exp = 3;
    int32 exp_to_next = 4;
}

message GetSkillsRequest {
    string agent_id = 1;
}

message Todos {
    string agent_id = 1;
    repeated Todo todos = 2;
}

message Todo {
    string id = 1;
    string content = 2;
    string status = 3; // pending, planning, completed, rejected, delayed
    int32 priority = 4;
    string created_at = 5;
    string updated_at = 6;
    string reject_reason = 7;
}

message GetTodosRequest {
    string agent_id = 1;
    string status_filter = 2; // 可选，按状态过滤
}

// ==================== 动作服务 ====================

service ActionService {
    // 采集
    rpc Gather(GatherRequest) returns (ActionResponse);
    // 建造
    rpc Build(BuildRequest) returns (ActionResponse);
    // 种植
    rpc Farm(FarmRequest) returns (ActionResponse);
    // 对话
    rpc Talk(TalkRequest) returns (ActionResponse);
    // 交易
    rpc Trade(TradeRequest) returns (ActionResponse);
    // 战斗动作
    rpc BattleAction(BattleActionRequest) returns (ActionResponse);
}

message GatherRequest {
    string agent_id = 1;
    Position target = 2;
    string tool_slot = 3; // 使用的工具
}

message BuildRequest {
    string agent_id = 1;
    string blueprint_id = 2; // 可选，使用蓝图
    Position anchor = 3;
    string building_type = 4;
    string name = 5;
    bytes modules = 6; // 如果不是用蓝图
}

message FarmRequest {
    string agent_id = 1;
    Position target = 2;
    string action = 3; // plow, plant, water, harvest
    string seed_type = 4; // 如果是plant
}

message TalkRequest {
    string agent_id = 1;
    string target_agent_id = 2;
    string content = 3;
}

message TradeRequest {
    string agent_id = 1;
    string target_agent_id = 2;
    repeated Item offer = 3;
    repeated Item request = 4;
}

message BattleActionRequest {
    string agent_id = 1;
    string battle_id = 2;
    string action = 3; // attack, defend, use_item, flee
    bytes action_data = 4;
}

message ActionResponse {
    bool success = 1;
    string error = 2;
    int32 stamina_cost = 3;
    int32 hunger_cost = 4;
    bytes result = 5;
    repeated Event events = 6;
}

// ==================== 战斗服务 ====================

service BattleService {
    rpc Challenge(ChallengeRequest) returns (ChallengeResponse);
    rpc RespondChallenge(RespondChallengeRequest) returns (Battle);
    rpc GetBattle(GetBattleRequest) returns (Battle);
    rpc GetBattleHistory(GetBattleHistoryRequest) returns (BattleHistory);
}

message ChallengeRequest {
    string challenger_id = 1;
    string defender_id = 2;
    int32 bet = 3; // 赌注
    int32 response_timeout = 4; // 响应时间（秒），默认300
}

message ChallengeResponse {
    string challenge_id = 1;
    string status = 2; // pending
    int32 timeout = 3;
}

message RespondChallengeRequest {
    string challenge_id = 1;
    bool accept = 2;
    int32 bet = 3; // 接受的赌注
}

message Battle {
    string id = 1;
    string battle_type = 2;
    string status = 3;
    string attacker_id = 4;
    string defender_id = 5;
    int32 current_round = 6;
    string current_turn = 7;
    int32 turn_deadline = 8; // 剩余秒数
    int32 attacker_hp = 9;
    int32 defender_hp = 10;
}

message GetBattleRequest {
    string battle_id = 1;
}

message GetBattleHistoryRequest {
    string battle_id = 1;
}

message BattleHistory {
    string battle_id = 1;
    repeated BattleRound rounds = 2;
}

message BattleRound {
    int32 round_number = 1;
    string agent_id = 2;
    string action = 3;
    int32 damage_dealt = 4;
    int32 damage_taken = 5;
    string result = 6;
}

// ==================== 任务服务 ====================

service TaskService {
    rpc ListTasks(ListTasksRequest) returns (ListTasksResponse);
    rpc AcceptTask(AcceptTaskRequest) returns (Task);
    rpc CompleteTask(CompleteTaskRequest) returns (Task);
    rpc CreateTask(CreateTaskRequest) returns (Task);
}

message ListTasksRequest {
    string status = 1; // open, accepted, all
    string task_type = 2;
}

message ListTasksResponse {
    repeated Task tasks = 1;
}

message Task {
    string id = 1;
    string publisher_id = 2;
    string task_type = 3;
    string title = 4;
    string description = 5;
    bytes requirements = 6;
    bytes rewards = 7;
    string status = 8;
    string acceptor_id = 9;
    string deadline = 10;
}

message AcceptTaskRequest {
    string agent_id = 1;
    string task_id = 2;
}

message CompleteTaskRequest {
    string agent_id = 1;
    string task_id = 2;
    bytes proof = 3; // 完成证明
}

message CreateTaskRequest {
    string agent_id = 1;
    string task_type = 2;
    string title = 3;
    string description = 4;
    bytes requirements = 5;
    bytes rewards = 6;
    string deadline = 7;
}

// ==================== 世界服务 ====================

service WorldService {
    rpc GetWorldInfo(GetWorldInfoRequest) returns (WorldInfo);
    rpc GetChunk(GetChunkRequest) returns (ChunkData);
    rpc GetTime(GetTimeRequest) returns (WorldTime);
    rpc GetBlock(GetBlockRequest) returns (BlockDetail);
    rpc ClaimBlock(ClaimBlockRequest) returns (ClaimBlockResponse);
}

message GetWorldInfoRequest {}

message WorldInfo {
    string id = 1;
    string name = 2;
    string seed = 3;
    int32 time_speed = 4;
    string current_time = 5;
    int64 agent_count = 6;
    int64 building_count = 7;
}

message GetChunkRequest {
    int32 x = 1;
    int32 y = 2;
}

message ChunkData {
    int32 x = 1;
    int32 y = 2;
    bool generated = 3;
    bytes terrain_data = 4; // 压缩数据
}

message GetTimeRequest {}

message WorldTime {
    string timestamp = 1;
    int32 year = 2;
    int32 month = 3;
    int32 day = 4;
    int32 hour = 5;
    int32 minute = 6;
    string season = 7; // spring, summer, autumn, winter
    bool is_daytime = 8;
}

message GetBlockRequest {
    int32 x = 1;
    int32 y = 2;
}

message BlockDetail {
    Position position = 1;
    int32 height = 2;
    string terrain_type = 3;
    string resource_type = 4;
    int32 resource_amount = 5;
    string owner_id = 6;
    repeated BuildingView buildings = 7;
}

message ClaimBlockRequest {
    string agent_id = 1;
    int32 x = 2;
    int32 y = 3;
    int32 price = 4; // 报价
}

message ClaimBlockResponse {
    bool success = 1;
    string error = 2;
    int32 actual_price = 3;
}

// ==================== 经济服务 ====================

service EconomyService {
    rpc Transfer(TransferRequest) returns (TransferResponse);
    rpc GetBalance(GetBalanceRequest) returns (Balance);
    rpc GetTransactions(GetTransactionsRequest) returns (Transactions);
    rpc AddFriend(AddFriendRequest) returns (AddFriendResponse);
    rpc GetFriends(GetFriendsRequest) returns (Friends);
}

message TransferRequest {
    string from_agent_id = 1;
    string to_agent_id = 2;
    int32 amount = 3;
    string description = 4;
}

message TransferResponse {
    bool success = 1;
    string error = 2;
    int32 new_balance = 3;
    string transaction_id = 4;
}

message GetBalanceRequest {
    string agent_id = 1;
}

message Balance {
    string agent_id = 1;
    int32 balance = 2;
    int32 total_income = 3;
    int32 total_expense = 4;
}

message GetTransactionsRequest {
    string agent_id = 1;
    int32 limit = 2;
    int32 offset = 3;
}

message Transactions {
    repeated Transaction transactions = 1;
}

message Transaction {
    string id = 1;
    string from_agent_id = 2;
    string to_agent_id = 3;
    int32 amount = 4;
    string type = 5;
    string description = 6;
    string created_at = 7;
}

message AddFriendRequest {
    string agent_id = 1;
    string friend_id = 2;
}

message AddFriendResponse {
    bool success = 1;
    string error = 2;
}

message GetFriendsRequest {
    string agent_id = 1;
}

message Friends {
    repeated Friend friends = 1;
}

message Friend {
    string agent_id = 1;
    string name = 2;
    bool is_online = 3;
}

// ==================== 管理服务 (Web端) ====================

service ManagementService {
    // TODO管理
    rpc CreateTodo(CreateTodoRequest) returns (Todo);
    rpc UpdateTodo(UpdateTodoRequest) returns (Todo);
    rpc DeleteTodo(DeleteTodoRequest) returns (DeleteTodoResponse);
    
    // 观察数据
    rpc GetWorldMap(GetWorldMapRequest) returns (WorldMap);
    rpc GetAgentDetail(GetAgentDetailRequest) returns (AgentDetail);
    rpc GetSnapshotList(GetSnapshotListRequest) returns (SnapshotList);
    rpc LoadSnapshot(LoadSnapshotRequest) returns (LoadSnapshotResponse);
}

message CreateTodoRequest {
    string todo_token = 1; // Agent生成的Token
    string content = 2;
    int32 priority = 3;
}

message UpdateTodoRequest {
    string todo_token = 1;
    string todo_id = 2;
    string content = 3;
    int32 priority = 4;
}

message DeleteTodoRequest {
    string todo_token = 1;
    string todo_id = 2;
}

message DeleteTodoResponse {
    bool success = 1;
}

message GetWorldMapRequest {
    int32 center_x = 1;
    int32 center_y = 2;
    int32 radius = 3;
}

message WorldMap {
    repeated Block blocks = 1;
    repeated AgentView agents = 2;
    repeated BuildingView buildings = 3;
}

message GetAgentDetailRequest {
    string agent_id = 1;
}

message AgentDetail {
    Agent agent = 1;
    Inventory inventory = 2;
    Skills skills = 3;
    repeated BuildingView buildings = 4;
    repeated Block owned_blocks = 5;
    int32 reputation_score = 6; // 平均名誉分
}

message GetSnapshotListRequest {
    int32 limit = 1;
}

message SnapshotList {
    repeated SnapshotInfo snapshots = 1;
}

message SnapshotInfo {
    string id = 1;
    string created_at = 2;
    int32 agent_count = 3;
    int32 building_count = 4;
}

message LoadSnapshotRequest {
    string snapshot_id = 1;
}

message LoadSnapshotResponse {
    bool success = 1;
    string error = 2;
}
```

### 3.2 HTTP REST API (Web端)

```yaml
# Web端使用HTTP API，与gRPC服务共享后端

# 观察接口
GET /api/world/map?x={x}&y={y}&radius={radius}
  -> 获取世界地图数据

GET /api/agents
  -> 获取所有Agent列表

GET /api/agents/{agent_id}
  -> 获取Agent详情

GET /api/agents/{agent_id}/todos
  -> 获取Agent TODO列表

POST /api/agents/{agent_id}/todos
  Header: X-Todo-Token: {token}
  Body: {content, priority}
  -> 创建TODO

PUT /api/agents/{agent_id}/todos/{todo_id}
  Header: X-Todo-Token: {token}
  Body: {content, priority}
  -> 更新TODO

DELETE /api/agents/{agent_id}/todos/{todo_id}
  Header: X-Todo-Token: {token}
  -> 删除TODO

GET /api/buildings
  -> 获取所有建筑列表

GET /api/snapshots
  -> 获取快照列表

POST /api/snapshots/{snapshot_id}/load
  -> 加载快照（管理员权限）

# WebSocket (可选，用于实时推送)
WS /api/ws
  -> 订阅世界事件
```

---

## 4. 开发路线图

### Phase 1: 基础设施 (Week 1-2)

#### Week 1: 项目搭建
- [ ] 项目结构初始化
- [ ] 数据库Schema设计
- [ ] gRPC proto定义
- [ ] 基础CI/CD

#### Week 2: 核心服务
- [ ] 数据库连接与迁移
- [ ] 基础服务框架
- [ ] 身份认证系统
- [ ] 世界生成算法

### Phase 2: Agent系统 (Week 3-4)

#### Week 3: Agent基础
- [ ] Agent注册/登录
- [ ] 位置管理
- [ ] 视野系统
- [ ] 背包系统

#### Week 4: 生命系统
- [ ] 体力/饥饿系统
- [ ] 死亡与重生
- [ ] 技能系统
- [ ] 基础CLI工具

### Phase 3: 世界与建造 (Week 5-6)

#### Week 5: 世界系统
- [ ] 无限地图生成
- [ ] 区块加载/卸载
- [ ] 资源系统
- [ ] 天气系统（视觉）

#### Week 6: 建造系统
- [ ] 模块化建造
- [ ] 建筑所有权
- [ ] 设施维护
- [ ] 所有权申报

### Phase 4: 经济与社交 (Week 7-8)

#### Week 7: 经济系统
- [ ] 货币系统
- [ ] 交易系统
- [ ] 终端系统
- [ ] 税收系统

#### Week 8: 社交系统
- [ ] 好友系统
- [ ] 名誉系统
- [ ] 任务系统
- [ ] 政府财政

### Phase 5: 战斗系统 (Week 9)

- [ ] PvE战斗
- [ ] PvP宣战机制
- [ ] 回合制战斗逻辑
- [ ] 逃跑机制

### Phase 6: Web端 (Week 10-11)

#### Week 10: Web基础
- [ ] Web框架搭建
- [ ] 3D地图渲染
- [ ] Agent观察界面
- [ ] TODO管理界面

#### Week 11: Web高级
- [ ] 时间控制
- [ ] 快照回放
- [ ] 统计面板
- [ ] 性能优化

### Phase 7: 集成与测试 (Week 12)

- [ ] 集成测试
- [ ] 性能测试
- [ ] 压力测试
- [ ] 文档完善

---

## 5. 目录结构

```
agent-town/
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── docker-compose.yml
│
├── docs/
│   ├── GAME_DESIGN.md
│   └── TECH_DESIGN.md
│
├── proto/
│   └── agenttown.proto
│
├── server/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── api/
│   │   │   ├── grpc/
│   │   │   └── http/
│   │   ├── service/
│   │   │   ├── agent.go
│   │   │   ├── world.go
│   │   │   ├── economy.go
│   │   │   ├── battle.go
│   │   │   ├── build.go
│   │   │   └── task.go
│   │   ├── repository/
│   │   │   ├── agent.go
│   │   │   ├── world.go
│   │   │   └── economy.go
│   │   ├── model/
│   │   │   └── entity.go
│   │   ├── worldgen/
│   │   │   └── generator.go
│   │   └── auth/
│   │       └── jwt.go
│   ├── migrations/
│   └── tests/
│       └── integration/
│
├── cli/
│   ├── cmd/
│   │   └── cli/
│   │       └── main.go
│   ├── internal/
│   │   ├── client/
│   │   │   └── grpc.go
│   │   ├── commands/
│   │   │   ├── register.go
│   │   │   ├── login.go
│   │   │   ├── status.go
│   │   │   ├── move.go
│   │   │   ├── gather.go
│   │   │   ├── build.go
│   │   │   ├── inventory.go
│   │   │   ├── todo.go
│   │   │   └── token.go
│   │   └── config/
│   └── tests/
│
└── web/
    ├── cmd/
    │   └── web/
    │       └── main.go
    ├── internal/
    │   ├── server/
    │   ├── handler/
    │   ├── static/
    │   └── template/
    ├── static/
    │   ├── js/
    │   ├── css/
    │   └── assets/
    └── templates/
```

---

## 6. 关键技术决策

### 6.1 世界生成
- **算法**: Simplex噪声 + 分层地形
- **区块**: 32x32格，按需生成
- **种子**: 全局统一，确保可复现
- **存储**: 压缩后存入PostgreSQL，热区块缓存于Redis

### 6.2 视野系统
- **计算**: Server计算视野范围，只返回可见数据
- **Agent记忆**: CLI自行维护超出视野的"印象地图"
- **优化**: 使用空间索引加速视野查询

### 6.3 时间系统
- **推进**: 独立goroutine推进游戏时间
- **快照**: 每小时自动保存
- **回溯**: 加载历史快照数据

### 6.4 战斗系统
- **回合管理**: 独立BattleService管理回合
- **超时**: 定时器检测回合超时
- **离线**: PvE战斗标记为僵持，不强制结束

### 6.5 CLI通信
- **短连接**: 大多数API使用短连接HTTP/gRPC
- **长连接**: 可选WebSocket订阅事件
- **轮询**: 推荐CLI定期轮询状态

---

## 7. 已确认的技术决策

| 决策 | 方案 | 说明 |
|------|------|------|
| **文件存储** | Server本地 | 蓝图和快照存本地，做好抽象接口，后续可扩展云存储 |
| **事件获取** | 长轮询(Long Polling) | CLI使用HTTP长轮询获取事件，实现简单，兼容性好 |
| **Redis缓存** | 热点数据 | 缓存在线Agent、热门区块等频繁访问数据 |
| **分库分表** | 后续扩展 | 当前单库，后期数据量大时考虑 |
| **监控告警** | 暂不考虑 | MVP阶段不接入监控 |

## 8. 长轮询实现方案

### 8.1 流程
```
CLI                                    Server
 |                                       |
 |--- GET /events?agent_id=xxx ---------->| 
 |                                       | 保持连接
 |                                       | 等待事件
 |                                       | (30秒超时)
 |                                       |
 |<-- 200 OK {events: [...]} -------------| 有事件立即返回
 |                                       |
 |--- GET /events?agent_id=xxx ---------->| 立即重新请求
 |                                       | 保持连接
 |                                       | ...
 |                                       |
 |<-- 200 OK {events: []} ----------------| 超时返回空
 |                                       |
 |--- GET /events?agent_id=xxx ---------->| 立即重新请求
```

### 8.2 优点
- **简单**: 基于HTTP，无需WebSocket握手
- **兼容**: 防火墙友好，穿透性好
- **容错**: 超时后自动重试，无需额外心跳

### 8.3 缺点
- **延迟**: 事件发生后到下次请求才能获取（最大延迟=轮询间隔）
- **开销**: 频繁建立连接（但HTTP/1.1 keep-alive可复用连接）

### 8.4 在我们的场景中适用
- CLI是非交互式调用，不需要超低延迟
- 游戏时间是可配置的（1分钟=5分钟），秒级延迟可接受
- 实现简单，减少复杂度

## 9. 文件存储抽象

### 9.1 接口设计
```go
type FileStore interface {
    // 保存蓝图
    SaveBlueprint(bp *Blueprint) (string, error)
    // 读取蓝图
    GetBlueprint(id string) (*Blueprint, error)
    // 删除蓝图
    DeleteBlueprint(id string) error
    
    // 保存快照
    SaveSnapshot(snap *Snapshot) (string, error)
    // 读取快照
    GetSnapshot(id string) (*Snapshot, error)
    // 列出现有快照
    ListSnapshots() ([]Snapshot, error)
}

// 本地实现
type LocalFileStore struct {
    basePath string
}

// 后续可扩展 S3FileStore
// type S3FileStore struct { ... }
```

### 9.2 目录结构
```
data/
├── blueprints/
│   ├── {agent_id}/
│   │   ├── {blueprint_id}.json
│   │   └── ...
│   └── ...
└── snapshots/
    ├── snapshot_{timestamp}.db
    └── ...
```
