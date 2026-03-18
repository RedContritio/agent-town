package world

import (
	"math/rand"
)

// Resource 资源
type Resource struct {
	Type   string `json:"type"`
	Amount int    `json:"amount"`
}

// ResourceGenerator 资源生成器
type ResourceGenerator struct {
	seed  int64
	noise *SimplexNoise
	rng   *rand.Rand
}

// NewResourceGenerator 创建资源生成器
func NewResourceGenerator(seed int64) *ResourceGenerator {
	return &ResourceGenerator{
		seed:  seed,
		noise: NewSimplexNoise(seed + 12345), // 使用不同种子避免与地形重合
		rng:   rand.New(rand.NewSource(seed)),
	}
}

// GenerateResources 为区块生成资源
// 基于区块坐标和噪声，按概率生成树木和矿产
func (rg *ResourceGenerator) GenerateResources(cx, cy int, blocks map[string]Block) map[string]Resource {
	resources := make(map[string]Resource)
	
	for key, block := range blocks {
		// 跳过水域
		if block.TerrainType == TerrainWater {
			continue
		}
		
		// 使用噪声生成资源概率
		// 结合区块坐标和区块内坐标，确保相同区块总是生成相同资源
		x := float64(cx*ChunkSize + block.X)
		y := float64(cy*ChunkSize + block.Y)
		
		switch block.TerrainType {
		case TerrainGrass, TerrainFarmland:
			// 树木生成概率
			if rg.shouldGenerateTree(x, y) {
				resources[key] = Resource{
					Type:   "tree",
					Amount: rg.rng.Intn(8) + 3, // 3-10
				}
			}
			
		case TerrainHill:
			// 矿产生成概率（山顶更可能）
			if rg.shouldGenerateOre(x, y, block.Height) {
				resources[key] = Resource{
					Type:   "ore",
					Amount: rg.rng.Intn(5) + 2, // 2-6
				}
			}
		}
	}
	
	return resources
}

// shouldGenerateTree 判断是否生成树木
// 基于噪声值和概率
func (rg *ResourceGenerator) shouldGenerateTree(x, y float64) bool {
	// 使用分形噪声生成连续的概率分布
	noise := rg.noise.FractalNoise(
		x*0.1,
		y*0.1,
		3,
		0.5,
		2.0,
	)
	
	// 噪声值 > 0.4 时生成树木，概率约 30%
	return noise > 0.4
}

// shouldGenerateOre 判断是否生成矿产
// 基于噪声值、高度和概率（越高越可能）
func (rg *ResourceGenerator) shouldGenerateOre(x, y float64, height int) bool {
	// 基础噪声
	noise := rg.noise.FractalNoise(
		x*0.08,
		y*0.08,
		3,
		0.5,
		2.0,
	)
	
	// 高度加成：每高一层，阈值降低 0.1
	// height=2: 阈值 0.6
	// height=3: 阈值 0.5
	threshold := 0.7 - float64(height-2)*0.1
	
	return noise > threshold
}

// GetResource 获取指定位置的资源
func (rg *ResourceGenerator) GetResource(
	cx, cy, bx, by int,
	terrainType string,
	height int,
) *Resource {
	x := float64(cx*ChunkSize + bx)
	y := float64(cy*ChunkSize + by)
	
	switch terrainType {
	case TerrainGrass, TerrainFarmland:
		if rg.shouldGenerateTree(x, y) {
			return &Resource{
				Type:   "tree",
				Amount: rg.rng.Intn(8) + 3,
			}
		}
	case TerrainHill:
		if rg.shouldGenerateOre(x, y, height) {
			return &Resource{
				Type:   "ore",
				Amount: rg.rng.Intn(5) + 2,
			}
		}
	}
	
	return nil
}
