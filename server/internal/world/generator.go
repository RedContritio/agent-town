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
func (wg *WorldGenerator) GenerateInitialWorld() (*World, error) {
	world := &World{
		ID:        fmt.Sprintf("world-%d", time.Now().Unix()),
		Seed:      wg.seed,
		CreatedAt: time.Now(),
	}
	
	// 1. 生成初始四建筑
	world.InitialBuildings = wg.buildings.GenerateInitialBuildings()
	
	// 2. 生成初始区块 (0,0) 的地形
	initialChunk, err := wg.GenerateChunk(0, 0)
	if err != nil {
		return nil, err
	}
	
	// 3. 应用道路到初始区块
	roads := wg.roads.GenerateInitialRoads()
	world.InitialRoads = roads
	
	for _, road := range roads {
		// 将道路坐标转换为区块内坐标
		bx, by := road.X+16, road.Y+16 // 偏移到 0-31 范围
		if bx >= 0 && bx < ChunkSize && by >= 0 && by < ChunkSize {
			key := fmt.Sprintf("%d,%d", bx, by)
			if block, ok := initialChunk.Blocks[key]; ok {
				block.TerrainType = TerrainRoad
				initialChunk.Blocks[key] = block
			}
		}
	}
	
	// 4. 应用建筑地基到初始区块
	for _, building := range world.InitialBuildings {
		// 将建筑区域设为地基
		for dx := 0; dx < building.Width; dx++ {
			for dy := 0; dy < building.Depth; dy++ {
				x := building.Anchor.X + dx
				y := building.Anchor.Y + dy
				bx, by := x+16, y+16
				if bx >= 0 && bx < ChunkSize && by >= 0 && by < ChunkSize {
					key := fmt.Sprintf("%d,%d", bx, by)
					if block, ok := initialChunk.Blocks[key]; ok {
						block.TerrainType = TerrainFoundation
						block.Z = 0
						block.Height = 0
						initialChunk.Blocks[key] = block
					}
				}
			}
		}
	}
	
	// 保存初始区块
	if world.ChunkManager == nil {
		world.ChunkManager = &ChunkManager{
			seed:      wg.seed,
			cache:     make(map[string]*Chunk),
			generator: wg,
			chunkSize: ChunkSize,
		}
	}
	world.ChunkManager.cache["0,0"] = initialChunk
	
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
