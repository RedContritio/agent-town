package world

import (
	"fmt"
)

// 地形类型常量
const (
	TerrainGrass      = "grass"
	TerrainRoad       = "road"
	TerrainWater      = "water"
	TerrainFarmland   = "farmland"
	TerrainSand       = "sand"
	TerrainHill       = "hill"
	TerrainFoundation = "foundation"
)

// TerrainGenerator 地形生成器
type TerrainGenerator struct {
	seed  int64
	noise *SimplexNoise
}

// NewTerrainGenerator 创建地形生成器
func NewTerrainGenerator(seed int64) *TerrainGenerator {
	return &TerrainGenerator{
		seed:  seed,
		noise: NewSimplexNoise(seed),
	}
}

// GenerateChunk 生成单个区块的地形
func (tg *TerrainGenerator) GenerateChunk(cx, cy int) map[string]Block {
	blocks := make(map[string]Block)
	cm := &ChunkManager{chunkSize: ChunkSize}
	
	minHeight, maxHeight := 999, -999
	heightDist := make(map[int]int)
	
	for bx := 0; bx < ChunkSize; bx++ {
		for by := 0; by < ChunkSize; by++ {
			x, y := cm.ChunkToWorld(cx, cy, bx, by)
			
			// 使用分形噪声生成高度
			heightValue := tg.noise.FractalNoise(
				float64(x)*0.05,
				float64(y)*0.05,
				4,   // octaves
				0.5, // persistence
				2.0, // lacunarity
			)
			
			// 根据高度决定地形类型
			terrainType := tg.getTerrainType(heightValue)
			
			// 计算实际高度
			height := tg.getHeight(heightValue, terrainType)
			
			block := Block{
				X:           x,
				Y:           y,
				Z:           height,
				Height:      height,
				TerrainType: terrainType,
			}
			
			key := fmt.Sprintf("%d,%d", bx, by)
			blocks[key] = block
			
			// Track stats
			if height < minHeight { minHeight = height }
			if height > maxHeight { maxHeight = height }
			heightDist[height]++
		}
	}
	
	fmt.Printf("[Terrain] Generated chunk (%d,%d): height range [%d, %d], dist=%v\n", 
		cx, cy, minHeight, maxHeight, heightDist)
	
	return blocks
}

// getTerrainType 根据噪声值获取地形类型
func (tg *TerrainGenerator) getTerrainType(value float64) string {
	switch {
	case value < -0.35:
		return TerrainWater
	case value < -0.15:
		return TerrainSand
	case value > 0.5:
		return TerrainHill
	default:
		// 在草地中随机分布农田
		// 使用另一个噪声值来决定
		farmlandNoise := tg.noise.Noise2D(value*10+100, value*10+200)
		if farmlandNoise > 0.5 {
			return TerrainFarmland
		}
		return TerrainGrass
	}
}

// getHeight 根据噪声值和地形类型计算高度
// 使用连续的噪声值映射，产生更自然的高度变化
func (tg *TerrainGenerator) getHeight(value float64, terrainType string) int {
	switch terrainType {
	case TerrainWater:
		// 水底有轻微起伏，从 -2 到 -1（显示为 0）
		return -2 + int((value+1)*0.5)
	case TerrainSand:
		// 沙滩从 0 开始，逐渐上升到 1
		normalized := (value + 0.35) / 0.2 // 归一化到 0-1
		return int(normalized * 1.5)
	case TerrainGrass, TerrainFarmland:
		// 草地/农田高度 1-3，基于噪声值平滑过渡
		// value 范围大约是 -0.15 到 0.5
		normalized := (value + 0.15) / 0.65 // 归一化到 0-1
		height := 1 + int(normalized*2.5)   // 1 到 3
		if height < 1 {
			height = 1
		}
		if height > 3 {
			height = 3
		}
		return height
	case TerrainHill:
		// 山丘 3-5，随噪声值增加
		normalized := (value - 0.5) / 0.5 // 归一化到 0-1
		height := 3 + int(normalized*3)   // 3 到 5
		if height < 3 {
			height = 3
		}
		if height > 5 {
			height = 5
		}
		return height
	default:
		return 1
	}
}

// ModifyBlock 修改地块（用于挖掘/填海等）
func (tg *TerrainGenerator) ModifyBlock(block *Block, newType string) {
	block.TerrainType = newType
	// 根据新类型重新计算高度
	block.Height = tg.getHeightForType(newType)
	block.Z = block.Height
}

// getHeightForType 获取地形类型的默认高度
func (tg *TerrainGenerator) getHeightForType(terrainType string) int {
	switch terrainType {
	case TerrainWater, TerrainSand:
		return 0
	case TerrainGrass, TerrainFarmland:
		return 1
	case TerrainHill:
		return 2
	default:
		return 0
	}
}
