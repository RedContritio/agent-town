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
// 仅连接四个建筑，不向外延伸
func (rg *RoadGenerator) GenerateInitialRoads() []Road {
	roads := make([]Road, 0)
	
	// 四个建筑位置
	townHall := Point{X: -15, Y: -15}
	bank := Point{X: 15, Y: -15}
	questBoard := Point{X: -15, Y: 15}
	shop := Point{X: 15, Y: 15}
	
	// 生成中心十字道路
	center := Point{X: 0, Y: 0}
	
	// 从中心向四个方向延伸到建筑
	roads = append(roads, rg.generateRoad(center, townHall)...)
	roads = append(roads, rg.generateRoad(center, bank)...)
	roads = append(roads, rg.generateRoad(center, questBoard)...)
	roads = append(roads, rg.generateRoad(center, shop)...)
	
	// 添加环绕道路（连接四个建筑形成环路）
	roads = append(roads, rg.generateRoad(townHall, bank)...)
	roads = append(roads, rg.generateRoad(bank, shop)...)
	roads = append(roads, rg.generateRoad(shop, questBoard)...)
	roads = append(roads, rg.generateRoad(questBoard, townHall)...)
	
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
