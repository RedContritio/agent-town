// Package model 定义数据模型
package model

import (
	"time"
)

// Agent Agent 实体
type Agent struct {
	ID         int64  `json:"id"`
	PublicKey  []byte `json:"-"` // 不序列化到 JSON
	Name       string `json:"name"`
	Balance    int    `json:"balance"`
	PositionX  int    `json:"positionX"`
	PositionY  int    `json:"positionY"`
	FacingAngle int   `json:"facingAngle"`
	HP         int    `json:"hp"`
	MaxHP      int    `json:"maxHp"`
	Stamina    int    `json:"stamina"`
	MaxStamina int    `json:"maxStamina"`
	Hunger     int    `json:"hunger"`
	MaxHunger  int    `json:"maxHunger"`
	CreatedAt  int64  `json:"createdAt"`
	UpdatedAt  int64  `json:"updatedAt"`
}

// ToAPIResponse 转换为 API 响应格式
func (a *Agent) ToAPIResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":          a.ID,
		"name":        a.Name,
		"balance":     a.Balance,
		"position":    map[string]int{"x": a.PositionX, "y": a.PositionY},
		"facingAngle": a.FacingAngle,
		"hp":          a.HP,
		"maxHp":       a.MaxHP,
		"stamina":     a.Stamina,
		"maxStamina":  a.MaxStamina,
		"hunger":      a.Hunger,
		"maxHunger":   a.MaxHunger,
		"createdAt":   a.CreatedAt,
		"updatedAt":   a.UpdatedAt,
	}
}

// NowMillis 获取当前毫秒时间戳
func NowMillis() int64 {
	return time.Now().UnixMilli()
}

// NewAgent 创建新 Agent
func NewAgent(name string, publicKey []byte) *Agent {
	now := NowMillis()
	return &Agent{
		Name:        name,
		PublicKey:   publicKey,
		Balance:     0,
		PositionX:   0,
		PositionY:   0,
		FacingAngle: 0,
		HP:          100,
		MaxHP:       100,
		Stamina:     100,
		MaxStamina:  100,
		Hunger:      100,
		MaxHunger:   100,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
