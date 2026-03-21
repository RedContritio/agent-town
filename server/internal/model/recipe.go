package model

// Recipe 制作配方
type Recipe struct {
	ID          int                `json:"id"`
	Name        string             `json:"name"`
	SkillType   int                `json:"skill_type"`   // 关联技能类型
	SkillLevel  int                `json:"skill_level"`  // 需要技能等级
	Materials   map[int]int        `json:"materials"`    // item_type_id -> quantity
	Result      RecipeResult       `json:"result"`
	TimeCost    int64              `json:"time_cost"`    // 制作耗时（毫秒）
	Description string             `json:"description"`
}

// RecipeResult 制作结果
type RecipeResult struct {
	ItemTypeID int `json:"item_type_id"`
	Quantity   int `json:"quantity"`
}

// SkillType 技能类型常量
const (
	SkillTypeFarming  = 0
	SkillTypeMining   = 1
	SkillTypeBuilding = 2
	SkillTypeCrafting = 3  // 工匠技能
	SkillTypeCombat   = 4
)

// SkillTypeName 技能类型名称
var SkillTypeName = map[int]string{
	SkillTypeFarming:  "farming",
	SkillTypeMining:   "mining",
	SkillTypeBuilding: "building",
	SkillTypeCrafting: "crafting",
	SkillTypeCombat:   "combat",
}

// DefaultRecipes 默认配方表
var DefaultRecipes = []*Recipe{
	// 工匠等级 1 - 基础配方
	{
		ID:         1,
		Name:       "木板",
		SkillType:  SkillTypeCrafting,
		SkillLevel: 1,
		Materials:  map[int]int{1: 1}, // 1 木材
		Result:     RecipeResult{ItemTypeID: 7, Quantity: 2}, // 2 木板
		TimeCost:   5000,
		Description: "将木材加工成木板",
	},
	{
		ID:         2,
		Name:       "木棍",
		SkillType:  SkillTypeCrafting,
		SkillLevel: 1,
		Materials:  map[int]int{1: 1}, // 1 木材
		Result:     RecipeResult{ItemTypeID: 9, Quantity: 4}, // 4 木棍
		TimeCost:   3000,
		Description: "将木材削成木棍",
	},
	
	// 工匠等级 2 - 工具配方
	{
		ID:         3,
		Name:       "木镐",
		SkillType:  SkillTypeCrafting,
		SkillLevel: 2,
		Materials:  map[int]int{1: 3, 9: 2}, // 3 木材 + 2 木棍
		Result:     RecipeResult{ItemTypeID: 5, Quantity: 1}, // 1 镐
		TimeCost:   10000,
		Description: "基础采矿工具",
	},
	{
		ID:         4,
		Name:       "木斧",
		SkillType:  SkillTypeCrafting,
		SkillLevel: 2,
		Materials:  map[int]int{1: 3, 9: 2}, // 3 木材 + 2 木棍
		Result:     RecipeResult{ItemTypeID: 6, Quantity: 1}, // 1 斧头
		TimeCost:   10000,
		Description: "基础伐木工具",
	},
	
	// 工匠等级 3 - 进阶工具
	{
		ID:         5,
		Name:       "石镐",
		SkillType:  SkillTypeCrafting,
		SkillLevel: 3,
		Materials:  map[int]int{2: 3, 9: 2}, // 3 石块 + 2 木棍
		Result:     RecipeResult{ItemTypeID: 5, Quantity: 1},
		TimeCost:   15000,
		Description: "进阶采矿工具",
	},
	{
		ID:         6,
		Name:       "石斧",
		SkillType:  SkillTypeCrafting,
		SkillLevel: 3,
		Materials:  map[int]int{2: 3, 9: 2}, // 3 石块 + 2 木棍
		Result:     RecipeResult{ItemTypeID: 6, Quantity: 1},
		TimeCost:   15000,
		Description: "进阶伐木工具",
	},
	
	// 工匠等级 5 - 高级工具
	{
		ID:         7,
		Name:       "铁镐",
		SkillType:  SkillTypeCrafting,
		SkillLevel: 5,
		Materials:  map[int]int{8: 3, 9: 2}, // 3 铁锭 + 2 木棍
		Result:     RecipeResult{ItemTypeID: 5, Quantity: 1},
		TimeCost:   30000,
		Description: "高级采矿工具",
	},
}

// GetRecipeByName 根据名称获取配方
func GetRecipeByName(name string) *Recipe {
	for _, recipe := range DefaultRecipes {
		if recipe.Name == name {
			return recipe
		}
	}
	return nil
}

// GetRecipesBySkill 获取某技能的所有配方
func GetRecipesBySkill(skillType int) []*Recipe {
	var recipes []*Recipe
	for _, recipe := range DefaultRecipes {
		if recipe.SkillType == skillType {
			recipes = append(recipes, recipe)
		}
	}
	return recipes
}

// GetUnlockableRecipes 获取某技能等级可解锁的配方
func GetUnlockableRecipes(skillType, skillLevel int) []*Recipe {
	var recipes []*Recipe
	for _, recipe := range DefaultRecipes {
		if recipe.SkillType == skillType && recipe.SkillLevel <= skillLevel {
			recipes = append(recipes, recipe)
		}
	}
	return recipes
}

// Skill 技能
type Skill struct {
	AgentID    int64 `json:"-"`
	SkillType  int   `json:"skill_type"`
	Level      int   `json:"level"`
	Exp        int   `json:"exp"`
}

// SkillResponse 技能响应
type SkillResponse struct {
	SkillType   string `json:"type"`
	Level       int    `json:"level"`
	Exp         int    `json:"exp"`
	ExpToNext   int    `json:"exp_to_next"`
}

// ToResponse 转换为响应
func (s *Skill) ToResponse() *SkillResponse {
	return &SkillResponse{
		SkillType: SkillTypeName[s.SkillType],
		Level:     s.Level,
		Exp:       s.Exp,
		ExpToNext: CalculateExpToNext(s.Level),
	}
}

// CalculateExpToNext 计算升级所需经验
func CalculateExpToNext(level int) int {
	// 每级需要 100 * 2^(level-1) 经验
	if level >= 10 {
		return 0 // 满级
	}
	return 100 * (1 << (level - 1))
}

// AddExp 添加经验，返回是否升级
func (s *Skill) AddExp(exp int) bool {
	s.Exp += exp
	expToNext := CalculateExpToNext(s.Level)
	
	if s.Exp >= expToNext && s.Level < 10 {
		s.Exp -= expToNext
		s.Level++
		return true
	}
	return false
}
