package world

import (
	"fmt"
	"math/rand"
	"time"
)

// 建筑类型常量
const (
	BuildingGovHall   = "gov_hall"   // 政府大厅
	BuildingGuideHall = "guide_hall" // 引导大厅
	BuildingQuest     = "quest"      // 委托处
	BuildingShop      = "shop"       // 商店
)

// BuildingColors 建筑颜色映射
var BuildingColors = map[string]string{
	BuildingGovHall:   "#8b4513", // 深棕色 - 政府权威
	BuildingGuideHall: "#4169e1", // 皇家蓝 - 引导信任
	BuildingQuest:     "#228b22", // 森林绿 - 委托任务
	BuildingShop:      "#ff8c00", // 深橙色 - 商店贸易
}

// Building 建筑
type Building struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	OwnerID string   `json:"ownerId"`
	Anchor  Position `json:"anchor"`
	Width   int      `json:"width"`
	Depth   int      `json:"depth"`
	Height  int      `json:"height"`
}

// Position 坐标
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
	Z int `json:"z"`
}

// BuildingGenerator 建筑生成器
type BuildingGenerator struct {
	seed int64
	rng  *rand.Rand
}

// NewBuildingGenerator 创建建筑生成器
func NewBuildingGenerator(seed int64) *BuildingGenerator {
	return &BuildingGenerator{
		seed: seed,
		rng:  rand.New(rand.NewSource(seed)),
	}
}

// GenerateInitialBuildings 生成初始四建筑
// 固定在初始区块的四个角落，2x2大小
func (bg *BuildingGenerator) GenerateInitialBuildings() []Building {
	buildings := []Building{
		{
			ID:      "building-gov-hall",
			Name:    "政府大厅",
			Type:    BuildingGovHall,
			OwnerID: "government",
			Anchor:  Position{X: -12, Y: -12, Z: 0},
			Width:   2,
			Depth:   2,
			Height:  3,
		},
		{
			ID:      "building-guide-hall",
			Name:    "引导大厅",
			Type:    BuildingGuideHall,
			OwnerID: "government",
			Anchor:  Position{X: 10, Y: -12, Z: 0},
			Width:   2,
			Depth:   2,
			Height:  3,
		},
		{
			ID:      "building-quest",
			Name:    "委托处",
			Type:    BuildingQuest,
			OwnerID: "government",
			Anchor:  Position{X: -12, Y: 10, Z: 0},
			Width:   2,
			Depth:   2,
			Height:  3,
		},
		{
			ID:      "building-shop",
			Name:    "商店",
			Type:    BuildingShop,
			OwnerID: "government",
			Anchor:  Position{X: 10, Y: 10, Z: 0},
			Width:   2,
			Depth:   2,
			Height:  3,
		},
	}

	return buildings
}

// CreateBuilding 运行时创建建筑（预留接口）
func (bg *BuildingGenerator) CreateBuilding(
	anchor Position,
	buildingType string,
	ownerID string,
	name string,
	width, depth, height int,
) (*Building, error) {
	building := &Building{
		ID:      bg.generateBuildingID(),
		Name:    name,
		Type:    buildingType,
		OwnerID: ownerID,
		Anchor:  anchor,
		Width:   width,
		Depth:   depth,
		Height:  height,
	}

	return building, nil
}

// DestroyBuilding 销毁建筑（预留接口）
func (bg *BuildingGenerator) DestroyBuilding(id string) error {
	// TODO: 实现建筑销毁逻辑
	return nil
}

// generateBuildingID 生成建筑 ID
func (bg *BuildingGenerator) generateBuildingID() string {
	return fmt.Sprintf("building-%d-%d", time.Now().UnixNano(), bg.rng.Intn(10000))
}

// GetBuildingColor 获取建筑颜色
func GetBuildingColor(buildingType string) string {
	if color, ok := BuildingColors[buildingType]; ok {
		return color
	}
	return "#a08060" // 默认颜色
}
