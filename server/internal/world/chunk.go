package world

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	ChunkSize   = 32  // 区块大小 32×32
	CacheLimit  = 100 // 内存中最多缓存的区块数
)

// Block 地块
type Block struct {
	X              int    `json:"x"`
	Y              int    `json:"y"`
	Z              int    `json:"z"`
	Height         int    `json:"height"`         // 地形高度
	TerrainType    string `json:"terrainType"`    // 地形类型
	ResourceType   string `json:"resourceType,omitempty"`   // 资源类型
	ResourceAmount int    `json:"resourceAmount,omitempty"` // 资源数量
}

// Chunk 区块
type Chunk struct {
	X           int               `json:"x"`
	Y           int               `json:"y"`
	Blocks      map[string]Block  `json:"blocks"`      // key: "bx,by"
	Generated   bool              `json:"generated"`
	LastAccess  time.Time         `json:"lastAccess"`
}

// ChunkManager 区块管理器
type ChunkManager struct {
	seed       int64
	cache      map[string]*Chunk
	generator  *WorldGenerator
	db         *sql.DB
	chunkSize  int
}

// NewChunkManager 创建区块管理器
func NewChunkManager(seed int64, generator *WorldGenerator, db *sql.DB) *ChunkManager {
	return &ChunkManager{
		seed:      seed,
		cache:     make(map[string]*Chunk),
		generator: generator,
		db:        db,
		chunkSize: ChunkSize,
	}
}

// GetChunkKey 获取区块缓存 key
func (cm *ChunkManager) GetChunkKey(cx, cy int) string {
	return fmt.Sprintf("%d,%d", cx, cy)
}

// WorldToChunk 世界坐标转区块坐标和区块内坐标
// 返回: 区块坐标 (cx, cy), 区块内坐标 (bx, by)
func (cm *ChunkManager) WorldToChunk(x, y int) (cx, cy, bx, by int) {
	// 处理负数坐标
	if x < 0 {
		cx = (x - cm.chunkSize + 1) / cm.chunkSize
	} else {
		cx = x / cm.chunkSize
	}
	if y < 0 {
		cy = (y - cm.chunkSize + 1) / cm.chunkSize
	} else {
		cy = y / cm.chunkSize
	}
	
	bx = x - cx*cm.chunkSize
	by = y - cy*cm.chunkSize
	
	return cx, cy, bx, by
}

// ChunkToWorld 区块内坐标转世界坐标
func (cm *ChunkManager) ChunkToWorld(cx, cy, bx, by int) (x, y int) {
	return cx*cm.chunkSize + bx, cy*cm.chunkSize + by
}

// GetChunk 获取区块（优先缓存，否则生成）
func (cm *ChunkManager) GetChunk(cx, cy int) (*Chunk, error) {
	key := cm.GetChunkKey(cx, cy)
	
	// 检查缓存
	if chunk, ok := cm.cache[key]; ok {
		chunk.LastAccess = time.Now()
		return chunk, nil
	}
	
	// 生成新区块
	chunk, err := cm.generator.GenerateChunk(cx, cy)
	if err != nil {
		return nil, err
	}
	
	// 加入缓存
	cm.cache[key] = chunk
	
	// 缓存淘汰
	cm.evictCache()
	
	return chunk, nil
}

// GetBlock 获取单个地块
func (cm *ChunkManager) GetBlock(x, y int) (*Block, error) {
	cx, cy, bx, by := cm.WorldToChunk(x, y)
	
	chunk, err := cm.GetChunk(cx, cy)
	if err != nil {
		return nil, err
	}
	
	key := fmt.Sprintf("%d,%d", bx, by)
	if block, ok := chunk.Blocks[key]; ok {
		return &block, nil
	}
	
	// Debug: log missing block
	fmt.Printf("[GetBlock] Missing block at world(%d,%d) -> chunk(%d,%d) local(%d,%d) key=%s\n",
		x, y, cx, cy, bx, by, key)
	
	// 如果区块已生成但地块不存在，返回默认草地
	return &Block{
		X:           x,
		Y:           y,
		Z:           0,
		Height:      0,
		TerrainType: "grass",
	}, nil
}

// GetBlocksInRadius 获取以 (x, y) 为中心，radius 为半径范围内的所有地块
func (cm *ChunkManager) GetBlocksInRadius(x, y, radius int) ([]Block, error) {
	blocks := make([]Block, 0)
	
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			bx, by := x+dx, y+dy
			block, err := cm.GetBlock(bx, by)
			if err != nil {
				continue
			}
			blocks = append(blocks, *block)
		}
	}
	
	return blocks, nil
}

// evictCache 缓存淘汰（LRU策略）
func (cm *ChunkManager) evictCache() {
	if len(cm.cache) <= CacheLimit {
		return
	}
	
	// 找到最久未访问的区块并删除
	var oldestKey string
	var oldestTime time.Time
	
	for key, chunk := range cm.cache {
		if oldestKey == "" || chunk.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = chunk.LastAccess
		}
	}
	
	if oldestKey != "" {
		delete(cm.cache, oldestKey)
	}
}
