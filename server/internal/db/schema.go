// Package db 数据库表结构定义
// 遵循 DATABASE_DESIGN.md 的规范
package db

import (
	"database/sql"
	"fmt"
)

// createSchema 创建所有表和索引
func createSchema(db *sql.DB) error {
	for _, sql := range createTableSQLs {
		if _, err := db.Exec(sql); err != nil {
			return fmt.Errorf("执行 SQL 失败 [%s...]: %w", sql[:50], err)
		}
	}
	return nil
}

// createTableSQLs 所有建表 SQL 语句
var createTableSQLs = []string{
	// 1. agents - Agent 身份与状态
	`CREATE TABLE IF NOT EXISTS agents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_key BLOB UNIQUE NOT NULL,
		name TEXT UNIQUE NOT NULL,
		balance INTEGER DEFAULT 0,
		position_x INTEGER DEFAULT 0,
		position_y INTEGER DEFAULT 0,
		facing_angle INTEGER DEFAULT 0,
		hp INTEGER DEFAULT 100,
		max_hp INTEGER DEFAULT 100,
		stamina INTEGER DEFAULT 100,
		max_stamina INTEGER DEFAULT 100,
		hunger INTEGER DEFAULT 100,
		max_hunger INTEGER DEFAULT 100,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	)`,

	// 2. tokens - Token 认证
	`CREATE TABLE IF NOT EXISTS tokens (
		token TEXT PRIMARY KEY,
		agent_id INTEGER NOT NULL,
		token_type INTEGER DEFAULT 0,
		scopes INTEGER DEFAULT 0,
		created_at INTEGER NOT NULL,
		expires_at INTEGER,
		last_used_at INTEGER,
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	)`,

	// 3. tasks - 任务栈（LIFO）
	`CREATE TABLE IF NOT EXISTS tasks (
		agent_id INTEGER NOT NULL,
		seq INTEGER NOT NULL,
		type INTEGER NOT NULL,
		status INTEGER DEFAULT 0,
		params TEXT,
		stack_depth INTEGER DEFAULT 0,
		result TEXT,
		error_code INTEGER,
		created_at INTEGER NOT NULL,
		started_at INTEGER,
		completed_at INTEGER,
		PRIMARY KEY (agent_id, seq),
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	)`,

	// 4. inventory - 背包
	`CREATE TABLE IF NOT EXISTS inventory (
		agent_id INTEGER NOT NULL,
		slot INTEGER NOT NULL,
		item_type INTEGER NOT NULL,
		quantity INTEGER DEFAULT 0,
		PRIMARY KEY (agent_id, slot),
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	)`,

	// 5. skills - 技能
	`CREATE TABLE IF NOT EXISTS skills (
		agent_id INTEGER NOT NULL,
		skill_type INTEGER NOT NULL,
		level INTEGER DEFAULT 0,
		exp INTEGER DEFAULT 0,
		PRIMARY KEY (agent_id, skill_type),
		FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
	)`,

	// 6. item_types - 物品类型枚举
	`CREATE TABLE IF NOT EXISTS item_types (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		category INTEGER DEFAULT 0,
		stack_max INTEGER DEFAULT 99
	)`,

	// 7. land_ownership - 土地所有权
	`CREATE TABLE IF NOT EXISTS land_ownership (
		x INTEGER NOT NULL,
		y INTEGER NOT NULL,
		owner_id INTEGER,
		acquired_at INTEGER,
		price INTEGER DEFAULT 0,
		PRIMARY KEY (x, y),
		FOREIGN KEY (owner_id) REFERENCES agents(id) ON DELETE SET NULL
	)`,

	// 8. land_tiles - 地块地形
	`CREATE TABLE IF NOT EXISTS land_tiles (
		x INTEGER NOT NULL,
		y INTEGER NOT NULL,
		terrain_type INTEGER DEFAULT 0,
		height INTEGER DEFAULT 0,
		PRIMARY KEY (x, y)
	)`,

	// 9. resources - 动态资源
	`CREATE TABLE IF NOT EXISTS resources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		x INTEGER NOT NULL,
		y INTEGER NOT NULL,
		resource_type INTEGER NOT NULL,
		state INTEGER DEFAULT 0,
		amount INTEGER DEFAULT 0,
		created_at INTEGER NOT NULL,
		owner_id INTEGER,
		FOREIGN KEY (owner_id) REFERENCES agents(id) ON DELETE SET NULL
	)`,

	// 10. buildings - 建筑
	`CREATE TABLE IF NOT EXISTS buildings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		owner_id INTEGER NOT NULL,
		x INTEGER NOT NULL,
		y INTEGER NOT NULL,
		z INTEGER DEFAULT 0,
		building_type INTEGER NOT NULL,
		name TEXT,
		durability INTEGER DEFAULT 100,
		max_durability INTEGER DEFAULT 100,
		created_at INTEGER NOT NULL,
		FOREIGN KEY (owner_id) REFERENCES agents(id) ON DELETE CASCADE
	)`,

	// 索引
	`CREATE INDEX IF NOT EXISTS idx_tasks_agent_status ON tasks(agent_id, status)`,
	`CREATE INDEX IF NOT EXISTS idx_tokens_agent ON tokens(agent_id)`,
	`CREATE INDEX IF NOT EXISTS idx_agents_pos ON agents(position_x, position_y)`,
	`CREATE INDEX IF NOT EXISTS idx_resources_pos ON resources(x, y)`,
	`CREATE INDEX IF NOT EXISTS idx_buildings_owner ON buildings(owner_id)`,
	`CREATE INDEX IF NOT EXISTS idx_buildings_pos ON buildings(x, y)`,
}

// 枚举常量定义

// TokenType Token 类型
const (
	TokenTypeCLI = 0
	TokenTypeWeb = 1
)

// TaskType 任务类型
const (
	TaskTypeMove    = 0
	TaskTypeHarvest = 1
	TaskTypeCraft   = 2
	TaskTypeBuild   = 3
	TaskTypeCombat  = 4
)

// TaskStatus 任务状态
const (
	TaskStatusPending   = 0
	TaskStatusRunning   = 1
	TaskStatusPaused    = 2
	TaskStatusCompleted = 3
	TaskStatusFailed    = 4
)

// SkillType 技能类型
const (
	SkillTypeFarming  = 0
	SkillTypeMining   = 1
	SkillTypeBuilding = 2
	SkillTypeCrafting = 3
	SkillTypeCombat   = 4
)

// ItemCategory 物品分类
const (
	ItemCategoryResource = 0
	ItemCategoryTool     = 1
	ItemCategoryFood     = 2
	ItemCategoryMaterial = 3
)

// TerrainType 地形类型
const (
	TerrainTypeGrass    = 0
	TerrainTypeRoad     = 1
	TerrainTypeWater    = 2
	TerrainTypeFarmland = 3
	TerrainTypeSand     = 4
	TerrainTypeHills    = 5
)

// ResourceType 资源类型
const (
	ResourceTypeTree  = 0
	ResourceTypeMine  = 1
	ResourceTypeWheat = 2
	ResourceTypeCorn  = 3
)

// ResourceState 资源状态
const (
	ResourceStateSeed    = 0
	ResourceStateGrowing = 1
	ResourceStateMature  = 2
	ResourceStateDepleted = 3
	ResourceStateWithered = 4
)

// BuildingType 建筑类型
const (
	BuildingTypeHouse    = 0
	BuildingTypeShop     = 1
	BuildingTypeWorkshop = 2
	BuildingTypeBank     = 3
)

// Scope 权限位掩码
const (
	ScopeRead   = 0x1  // 1
	ScopeWrite  = 0x2  // 2
	ScopeCombat = 0x4  // 4
	ScopeTrade  = 0x8  // 8
	ScopeTodo   = 0x10 // 16
)

// HasScope 检查权限位
func HasScope(scopes, scope int) bool {
	return scopes&scope != 0
}

// AddScope 添加权限
func AddScope(scopes, scope int) int {
	return scopes | scope
}

// RemoveScope 移除权限
func RemoveScope(scopes, scope int) int {
	return scopes & ^scope
}
