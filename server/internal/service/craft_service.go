package service

import (
	"errors"
	"fmt"

	"github.com/RedContritio/agent-town/server/internal/model"
	"github.com/RedContritio/agent-town/server/internal/repository"
)

// CraftService 制作服务
type CraftService struct {
	invRepo   *repository.InventoryRepository
	skillsRepo *repository.SkillsRepository
}

// NewCraftService 创建制作服务
func NewCraftService(invRepo *repository.InventoryRepository, skillsRepo *repository.SkillsRepository) *CraftService {
	return &CraftService{
		invRepo:    invRepo,
		skillsRepo: skillsRepo,
	}
}

// CraftRequest 制作请求
type CraftRequest struct {
	RecipeName string `json:"recipe"`
	Count      int    `json:"count"`
}

// CraftResponse 制作响应
type CraftResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message,omitempty"`
	Recipe      string `json:"recipe,omitempty"`
	Crafted     int    `json:"crafted"`
	NewSkillExp int    `json:"new_skill_exp,omitempty"`
	LevelUp     bool   `json:"level_up,omitempty"`
	NewLevel    int    `json:"new_level,omitempty"`
}

// CanCraft 检查是否可以制作
func (s *CraftService) CanCraft(agentID int64, recipe *model.Recipe) (bool, string) {
	// 检查技能等级
	skill, err := s.skillsRepo.GetSkill(agentID, model.SkillTypeCrafting)
	if err != nil {
		return false, "无法获取技能信息"
	}

	if skill.Level < recipe.SkillLevel {
		return false, fmt.Sprintf("需要工匠等级 %d，当前等级 %d", recipe.SkillLevel, skill.Level)
	}

	// 检查材料
	for itemTypeID, need := range recipe.Materials {
		has, total, err := s.invRepo.HasItem(agentID, itemTypeID, need)
		if err != nil {
			return false, "检查材料失败"
		}
		if !has {
			itemName := model.GetItemTypeName(itemTypeID)
			return false, fmt.Sprintf("材料不足: %s 需要 %d，拥有 %d", itemName, need, total)
		}
	}

	return true, ""
}

// ExecuteCraft 执行制作
func (s *CraftService) ExecuteCraft(agentID int64, recipe *model.Recipe, count int) (*CraftResponse, error) {
	if count <= 0 {
		count = 1
	}

	resp := &CraftResponse{
		Recipe:  recipe.Name,
		Crafted: 0,
	}

	// 检查并消耗材料
	for itemTypeID, need := range recipe.Materials {
		totalNeed := need * count
		success, err := s.invRepo.RemoveItem(agentID, itemTypeID, totalNeed)
		if err != nil {
			return nil, err
		}
		if !success {
			itemName := model.GetItemTypeName(itemTypeID)
			resp.Success = false
			resp.Message = fmt.Sprintf("无法扣除材料: %s", itemName)
			return resp, nil
		}
	}

	// 添加产物
	totalResult := recipe.Result.Quantity * count
	success, added, err := s.invRepo.AddItem(agentID, recipe.Result.ItemTypeID, totalResult)
	if err != nil {
		return nil, err
	}

	if !success {
		// 背包满了，回滚材料（简化处理：暂不实现回滚，直接返回）
		resp.Success = false
		resp.Message = "背包已满"
		return resp, nil
	}

	resp.Success = true
	resp.Crafted = added

	// 添加技能经验（每个产物 10 点）
	expGain := 10 * added
	if err := s.skillsRepo.AddExp(agentID, model.SkillTypeCrafting, expGain); err != nil {
		// 技能更新失败不影响制作结果
		fmt.Printf("Add crafting exp failed: %v\n", err)
	}

	// 获取更新后的技能
	skill, _ := s.skillsRepo.GetSkill(agentID, model.SkillTypeCrafting)
	resp.NewSkillExp = skill.Exp
	resp.NewLevel = skill.Level

	// 检查是否升级（通过经验变化判断）
	// 简化：如果经验小于 10 * added，说明升级了（经验被重置）
	if skill.Exp < expGain && skill.Level > 0 {
		resp.LevelUp = true
	}

	return resp, nil
}

// GetAvailableRecipes 获取可用的配方列表
func (s *CraftService) GetAvailableRecipes(agentID int64) ([]*model.Recipe, error) {
	// 获取工匠技能等级
	skill, err := s.skillsRepo.GetSkill(agentID, model.SkillTypeCrafting)
	if err != nil {
		return nil, err
	}

	// 获取所有可解锁的配方
	recipes := model.GetUnlockableRecipes(model.SkillTypeCrafting, skill.Level)
	return recipes, nil
}

// GetRecipe 获取配方
func (s *CraftService) GetRecipe(name string) *model.Recipe {
	return model.GetRecipeByName(name)
}

// ValidateRecipe 验证配方名称
func (s *CraftService) ValidateRecipe(name string) (*model.Recipe, error) {
	recipe := model.GetRecipeByName(name)
	if recipe == nil {
		return nil, errors.New("未知配方: " + name)
	}
	return recipe, nil
}

// ExecuteCraftInterface 执行制作（返回 interface{} 以兼容 engine.CraftHandler）
func (s *CraftService) ExecuteCraftInterface(agentID int64, recipe *model.Recipe, count int) (interface{}, error) {
	return s.ExecuteCraft(agentID, recipe, count)
}
