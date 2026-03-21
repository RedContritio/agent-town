// Package repository 数据访问层
package repository

import (
	"database/sql"
	"errors"

	"github.com/RedContritio/agent-town/server/internal/model"
)

// AgentRepository Agent 数据访问
type AgentRepository struct {
	db *sql.DB
}

// NewAgentRepository 创建 AgentRepository
func NewAgentRepository(db *sql.DB) *AgentRepository {
	return &AgentRepository{db: db}
}

// Create 创建 Agent
func (r *AgentRepository) Create(agent *model.Agent) error {
	result, err := r.db.Exec(
		`INSERT INTO agents (public_key, name, balance, position_x, position_y, facing_angle,
			hp, max_hp, stamina, max_stamina, hunger, max_hunger, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		agent.PublicKey, agent.Name, agent.Balance, agent.PositionX, agent.PositionY, agent.FacingAngle,
		agent.HP, agent.MaxHP, agent.Stamina, agent.MaxStamina, agent.Hunger, agent.MaxHunger,
		agent.CreatedAt, agent.UpdatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	agent.ID = id
	return nil
}

// GetByID 根据 ID 获取 Agent
func (r *AgentRepository) GetByID(id int64) (*model.Agent, error) {
	agent := &model.Agent{}
	err := r.db.QueryRow(
		`SELECT id, public_key, name, balance, position_x, position_y, facing_angle,
			hp, max_hp, stamina, max_stamina, hunger, max_hunger, created_at, updated_at
		 FROM agents WHERE id=?`,
		id,
	).Scan(
		&agent.ID, &agent.PublicKey, &agent.Name, &agent.Balance,
		&agent.PositionX, &agent.PositionY, &agent.FacingAngle,
		&agent.HP, &agent.MaxHP, &agent.Stamina, &agent.MaxStamina,
		&agent.Hunger, &agent.MaxHunger, &agent.CreatedAt, &agent.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return agent, nil
}

// GetByName 根据名称获取 Agent
func (r *AgentRepository) GetByName(name string) (*model.Agent, error) {
	agent := &model.Agent{}
	err := r.db.QueryRow(
		`SELECT id, public_key, name, balance, position_x, position_y, facing_angle,
			hp, max_hp, stamina, max_stamina, hunger, max_hunger, created_at, updated_at
		 FROM agents WHERE name=?`,
		name,
	).Scan(
		&agent.ID, &agent.PublicKey, &agent.Name, &agent.Balance,
		&agent.PositionX, &agent.PositionY, &agent.FacingAngle,
		&agent.HP, &agent.MaxHP, &agent.Stamina, &agent.MaxStamina,
		&agent.Hunger, &agent.MaxHunger, &agent.CreatedAt, &agent.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return agent, nil
}

// GetByPublicKey 根据公钥获取 Agent
func (r *AgentRepository) GetByPublicKey(publicKey []byte) (*model.Agent, error) {
	agent := &model.Agent{}
	err := r.db.QueryRow(
		`SELECT id, public_key, name, balance, position_x, position_y, facing_angle,
			hp, max_hp, stamina, max_stamina, hunger, max_hunger, created_at, updated_at
		 FROM agents WHERE public_key=?`,
		publicKey,
	).Scan(
		&agent.ID, &agent.PublicKey, &agent.Name, &agent.Balance,
		&agent.PositionX, &agent.PositionY, &agent.FacingAngle,
		&agent.HP, &agent.MaxHP, &agent.Stamina, &agent.MaxStamina,
		&agent.Hunger, &agent.MaxHunger, &agent.CreatedAt, &agent.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return agent, nil
}

// Update 更新 Agent
func (r *AgentRepository) Update(agent *model.Agent) error {
	agent.UpdatedAt = model.NowMillis()
	_, err := r.db.Exec(
		`UPDATE agents SET name=?, balance=?, position_x=?, position_y=?, facing_angle=?,
			hp=?, max_hp=?, stamina=?, max_stamina=?, hunger=?, max_hunger=?, updated_at=?
		 WHERE id=?`,
		agent.Name, agent.Balance, agent.PositionX, agent.PositionY, agent.FacingAngle,
		agent.HP, agent.MaxHP, agent.Stamina, agent.MaxStamina, agent.Hunger, agent.MaxHunger,
		agent.UpdatedAt, agent.ID,
	)
	return err
}

// Delete 删除 Agent
func (r *AgentRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM agents WHERE id=?", id)
	return err
}

// ExistsByName 检查名称是否已存在
func (r *AgentRepository) ExistsByName(name string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM agents WHERE name=?", name).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByPublicKey 检查公钥是否已存在
func (r *AgentRepository) ExistsByPublicKey(publicKey []byte) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM agents WHERE public_key=?", publicKey).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
