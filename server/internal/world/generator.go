package world

import (
	"fmt"
	"time"
)

// WorldGenerator 世界生成器总入口
type WorldGenerator struct {
	seed      int64
	terrain   *TerrainGenerator
	resources *ResourceGenerator
	roads     *RoadGenerator
	buildings *BuildingGenerator
	agents    *AgentGenerator
}

// NewWorldGenerator 创建世界生成器
func NewWorldGenerator(seed int64) *WorldGenerator {
	return &WorldGenerator{
		seed:      seed,
		terrain:   NewTerrainGenerator(seed),
		resources: NewResourceGenerator(seed),
		roads:     NewRoadGenerator(),
		buildings: NewBuildingGenerator(seed),
		agents:    NewAgentGenerator(seed),
	}
}

// GenerateInitialWorld 生成初始世界
// 生成 3×3 区块（9 个区块）和四个初始建筑
func (wg *WorldGenerator) GenerateInitialWorld() (*World, error) {
	world := &World{
		ID:        fmt.Sprintf("world-%d", time.Now().Unix()),
		Seed:      wg.seed,
		CreatedAt: time.Now(),
	}
	
	// 1. 生成初始四建筑
	world.InitialBuildings = wg.buildings.GenerateInitialBuildings()
	
	// 初始化区块管理器
	if world.ChunkManager == nil {
		world.ChunkManager = &ChunkManager{
			seed:      wg.seed,
			cache:     make(map[string]*Chunk),
			generator: wg,
			chunkSize: ChunkSize,
		}
	}
	
	// 2. 生成 3×3 区块（中心区块 (0,0) 和周围 8 个区块）
	chunkCoords := [][2]int{
		{-1, -1}, {0, -1}, {1, -1},
		{-1, 0},  {0, 0},  {1, 0},
		{-1, 1},  {0, 1},  {1, 1},
	}
	
	for _, coord := range chunkCoords {
		cx, cy := coord[0], coord[1]
		chunk, err := wg.GenerateChunk(cx, cy)
		if err != nil {
			return nil, err
		}
		world.ChunkManager.cache[fmt.Sprintf("%d,%d", cx, cy)] = chunk
	}
	
	// 3. 应用道路到所有区块
	roads := wg.roads.GenerateInitialRoads()
	world.InitialRoads = roads
	
	for _, road := range roads {
		// 获取道路所在的区块坐标和区块内坐标
		cx, cy, bx, by := world.ChunkManager.WorldToChunk(road.X, road.Y)
		chunkKey := fmt.Sprintf("%d,%d", cx, cy)
		
		if chunk, ok := world.ChunkManager.cache[chunkKey]; ok {
			key := fmt.Sprintf("%d,%d", bx, by)
			if block, ok := chunk.Blocks[key]; ok {
				block.TerrainType = TerrainRoad
				chunk.Blocks[key] = block
			}
		}
	}
	
	// 4. 应用建筑地基到对应区块
	for _, building := range world.InitialBuildings {
		// 将建筑区域设为地基
		for dx := 0; dx < building.Width; dx++ {
			for dy := 0; dy < building.Depth; dy++ {
				x := building.Anchor.X + dx
				y := building.Anchor.Y + dy
				
				// 转换为区块坐标
				cx, cy, bx, by := world.ChunkManager.WorldToChunk(x, y)
				chunkKey := fmt.Sprintf("%d,%d", cx, cy)
				
				if chunk, ok := world.ChunkManager.cache[chunkKey]; ok {
					key := fmt.Sprintf("%d,%d", bx, by)
					if block, ok := chunk.Blocks[key]; ok {
						block.TerrainType = TerrainFoundation
						block.Z = 0
						block.Height = 0
						chunk.Blocks[key] = block
					}
				}
			}
		}
	}
	
	// 5. 生成调试 NPC（在建筑附近）
	spawnPoints := []Position{
		{X: -10, Y: -10}, // 市政厅附近
		{X: 10, Y: -10},  // 银行附近
		{X: -10, Y: 10},  // 任务中心附近
		{X: 10, Y: 10},   // 商店附近
		{X: 0, Y: 0},     // 中心
	}
	world.InitialAgents = wg.agents.GenerateDebugNPCs(spawnPoints, 6)
	
	return world, nil
}

// GenerateChunk 按需生成新区块（仅地形和资源）
func (wg *WorldGenerator) GenerateChunk(cx, cy int) (*Chunk, error) {
	chunk := &Chunk{
		X:          cx,
		Y:          cy,
		Blocks:     make(map[string]Block),
		Generated:  true,
		LastAccess: time.Now(),
	}
	
	// 1. 生成地形
	terrainBlocks := wg.terrain.GenerateChunk(cx, cy)
	
	// 2. 生成资源
	resources := wg.resources.GenerateResources(cx, cy, terrainBlocks)
	
	// 3. 合并地形和资源
	for key, block := range terrainBlocks {
		if resource, ok := resources[key]; ok {
			block.ResourceType = resource.Type
			block.ResourceAmount = resource.Amount
		}
		chunk.Blocks[key] = block
	}
	
	return chunk, nil
}
