package service

import (
	"fmt"
	"math"

	"github.com/RedContritio/agent-town/server/internal/repository"
	"github.com/RedContritio/agent-town/server/internal/world"
)

const (
	// VisionRadius 视野半径（chunk 数）
	// 半径 2 表示 3x3 = 9 个 chunk
	VisionRadius = 2
	
	// ChunkSize 每个 chunk 的格子数
	ChunkSize = 32
)

// VisionService 视野服务
type VisionService struct {
	worldRepo     *repository.WorldRepository
	chunkManager  *world.ChunkManager
}

// NewVisionService 创建视野服务
func NewVisionService(worldRepo *repository.WorldRepository, chunkManager *world.ChunkManager) *VisionService {
	return &VisionService{
		worldRepo:    worldRepo,
		chunkManager: chunkManager,
	}
}

// LookRequest 观察请求
type LookRequest struct {
	AgentID   int64
	PosX      int
	PosY      int
}

// LookResponse 观察响应
type LookResponse struct {
	CenterX    int                      `json:"center_x"`    // 中心 chunk X
	CenterY    int                      `json:"center_y"`    // 中心 chunk Y
	Radius     int                      `json:"radius"`      // 视野半径
	Chunks     []ChunkView              `json:"chunks"`      // 可见 chunks
	Agents     []AgentBriefView         `json:"agents"`      // 视野内其他 Agent
	Resources  []ResourceView           `json:"resources"`   // 可见资源
}

// ChunkView Chunk 视图
type ChunkView struct {
	X      int             `json:"x"`
	Y      int             `json:"y"`
	Blocks []BlockBriefView `json:"blocks,omitempty"`
}

// BlockBriefView 方块简要视图
type BlockBriefView struct {
	X           int    `json:"x"`
	Y           int    `json:"y"`
	TerrainType int    `json:"terrain_type"`
	Height      int    `json:"height"`
}

// AgentBriefView Agent 简要视图
type AgentBriefView struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Distance int    `json:"distance"` // 距离（格子数）
}

// ResourceView 资源视图
type ResourceView struct {
	ID       int    `json:"id"`
	Type     int    `json:"type"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	State    int    `json:"state"`
	Amount   int    `json:"amount"`
}

// ScanResponse 扫描响应（当前 chunk 详细信息）
type ScanResponse struct {
	ChunkX    int              `json:"chunk_x"`
	ChunkY    int              `json:"chunk_y"`
	Blocks    []BlockFullView  `json:"blocks"`
	Resources []ResourceView   `json:"resources"`
}

// BlockFullView 方块完整视图
type BlockFullView struct {
	X            int    `json:"x"`
	Y            int    `json:"y"`
	Height       int    `json:"height"`
	TerrainType  int    `json:"terrain_type"`
	ResourceType *int   `json:"resource_type,omitempty"`
	ResourceAmount *int `json:"resource_amount,omitempty"`
}

// Look 观察周围（视野范围内）
func (s *VisionService) Look(req *LookRequest) (*LookResponse, error) {
	// 计算所在 chunk
	centerChunkX, centerChunkY := s.worldToChunk(req.PosX, req.PosY)
	
	resp := &LookResponse{
		CenterX: centerChunkX,
		CenterY: centerChunkY,
		Radius:  VisionRadius,
		Chunks:  make([]ChunkView, 0),
	}
	
	// 收集视野范围内的 chunks
	for dy := -VisionRadius; dy <= VisionRadius; dy++ {
		for dx := -VisionRadius; dx <= VisionRadius; dx++ {
			chunkX := centerChunkX + dx
			chunkY := centerChunkY + dy
			
			chunkView, err := s.getChunkView(chunkX, chunkY)
			if err != nil {
				continue // 跳过无法加载的 chunk
			}
			
			resp.Chunks = append(resp.Chunks, *chunkView)
		}
	}
	
	return resp, nil
}

// Scan 扫描脚下（当前 chunk 详细信息）
func (s *VisionService) Scan(agentID int64, posX, posY int) (*ScanResponse, error) {
	chunkX, chunkY := s.worldToChunk(posX, posY)
	
	resp := &ScanResponse{
		ChunkX:    chunkX,
		ChunkY:    chunkY,
		Blocks:    make([]BlockFullView, 0),
		Resources: make([]ResourceView, 0),
	}
	
	// 获取 chunk 内的所有方块
	chunk, err := s.chunkManager.GetChunk(chunkX, chunkY)
	if err != nil {
		return nil, fmt.Errorf("get chunk failed: %w", err)
	}
	
	// 转换方块数据
	for _, block := range chunk.Blocks {
		blockView := BlockFullView{
			X:           block.X,
			Y:           block.Y,
			Height:      block.Height,
			TerrainType: s.terrainTypeToInt(block.TerrainType),
		}
		
		// 添加资源信息
		if block.ResourceType != "" {
			resType := s.resourceTypeToInt(block.ResourceType)
			blockView.ResourceType = &resType
			blockView.ResourceAmount = &block.ResourceAmount
			
			// 同时添加到资源列表
			resp.Resources = append(resp.Resources, ResourceView{
				Type:   resType,
				X:      block.X,
				Y:      block.Y,
				State:  2, // mature
				Amount: block.ResourceAmount,
			})
		}
		
		resp.Blocks = append(resp.Blocks, blockView)
	}
	
	return resp, nil
}

// getChunkView 获取 chunk 简要视图
func (s *VisionService) getChunkView(chunkX, chunkY int) (*ChunkView, error) {
	chunk, err := s.chunkManager.GetChunk(chunkX, chunkY)
	if err != nil {
		return nil, err
	}
	
	view := &ChunkView{
		X:      chunkX,
		Y:      chunkY,
		Blocks: make([]BlockBriefView, 0, len(chunk.Blocks)),
	}
	
	// 只返回表面信息（简化传输）
	for _, block := range chunk.Blocks {
		view.Blocks = append(view.Blocks, BlockBriefView{
			X:           block.X,
			Y:           block.Y,
			TerrainType: s.terrainTypeToInt(block.TerrainType),
			Height:      block.Height,
		})
	}
	
	return view, nil
}

// worldToChunk 世界坐标转 chunk 坐标
func (s *VisionService) worldToChunk(x, y int) (int, int) {
	// 处理负数坐标
	chunkX := int(math.Floor(float64(x) / ChunkSize))
	chunkY := int(math.Floor(float64(y) / ChunkSize))
	return chunkX, chunkY
}

// terrainTypeToInt 地形类型转整数
func (s *VisionService) terrainTypeToInt(terrainType string) int {
	switch terrainType {
	case "grass":
		return 0
	case "road":
		return 1
	case "water":
		return 2
	case "farmland":
		return 3
	case "sand":
		return 4
	case "hill":
		return 5
	default:
		return 0
	}
}

// resourceTypeToInt 资源类型转整数
func (s *VisionService) resourceTypeToInt(resourceType string) int {
	switch resourceType {
	case "tree":
		return 0
	case "mine":
		return 1
	case "wheat":
		return 2
	case "corn":
		return 3
	default:
		return 0
	}
}

// CalculateDistance 计算两点距离
func CalculateDistance(x1, y1, x2, y2 int) int {
	dx := x2 - x1
	dy := y2 - y1
	return int(math.Sqrt(float64(dx*dx + dy*dy)))
}
