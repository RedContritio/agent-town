package model

// ItemType 物品类型
type ItemType struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Category int    `json:"category"` // 0=resource, 1=tool, 2=food, 3=material
	StackMax int    `json:"stack_max"`
}

// InventoryItem 背包物品
type InventoryItem struct {
	AgentID  int64 `json:"-"`
	Slot     int   `json:"slot"`
	ItemType int   `json:"item_type"`
	Quantity int   `json:"quantity"`
}

// ToAPIResponse 转换为API响应
func (i *InventoryItem) ToAPIResponse(itemTypes map[int]*ItemType) map[string]interface{} {
	resp := map[string]interface{}{
		"slot":     i.Slot,
		"quantity": i.Quantity,
	}
	
	if itemType, ok := itemTypes[i.ItemType]; ok {
		resp["item_id"] = i.ItemType
		resp["name"] = itemType.Name
		resp["category"] = itemType.Category
	} else {
		resp["item_id"] = i.ItemType
		resp["name"] = "unknown"
		resp["category"] = 0
	}
	
	return resp
}

// InventoryResponse 背包响应
type InventoryResponse struct {
	AgentID   int64                    `json:"agent_id"`
	Capacity  int                      `json:"capacity"`
	Items     []map[string]interface{} `json:"items"`
	UsedSlots int                      `json:"used_slots"`
}

// 物品类型常量
const (
	ItemCategoryResource = 0
	ItemCategoryTool     = 1
	ItemCategoryFood     = 2
	ItemCategoryMaterial = 3
)

// 预定义物品类型
var DefaultItemTypes = []*ItemType{
	{ID: 1, Name: "木材", Category: ItemCategoryResource, StackMax: 99},
	{ID: 2, Name: "石块", Category: ItemCategoryResource, StackMax: 99},
	{ID: 3, Name: "铁矿石", Category: ItemCategoryResource, StackMax: 99},
	{ID: 4, Name: "小麦", Category: ItemCategoryFood, StackMax: 99},
	{ID: 5, Name: "斧头", Category: ItemCategoryTool, StackMax: 1},
	{ID: 6, Name: "镐", Category: ItemCategoryTool, StackMax: 1},
	{ID: 7, Name: "木板", Category: ItemCategoryMaterial, StackMax: 99},
	{ID: 8, Name: "铁锭", Category: ItemCategoryMaterial, StackMax: 99},
}

// GetItemTypeName 获取物品类型名称
func GetItemTypeName(itemTypeID int) string {
	for _, itemType := range DefaultItemTypes {
		if itemType.ID == itemTypeID {
			return itemType.Name
		}
	}
	return "unknown"
}

// GetItemTypeByName 根据名称获取物品类型
func GetItemTypeByName(name string) *ItemType {
	for _, itemType := range DefaultItemTypes {
		if itemType.Name == name {
			return itemType
		}
	}
	return nil
}
