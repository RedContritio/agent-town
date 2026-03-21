package repository

import (
	"database/sql"
	"errors"

	"github.com/RedContritio/agent-town/server/internal/model"
)

// SkillsRepository 技能数据访问
type SkillsRepository struct {
	db *sql.DB
}

// NewSkillsRepository 创建 SkillsRepository
func NewSkillsRepository(db *sql.DB) *SkillsRepository {
	return &SkillsRepository{db: db}
}

// GetByAgentID 获取 Agent 的所有技能
func (r *SkillsRepository) GetByAgentID(agentID int64) ([]*model.Skill, error) {
	rows, err := r.db.Query(
		"SELECT agent_id, skill_type, level, exp FROM skills WHERE agent_id=?",
		agentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []*model.Skill
	for rows.Next() {
		skill := &model.Skill{}
		err := rows.Scan(&skill.AgentID, &skill.SkillType, &skill.Level, &skill.Exp)
		if err != nil {
			return nil, err
		}
		skills = append(skills, skill)
	}

	return skills, rows.Err()
}

// GetSkill 获取特定技能
func (r *SkillsRepository) GetSkill(agentID int64, skillType int) (*model.Skill, error) {
	skill := &model.Skill{}
	err := r.db.QueryRow(
		"SELECT agent_id, skill_type, level, exp FROM skills WHERE agent_id=? AND skill_type=?",
		agentID, skillType,
	).Scan(&skill.AgentID, &skill.SkillType, &skill.Level, &skill.Exp)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// 不存在，返回默认技能（等级0）
			return &model.Skill{
				AgentID:   agentID,
				SkillType: skillType,
				Level:     0,
				Exp:       0,
			}, nil
		}
		return nil, err
	}
	return skill, nil
}

// CreateOrUpdate 创建或更新技能
func (r *SkillsRepository) CreateOrUpdate(skill *model.Skill) error {
	_, err := r.db.Exec(
		`INSERT INTO skills (agent_id, skill_type, level, exp) 
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(agent_id, skill_type) DO UPDATE SET
		 level=excluded.level, exp=excluded.exp`,
		skill.AgentID, skill.SkillType, skill.Level, skill.Exp,
	)
	return err
}

// AddExp 添加经验（原子操作）
func (r *SkillsRepository) AddExp(agentID int64, skillType, exp int) error {
	// 先获取当前技能
	skill, err := r.GetSkill(agentID, skillType)
	if err != nil {
		return err
	}

	// 添加经验
	leveledUp := skill.AddExp(exp)

	// 保存
	if err := r.CreateOrUpdate(skill); err != nil {
		return err
	}

	// 如果升级了，可以触发其他逻辑
	if leveledUp {
		// TODO: 发送升级通知
	}

	return nil
}

// InitDefaultSkills 初始化默认技能
func (r *SkillsRepository) InitDefaultSkills(agentID int64) error {
	// 为新 Agent 创建默认技能（等级 0）
	skills := []int{
		model.SkillTypeFarming,
		model.SkillTypeMining,
		model.SkillTypeBuilding,
		model.SkillTypeCrafting,
		model.SkillTypeCombat,
	}

	for _, skillType := range skills {
		skill := &model.Skill{
			AgentID:   agentID,
			SkillType: skillType,
			Level:     0,
			Exp:       0,
		}
		if err := r.CreateOrUpdate(skill); err != nil {
			return err
		}
	}

	return nil
}
