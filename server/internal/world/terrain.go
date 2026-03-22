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
// 初始区块需要非常平坦以容纳基础建筑
func (tg *TerrainGenerator) GenerateChunk(cx, cy int) map[string]Block {
	blocks := make(map[string]Block)
	cm := &ChunkManager{chunkSize: ChunkSize}

	minHeight, maxHeight := 999, -999
	heightDist := make(map[int]int)

	// 判断是否为初始区块（中心3x3区域）
	isInitialArea := cx >= -1 && cx <= 1 && cy >= -1 && cy <= 1

	for bx := 0; bx < ChunkSize; bx++ {
		for by := 0; by < ChunkSize; by++ {
			x, y := cm.ChunkToWorld(cx, cy, bx, by)

			var height int
			var terrainType string

			if isInitialArea {
				// 初始区域：完全平坦，固定高度为1
				height = 1
				terrainType = TerrainGrass
				// 在边缘添加少量水域作为边界
				if cx == -1 && bx < 4 || cx == 1 && bx > 27 ||
					cy == -1 && by < 4 || cy == 1 && by > 27 {
					// 边缘区域保持原有生成逻辑
					heightValue := tg.noise.FractalNoise(
						float64(x)*0.02,
						float64(y)*0.02,
						3, 0.4, 2.0,
					)
					terrainType = tg.getTerrainType(heightValue)
					height = tg.getHeight(heightValue, terrainType)
				}
			} else {
				// 外部区域：使用噪声生成，但也很平坦
				heightValue := tg.noise.FractalNoise(
					float64(x)*0.015, // 更低频率
					float64(y)*0.015,
					2,   // 更少octaves
					0.3, // 更低persistence
					2.0,
				)
				terrainType = tg.getTerrainType(heightValue)
				height = tg.getHeight(heightValue, terrainType)
			}

			block := Block{
				X:           x,
				Y:           y,
				Z:           height,
				Height:      height,
				TerrainType: terrainType,
			}

			key := fmt.Sprintf("%d,%d", bx, by)
			blocks[key] = block

			if height < minHeight {
				minHeight = height
			}
			if height > maxHeight {
				maxHeight = height
			}
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
	case value < -0.6:
		return TerrainWater
	case value < -0.4:
		return TerrainSand
	case value > 0.5:
		return TerrainHill
	default:
		// 减少农田分布
		farmlandNoise := tg.noise.Noise2D(value*10+100, value*10+200)
		if farmlandNoise > 0.7 {
			return TerrainFarmland
		}
		return TerrainGrass
	}
}

// getHeight 根据噪声值和地形类型计算高度
// 非常平坦的高度分布
func (tg *TerrainGenerator) getHeight(value float64, terrainType string) int {
	switch terrainType {
	case TerrainWater:
		return -1
	case TerrainSand:
		return 0
	case TerrainGrass, TerrainFarmland:
		return 1 // 固定高度1
	case TerrainHill:
		return 2 // 山丘也降低
	default:
		return 1
	}
}

// ModifyBlock 修改地块（用于挖掘/填海等）
func (tg *TerrainGenerator) ModifyBlock(block *Block, newType string) {
	block.TerrainType = newType
	block.Height = tg.getHeightForType(newType)
	block.Z = block.Height
}

// getHeightForType 获取地形类型的默认高度
func (tg *TerrainGenerator) getHeightForType(terrainType string) int {
	switch terrainType {
	case TerrainWater:
		return -1
	case TerrainSand:
		return 0
	case TerrainGrass, TerrainFarmland:
		return 1
	case TerrainHill:
		return 2
	default:
		return 1
	}
}
