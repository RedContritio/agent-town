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
		}
	}
	
	return blocks
}

// getTerrainType 根据噪声值获取地形类型
func (tg *TerrainGenerator) getTerrainType(value float64) string {
	switch {
	case value < -0.4:
		return TerrainWater
	case value < -0.2:
		return TerrainSand
	case value > 0.6:
		return TerrainHill
	default:
		// 在草地中随机分布农田
		// 使用另一个噪声值来决定
		farmlandNoise := tg.noise.Noise2D(value*10, value*10)
		if farmlandNoise > 0.6 {
			return TerrainFarmland
		}
		return TerrainGrass
	}
}

// getHeight 根据噪声值和地形类型计算高度
func (tg *TerrainGenerator) getHeight(value float64, terrainType string) int {
	switch terrainType {
	case TerrainWater:
		return 0
	case TerrainSand:
		return 0
	case TerrainGrass, TerrainFarmland:
		// 基础高度 0-1
		height := int((value + 0.2) * 2)
		if height < 0 {
			height = 0
		}
		if height > 1 {
			height = 1
		}
		return height
	case TerrainHill:
		// 山丘 2-3
		height := int((value-0.6)*5) + 2
		if height < 2 {
			height = 2
		}
		if height > 3 {
			height = 3
		}
		return height
	default:
		return 0
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
