package service

import (
	"errors"

	"github.com/RedContritio/agent-town/server/internal/model"
	"github.com/RedContritio/agent-town/server/internal/repository"
)

// AgentService Agent 服务
type AgentService struct {
	agentRepo    *repository.AgentRepository
	skillsRepo  *repository.SkillsRepository
	invRepo     *repository.InventoryRepository
}

// NewAgentService 创建 Agent 服务
func NewAgentService(
	agentRepo *repository.AgentRepository,
	skillsRepo *repository.SkillsRepository,
	invRepo *repository.InventoryRepository,
) *AgentService {
	return &AgentService{
		agentRepo:   agentRepo,
		skillsRepo:  skillsRepo,
		invRepo:     invRepo,
	}
}

// GetAgent 获取 Agent 信息
func (s *AgentService) GetAgent(agentID int64) (*model.Agent, error) {
	return s.agentRepo.GetByID(agentID)
}

// GetAgentStatus 获取 Agent 状态（包含技能等）
func (s *AgentService) GetAgentStatus(agentID int64) (map[string]interface{}, error) {
	agent, err := s.agentRepo.GetByID(agentID)
	if err != nil {
		return nil, err
	}
	if agent == nil {
		return nil, errors.New("agent not found")
	}

	return agent.ToAPIResponse(), nil
}

// UpdatePosition 更新位置
func (s *AgentService) UpdatePosition(agentID int64, x, y, facing int) error {
	agent, err := s.agentRepo.GetByID(agentID)
	if err != nil {
		return err
	}
	if agent == nil {
		return errors.New("agent not found")
	}

	agent.PositionX = x
	agent.PositionY = y
	agent.FacingAngle = facing
	agent.UpdatedAt = model.NowMillis()

	return s.agentRepo.Update(agent)
}

// UpdateHP 更新生命值
func (s *AgentService) UpdateHP(agentID int64, delta int) error {
	agent, err := s.agentRepo.GetByID(agentID)
	if err != nil {
		return err
	}
	if agent == nil {
		return errors.New("agent not found")
	}

	agent.HP += delta
	if agent.HP > agent.MaxHP {
		agent.HP = agent.MaxHP
	}
	if agent.HP < 0 {
		agent.HP = 0
	}
	agent.UpdatedAt = model.NowMillis()

	return s.agentRepo.Update(agent)
}

// UpdateBalance 更新余额
func (s *AgentService) UpdateBalance(agentID int64, delta int) error {
	agent, err := s.agentRepo.GetByID(agentID)
	if err != nil {
		return err
	}
	if agent == nil {
		return errors.New("agent not found")
	}

	agent.Balance += delta
	if agent.Balance < 0 {
		return errors.New("insufficient balance")
	}
	agent.UpdatedAt = model.NowMillis()

	return s.agentRepo.Update(agent)
}

// AgentResponse Agent 响应结构
type AgentResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Balance     int    `json:"balance"`
	Position    Position `json:"position"`
	FacingAngle int    `json:"facingAngle"`
	HP          int    `json:"hp"`
	MaxHP       int    `json:"maxHp"`
	Stamina     int    `json:"stamina"`
	MaxStamina  int    `json:"maxStamina"`
	Hunger      int    `json:"hunger"`
	MaxHunger   int    `json:"maxHunger"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}

// Position 位置结构
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// ToResponse 转换为响应
func (s *AgentService) ToResponse(agent *model.Agent) *AgentResponse {
	return &AgentResponse{
		ID:          agent.ID,
		Name:        agent.Name,
		Balance:     agent.Balance,
		Position:    Position{X: agent.PositionX, Y: agent.PositionY},
		FacingAngle: agent.FacingAngle,
		HP:          agent.HP,
		MaxHP:       agent.MaxHP,
		Stamina:     agent.Stamina,
		MaxStamina:  agent.MaxStamina,
		Hunger:      agent.Hunger,
		MaxHunger:   agent.MaxHunger,
		CreatedAt:   agent.CreatedAt,
		UpdatedAt:   agent.UpdatedAt,
	}
}
