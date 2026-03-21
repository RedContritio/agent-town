package repository

import (
	"database/sql"
	"errors"

	"github.com/RedContritio/agent-town/server/internal/model"
)

const InventoryCapacity = 40 // 背包容量

// InventoryRepository 背包数据访问
type InventoryRepository struct {
	db *sql.DB
}

// NewInventoryRepository 创建 InventoryRepository
func NewInventoryRepository(db *sql.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// GetByAgentID 获取Agent的背包
func (r *InventoryRepository) GetByAgentID(agentID int64) ([]*model.InventoryItem, error) {
	rows, err := r.db.Query(
		"SELECT agent_id, slot, item_type, quantity FROM inventory WHERE agent_id=? ORDER BY slot",
		agentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var items []*model.InventoryItem
	for rows.Next() {
		item := &model.InventoryItem{}
		err := rows.Scan(&item.AgentID, &item.Slot, &item.ItemType, &item.Quantity)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	
	return items, rows.Err()
}

// GetItemAt 获取指定格子的物品
func (r *InventoryRepository) GetItemAt(agentID int64, slot int) (*model.InventoryItem, error) {
	item := &model.InventoryItem{}
	err := r.db.QueryRow(
		"SELECT agent_id, slot, item_type, quantity FROM inventory WHERE agent_id=? AND slot=?",
		agentID, slot,
	).Scan(&item.AgentID, &item.Slot, &item.ItemType, &item.Quantity)
	
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return item, nil
}

// AddItem 添加物品到背包
// 返回: 是否成功, 实际添加数量, 错误
func (r *InventoryRepository) AddItem(agentID int64, itemType, quantity int) (bool, int, error) {
	// 查找已有堆叠或空槽
	for slot := 0; slot < InventoryCapacity; slot++ {
		existing, err := r.GetItemAt(agentID, slot)
		if err != nil {
			return false, 0, err
		}
		
		if existing == nil {
			// 空槽，直接放入
			_, err = r.db.Exec(
				"INSERT INTO inventory (agent_id, slot, item_type, quantity) VALUES (?, ?, ?, ?)",
				agentID, slot, itemType, quantity,
			)
			if err != nil {
				return false, 0, err
			}
			return true, quantity, nil
		}
		
		// 同类型且未满，堆叠
		if existing.ItemType == itemType && existing.Quantity < 99 {
			addable := 99 - existing.Quantity
			if addable > quantity {
				addable = quantity
			}
			
			_, err = r.db.Exec(
				"UPDATE inventory SET quantity=? WHERE agent_id=? AND slot=?",
				existing.Quantity+addable, agentID, slot,
			)
			if err != nil {
				return false, 0, err
			}
			
			quantity -= addable
			if quantity == 0 {
				return true, quantity, nil
			}
			// 还有剩余，继续找下一个槽
		}
	}
	
	// 背包已满，返回部分添加
	return false, quantity, nil
}

// RemoveItem 从背包移除物品
// 返回: 是否成功, 错误
func (r *InventoryRepository) RemoveItem(agentID int64, itemType, quantity int) (bool, error) {
	// 先查找所有该类型的物品
	rows, err := r.db.Query(
		"SELECT slot, quantity FROM inventory WHERE agent_id=? AND item_type=? ORDER BY slot",
		agentID, itemType,
	)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	
	type slotQty struct {
		slot     int
		quantity int
	}
	var slots []slotQty
	total := 0
	
	for rows.Next() {
		var sq slotQty
		if err := rows.Scan(&sq.slot, &sq.quantity); err != nil {
			return false, err
		}
		slots = append(slots, sq)
		total += sq.quantity
	}
	
	if total < quantity {
		return false, nil // 数量不足
	}
	
	// 开始移除
	for _, sq := range slots {
		if sq.quantity <= quantity {
			// 移除整个格子
			_, err = r.db.Exec(
				"DELETE FROM inventory WHERE agent_id=? AND slot=?",
				agentID, sq.slot,
			)
			if err != nil {
				return false, err
			}
			quantity -= sq.quantity
		} else {
			// 部分移除
			_, err = r.db.Exec(
				"UPDATE inventory SET quantity=? WHERE agent_id=? AND slot=?",
				sq.quantity-quantity, agentID, sq.slot,
			)
			if err != nil {
				return false, err
			}
			quantity = 0
		}
		
		if quantity == 0 {
			break
		}
	}
	
	return true, nil
}

// HasItem 检查是否有足够物品
func (r *InventoryRepository) HasItem(agentID int64, itemType, quantity int) (bool, int, error) {
	var total int
	err := r.db.QueryRow(
		"SELECT COALESCE(SUM(quantity), 0) FROM inventory WHERE agent_id=? AND item_type=?",
		agentID, itemType,
	).Scan(&total)
	
	if err != nil {
		return false, 0, err
	}
	
	return total >= quantity, total, nil
}

// Clear 清空背包（用于测试）
func (r *InventoryRepository) Clear(agentID int64) error {
	_, err := r.db.Exec("DELETE FROM inventory WHERE agent_id=?", agentID)
	return err
}
