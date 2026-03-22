package world

import (
	"fmt"
)

// Road 道路
type Road struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// RoadGenerator 道路生成器
type RoadGenerator struct{}

// NewRoadGenerator 创建道路生成器
func NewRoadGenerator() *RoadGenerator {
	return &RoadGenerator{}
}

// GenerateInitialRoads 生成初始区块内的基础道路
// 连接四个建筑，形成路网
func (rg *RoadGenerator) GenerateInitialRoads() []Road {
	roads := make([]Road, 0)

	// 四个建筑位置（中心点）
	// 政府大厅 (-12, -12)，引导大厅 (10, -12)，委托处 (-12, 10)，商店 (10, 10)
	govHall := Point{X: -11, Y: -11}     // 政府大厅中心
	guideHall := Point{X: 11, Y: -11}    // 引导大厅中心
	questCenter := Point{X: -11, Y: 11}  // 委托处中心
	shop := Point{X: 11, Y: 11}          // 商店中心

	// 中心广场
	center := Point{X: 0, Y: 0}

	// 从中心向四个建筑延伸道路
	roads = append(roads, rg.generateRoad(center, govHall)...)
	roads = append(roads, rg.generateRoad(center, guideHall)...)
	roads = append(roads, rg.generateRoad(center, questCenter)...)
	roads = append(roads, rg.generateRoad(center, shop)...)

	// 建筑之间的环路（形成方形环路）
	roads = append(roads, rg.generateRoad(govHall, guideHall)...)    // 北边路
	roads = append(roads, rg.generateRoad(guideHall, shop)...)       // 东边路
	roads = append(roads, rg.generateRoad(shop, questCenter)...)     // 南路
	roads = append(roads, rg.generateRoad(questCenter, govHall)...)  // 西边路

	// 去重
	roads = rg.deduplicateRoads(roads)

	return roads
}

// Point 坐标点
type Point struct {
	X, Y int
}

// generateRoad 生成两点之间的直线路径（L形）
func (rg *RoadGenerator) generateRoad(from, to Point) []Road {
	roads := make([]Road, 0)

	// L形路径：先水平后垂直（选择较短的路径）
	if abs(from.X-to.X) > abs(from.Y-to.Y) {
		// 先水平后垂直
		for x := from.X; x != to.X; {
			roads = append(roads, Road{X: x, Y: from.Y})
			if x < to.X {
				x++
			} else {
				x--
			}
		}
		for y := from.Y; y != to.Y; {
			roads = append(roads, Road{X: to.X, Y: y})
			if y < to.Y {
				y++
			} else {
				y--
			}
		}
	} else {
		// 先垂直后水平
		for y := from.Y; y != to.Y; {
			roads = append(roads, Road{X: from.X, Y: y})
			if y < to.Y {
				y++
			} else {
				y--
			}
		}
		for x := from.X; x != to.X; {
			roads = append(roads, Road{X: x, Y: to.Y})
			if x < to.X {
				x++
			} else {
				x--
			}
		}
	}

	// 添加终点
	roads = append(roads, Road{X: to.X, Y: to.Y})

	return roads
}

// deduplicateRoads 去除重复的道路
func (rg *RoadGenerator) deduplicateRoads(roads []Road) []Road {
	seen := make(map[string]bool)
	result := make([]Road, 0)

	for _, road := range roads {
		key := fmt.Sprintf("%d,%d", road.X, road.Y)
		if !seen[key] {
			seen[key] = true
			result = append(result, road)
		}
	}

	return result
}

// abs 绝对值
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
