# Agent-Town 数据库设计文档

## 设计原则

1. **单文件 SQLite**：当前阶段使用单文件存储，简化部署
2. **全整数主键**：所有 ID 使用自增整数，外键关联
3. **枚举类型用整数**：状态、类型等用整数表示，提升查询效率
4. **最小 TEXT 字段**：仅保留必要的字符串（名称、UUID、JSON 参数）
5. **时间戳统一用 INTEGER**：Unix 时间戳（毫秒）

---

## 表清单

共 10 张表，分为 4 个模块：

| 模块 | 表名 | 说明 |
|------|------|------|
| **身份认证** | `agents` | Agent 身份与状态 |
| | `tokens` | Token 认证 |
| **任务系统** | `tasks` | 任务栈（LIFO）|
| **物品技能** | `inventory` | 背包物品 |
| | `skills` | 技能等级 |
| | `item_types` | 物品类型枚举 |
| **世界** | `land_ownership` | 土地所有权 |
| | `land_tiles` | 地块地形（地形类型、高度）|
| | `resources` | 动态资源（树、矿、农作物）|
| | `buildings` | 建筑 |

---

## 详细表结构

### 1. agents（Agent 身份）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | INTEGER | 主键，自增 |
| `public_key` | BLOB | ed25519 公钥，唯一索引 |
| `name` | TEXT | 显示名称，唯一 |
| `balance` | INTEGER | 金币数量 |
| `position_x` | INTEGER | X 坐标 |
| `position_y` | INTEGER | Y 坐标 |
| `facing_angle` | INTEGER | 朝向角度 0-359 |
| `hp` | INTEGER | 当前生命值 |
| `max_hp` | INTEGER | 最大生命值 |
| `stamina` | INTEGER | 当前体力 |
| `max_stamina` | INTEGER | 最大体力 |
| `hunger` | INTEGER | 当前饥饿值 |
| `max_hunger` | INTEGER | 最大饥饿值 |
| `created_at` | INTEGER | 创建时间（毫秒时间戳）|
| `updated_at` | INTEGER | 更新时间（毫秒时间戳）|

### 2. tokens（Token 认证）

| 字段 | 类型 | 说明 |
|------|------|------|
| `token` | TEXT | UUID，主键 |
| `agent_id` | INTEGER | 外键 → agents.id |
| `token_type` | INTEGER | 0=cli, 1=web |
| `scopes` | INTEGER | 权限位掩码 |
| `created_at` | INTEGER | 创建时间 |
| `expires_at` | INTEGER | 过期时间，NULL 表示永不过期 |
| `last_used_at` | INTEGER | 最后使用时间 |

**权限位掩码（scopes）**：
- `0x1` (1) = read
- `0x2` (2) = write
- `0x4` (4) = combat
- `0x8` (8) = trade
- `0x10` (16) = todo

### 3. tasks（任务栈）

| 字段 | 类型 | 说明 |
|------|------|------|
| `agent_id` | INTEGER | 外键 → agents.id，复合主键第一部分 |
| `seq` | INTEGER | Agent 内序列号，复合主键第二部分 |
| `type` | INTEGER | 0=move, 1=harvest, 2=craft, 3=build, 4=combat |
| `status` | INTEGER | 0=pending, 1=running, 2=paused, 3=completed, 4=failed |
| `params` | TEXT | JSON 参数（路径、目标等）|
| `stack_depth` | INTEGER | 栈中的位置，0=栈顶 |
| `result` | TEXT | JSON 结果 |
| `error_code` | INTEGER | 错误码 |
| `created_at` | INTEGER | 创建时间 |
| `started_at` | INTEGER | 开始执行时间 |
| `completed_at` | INTEGER | 完成时间 |

### 4. inventory（背包）

| 字段 | 类型 | 说明 |
|------|------|------|
| `agent_id` | INTEGER | 外键 → agents.id，复合主键第一部分 |
| `slot` | INTEGER | 背包格子 0-39，复合主键第二部分 |
| `item_type` | INTEGER | 外键 → item_types.id |
| `quantity` | INTEGER | 数量 |

### 5. skills（技能）

| 字段 | 类型 | 说明 |
|------|------|------|
| `agent_id` | INTEGER | 外键 → agents.id，复合主键第一部分 |
| `skill_type` | INTEGER | 0=farming, 1=mining, 2=building, 3=crafting, 4=combat... |
| `level` | INTEGER | 等级 0-10 |
| `exp` | INTEGER | 当前经验值 |

### 6. item_types（物品类型枚举）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | INTEGER | 主键，物品类型 ID |
| `name` | TEXT | 显示名称（"木材"、"铁矿石"）|
| `category` | INTEGER | 0=resource, 1=tool, 2=food, 3=material... |
| `stack_max` | INTEGER | 最大堆叠数量 |

### 7. land_ownership（土地所有权）

| 字段 | 类型 | 说明 |
|------|------|------|
| `x` | INTEGER | 坐标 X，复合主键第一部分 |
| `y` | INTEGER | 坐标 Y，复合主键第二部分 |
| `owner_id` | INTEGER | 外键 → agents.id，NULL 表示政府所有 |
| `acquired_at` | INTEGER | 获得时间 |
| `price` | INTEGER | 购买价格 |

### 8. land_tiles（地块地形）

| 字段 | 类型 | 说明 |
|------|------|------|
| `x` | INTEGER | 坐标 X，复合主键第一部分 |
| `y` | INTEGER | 坐标 Y，复合主键第二部分 |
| `terrain_type` | INTEGER | 0=grass, 1=road, 2=water, 3=farmland, 4=sand, 5=hills... |
| `height` | INTEGER | 高度 |

### 9. resources（动态资源）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | INTEGER | 主键，自增 |
| `x` | INTEGER | 坐标 X |
| `y` | INTEGER | 坐标 Y |
| `resource_type` | INTEGER | 0=tree, 1=mine, 2=wheat, 3=corn... |
| `state` | INTEGER | 0=seed, 1=growing, 2=mature, 3=depleted, 4=withered |
| `amount` | INTEGER | 剩余数量 |
| `created_at` | INTEGER | 生成/种植时间 |
| `owner_id` | INTEGER | 外键 → agents.id，种植者，NULL 表示野生 |

### 10. buildings（建筑）

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | INTEGER | 主键，自增 |
| `owner_id` | INTEGER | 外键 → agents.id |
| `x` | INTEGER | 坐标 X |
| `y` | INTEGER | 坐标 Y |
| `z` | INTEGER | 高度/楼层 |
| `building_type` | INTEGER | 0=house, 1=shop, 2=workshop, 3=bank... |
| `name` | TEXT | 自定义名称 |
| `durability` | INTEGER | 当前耐久度 |
| `max_durability` | INTEGER | 最大耐久度 |
| `created_at` | INTEGER | 创建时间 |

---

## 索引

```sql
-- 任务查询
CREATE INDEX idx_tasks_agent_status ON tasks(agent_id, status);

-- Token 查询
CREATE INDEX idx_tokens_agent ON tokens(agent_id);

-- 位置查询
CREATE INDEX idx_agents_pos ON agents(position_x, position_y);

-- 资源位置查询
CREATE INDEX idx_resources_pos ON resources(x, y);

-- 建筑查询
CREATE INDEX idx_buildings_owner ON buildings(owner_id);
CREATE INDEX idx_buildings_pos ON buildings(x, y);
```

---

## 扩展预留

未来如需分库分表，可按以下维度拆分：

| 表 | 分片键 | 说明 |
|----|--------|------|
| `agents`, `tasks`, `inventory`, `skills` | `id` / `agent_id` | 按 Agent 分片 |
| `land_tiles`, `resources`, `buildings` | `x,y` 坐标 | 按地理区块分片 |
| `tokens`, `item_types` | 不分片 | 全局表 |

当前单文件 SQLite 足够支撑 1-10k Agent，分片为预留设计。
