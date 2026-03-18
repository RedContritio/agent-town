package world

import (
	"database/sql"
	"time"
)

// World 世界状态
type World struct {
	ID               string
	Seed             int64
	CreatedAt        time.Time
	InitialBuildings []Building
	InitialAgents    []Agent
	InitialRoads     []Road
	ChunkManager     *ChunkManager
}

// worldState 全局世界状态（单例）
var worldState *World

// InitWorld 初始化世界
func InitWorld(seed int64, db *sql.DB) error {
	generator := NewWorldGenerator(seed)
	
	world, err := generator.GenerateInitialWorld()
	if err != nil {
		return err
	}
	
	// 设置数据库连接
	world.ChunkManager.db = db
	
	worldState = world
	return nil
}

// GetWorld 获取世界状态
func GetWorld() *World {
	return worldState
}

// GetBuildingByID 根据 ID 获取建筑
func (w *World) GetBuildingByID(id string) *Building {
	for _, b := range w.InitialBuildings {
		if b.ID == id {
			return &b
		}
	}
	return nil
}

// GetAgentByID 根据 ID 获取 Agent
func (w *World) GetAgentByID(id string) *Agent {
	for _, a := range w.InitialAgents {
		if a.ID == id {
			return &a
		}
	}
	return nil
}

// GetAllBuildings 获取所有建筑（包括初始建筑）
func (w *World) GetAllBuildings() []Building {
	// TODO: 后续支持动态建筑时，需要合并初始建筑和动态建筑
	return w.InitialBuildings
}

// GetAllAgents 获取所有 Agent（包括初始 Agents）
func (w *World) GetAllAgents() []Agent {
	// TODO: 后续支持动态 Agent 时，需要合并
	return w.InitialAgents
}

// GetBlocksInRadius 获取指定范围内的地块
func (w *World) GetBlocksInRadius(x, y, radius int) ([]Block, error) {
	if w.ChunkManager == nil {
		return nil, nil
	}
	return w.ChunkManager.GetBlocksInRadius(x, y, radius)
}

// GetBlock 获取单个地块
func (w *World) GetBlock(x, y int) (*Block, error) {
	if w.ChunkManager == nil {
		return nil, nil
	}
	return w.ChunkManager.GetBlock(x, y)
}
