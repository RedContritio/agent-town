package world

import (
	"fmt"
	"math/rand"
	"time"
)

// Agent Agent 实体
type Agent struct {
	ID         string   `json:"id"`
	PublicKey  string   `json:"publicKey"`
	Name       string   `json:"name"`
	Position   Position `json:"position"`
	Facing     int      `json:"facing"`
	HP         int      `json:"hp"`
	MaxHP      int      `json:"maxHp"`
	Stamina    int      `json:"stamina"`
	MaxStamina int      `json:"maxStamina"`
	Hunger     int      `json:"hunger"`
	MaxHunger  int      `json:"maxHunger"`
	Balance    int      `json:"balance"`
	IsOnline   bool     `json:"isOnline"`
	InBattle   bool     `json:"inBattle"`
	Skills     []Skill  `json:"skills,omitempty"`
}

// Skill 技能
type Skill struct {
	Type      string `json:"type"`
	Level     int    `json:"level"`
	Exp       int    `json:"exp"`
	ExpToNext int    `json:"expToNext"`
}

// AgentGenerator Agent 生成器
type AgentGenerator struct {
	seed  int64
	count int
	rng   *rand.Rand
}

// NewAgentGenerator 创建 Agent 生成器
func NewAgentGenerator(seed int64) *AgentGenerator {
	return &AgentGenerator{
		seed:  seed,
		count: 0,
		rng:   rand.New(rand.NewSource(seed)),
	}
}

// GenerateDebugNPCs 生成调试用的初始 NPC
// 命名：debug-npc-01, debug-npc-02, ...
func (ag *AgentGenerator) GenerateDebugNPCs(spawnPoints []Position, count int) []Agent {
	agents := make([]Agent, count)
	
	for i := 0; i < count; i++ {
		ag.count++
		
		// 选择出生点
		spawnPos := spawnPoints[i%len(spawnPoints)]
		// 在出生点附近随机偏移
		spawnPos.X += ag.rng.Intn(5) - 2
		spawnPos.Y += ag.rng.Intn(5) - 2
		
		agents[i] = Agent{
			ID:         fmt.Sprintf("agent-debug-%02d", ag.count),
			PublicKey:  fmt.Sprintf("debug-pk-%02d", ag.count),
			Name:       fmt.Sprintf("debug-npc-%02d", ag.count),
			Position:   spawnPos,
			Facing:     ag.rng.Intn(4),
			HP:         100,
			MaxHP:      100,
			Stamina:    ag.rng.Intn(30) + 70, // 70-100
			MaxStamina: 100,
			Hunger:     ag.rng.Intn(20) + 80, // 80-100
			MaxHunger:  100,
			Balance:    ag.rng.Intn(500) + 100, // 100-600
			IsOnline:   true,
			InBattle:   false,
			Skills:     ag.generateRandomSkills(),
		}
	}
	
	return agents
}

// RegisterAgent 运行时注册新 Agent（预留接口）
func (ag *AgentGenerator) RegisterAgent(
	publicKey string,
	name string,
	spawnPoint Position,
) (*Agent, error) {
	ag.count++
	
	agent := &Agent{
		ID:         fmt.Sprintf("agent-%d", time.Now().UnixNano()),
		PublicKey:  publicKey,
		Name:       name,
		Position:   spawnPoint,
		Facing:     ag.rng.Intn(4),
		HP:         100,
		MaxHP:      100,
		Stamina:    100,
		MaxStamina: 100,
		Hunger:     100,
		MaxHunger:  100,
		Balance:    0,
		IsOnline:   true,
		InBattle:   false,
		Skills:     []Skill{},
	}
	
	return agent, nil
}

// RespawnAgent Agent 死亡后重生（预留接口）
func (ag *AgentGenerator) RespawnAgent(agentID string, spawnPoints []Position) (*Agent, error) {
	// TODO: 实现重生逻辑
	return nil, nil
}

// generateRandomSkills 生成随机技能
func (ag *AgentGenerator) generateRandomSkills() []Skill {
	skillTypes := []string{"farming", "mining", "building", "crafting", "combat"}
	skills := make([]Skill, 0)
	
	// 随机给 1-2 个技能 1 级
	numSkills := ag.rng.Intn(2) + 1
	for i := 0; i < numSkills; i++ {
		skillType := skillTypes[ag.rng.Intn(len(skillTypes))]
		
		// 检查是否已存在
		exists := false
		for _, s := range skills {
			if s.Type == skillType {
				exists = true
				break
			}
		}
		
		if !exists {
			skills = append(skills, Skill{
				Type:      skillType,
				Level:     1,
				Exp:       0,
				ExpToNext: 100,
			})
		}
	}
	
	return skills
}
