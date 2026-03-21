package service

import (
	"errors"

	"github.com/RedContritio/agent-town/server/internal/model"
	"github.com/RedContritio/agent-town/server/internal/repository"
)

// InventoryService 背包服务
type InventoryService struct {
	invRepo *repository.InventoryRepository
}

// NewInventoryService 创建背包服务
func NewInventoryService(invRepo *repository.InventoryRepository) *InventoryService {
	return &InventoryService{invRepo: invRepo}
}

// GetInventory 获取背包内容
func (s *InventoryService) GetInventory(agentID int64) (*model.InventoryResponse, error) {
	items, err := s.invRepo.GetByAgentID(agentID)
	if err != nil {
		return nil, err
	}
	
	// 构建物品类型映射
	itemTypes := make(map[int]*model.ItemType)
	for _, it := range model.DefaultItemTypes {
		itemTypes[it.ID] = it
	}
	
	// 转换响应格式
	var itemResponses []map[string]interface{}
	for _, item := range items {
		itemResponses = append(itemResponses, item.ToAPIResponse(itemTypes))
	}
	
	return &model.InventoryResponse{
		AgentID:   agentID,
		Capacity:  repository.InventoryCapacity,
		Items:     itemResponses,
		UsedSlots: len(items),
	}, nil
}

// AddItem 添加物品到背包
func (s *InventoryService) AddItem(agentID int64, itemTypeName string, quantity int) (bool, int, error) {
	itemType := model.GetItemTypeByName(itemTypeName)
	if itemType == nil {
		return false, 0, errors.New("unknown item type: " + itemTypeName)
	}
	
	return s.invRepo.AddItem(agentID, itemType.ID, quantity)
}

// RemoveItem 从背包移除物品
func (s *InventoryService) RemoveItem(agentID int64, itemTypeName string, quantity int) (bool, error) {
	itemType := model.GetItemTypeByName(itemTypeName)
	if itemType == nil {
		return false, errors.New("unknown item type: " + itemTypeName)
	}
	
	return s.invRepo.RemoveItem(agentID, itemType.ID, quantity)
}

// HasItem 检查是否有足够物品
func (s *InventoryService) HasItem(agentID int64, itemTypeName string, quantity int) (bool, int, error) {
	itemType := model.GetItemTypeByName(itemTypeName)
	if itemType == nil {
		return false, 0, errors.New("unknown item type: " + itemTypeName)
	}
	
	return s.invRepo.HasItem(agentID, itemType.ID, quantity)
}

// AddHarvestResult 添加采集结果到背包
func (s *InventoryService) AddHarvestResult(agentID int64, resourceType string) (map[string]interface{}, error) {
	// 根据资源类型决定产出
	var itemName string
	var amount int
	
	switch resourceType {
	case "tree":
		itemName = "木材"
		amount = 5
	case "mine":
		itemName = "石块"
		amount = 3
	default:
		itemName = "木材"
		amount = 1
	}
	
	success, added, err := s.AddItem(agentID, itemName, amount)
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"success":    success,
		"item":       itemName,
		"amount":     added,
		"total":      amount,
		"backpack_full": !success && added < amount,
	}, nil
}


