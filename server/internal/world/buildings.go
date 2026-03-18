package world

import (
	"fmt"
	"math/rand"
	"time"
)

// 建筑类型常量
const (
	BuildingTownHall   = "town_hall"
	BuildingBank       = "bank"
	BuildingQuestBoard = "quest_board"
	BuildingShop       = "shop"
)

// BuildingColors 建筑颜色映射
var BuildingColors = map[string]string{
	BuildingTownHall:   "#8b7355", // 棕色
	BuildingBank:       "#4a90d9", // 冷蓝
	BuildingQuestBoard: "#50c878", // 清爽绿
	BuildingShop:       "#e8a87c", // 暖橙
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
// 固定在初始区块的四个角落
func (bg *BuildingGenerator) GenerateInitialBuildings() []Building {
	buildings := []Building{
		{
			ID:      "building-townhall",
			Name:    "Town Hall",
			Type:    BuildingTownHall,
			OwnerID: "government",
			Anchor:  Position{X: -15, Y: -15, Z: 0},
			Width:   4,
			Depth:   4,
			Height:  2,
		},
		{
			ID:      "building-bank",
			Name:    "Bank",
			Type:    BuildingBank,
			OwnerID: "government",
			Anchor:  Position{X: 15, Y: -15, Z: 0},
			Width:   3,
			Depth:   3,
			Height:  2,
		},
		{
			ID:      "building-quest",
			Name:    "Quest Board",
			Type:    BuildingQuestBoard,
			OwnerID: "government",
			Anchor:  Position{X: -15, Y: 15, Z: 0},
			Width:   3,
			Depth:   3,
			Height:  1,
		},
		{
			ID:      "building-shop",
			Name:    "General Shop",
			Type:    BuildingShop,
			OwnerID: "government",
			Anchor:  Position{X: 15, Y: 15, Z: 0},
			Width:   3,
			Depth:   3,
			Height:  1,
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
