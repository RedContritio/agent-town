package repository

import (
	"database/sql"
)

// WorldRepository 世界数据访问
type WorldRepository struct {
	db *sql.DB
}

// NewWorldRepository 创建 WorldRepository
func NewWorldRepository(db *sql.DB) *WorldRepository {
	return &WorldRepository{db: db}
}

// GetBlock 获取地块信息
func (r *WorldRepository) GetBlock(x, y int) (*BlockInfo, error) {
	var info BlockInfo
	err := r.db.QueryRow(
		"SELECT x, y, terrain_type, height FROM land_tiles WHERE x=? AND y=?",
		x, y,
	).Scan(&info.X, &info.Y, &info.TerrainType, &info.Height)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &info, nil
}

// GetResourceAt 获取位置资源
func (r *WorldRepository) GetResourceAt(x, y int) (*ResourceInfo, error) {
	var info ResourceInfo
	var state sql.NullInt32
	var amount sql.NullInt32
	
	err := r.db.QueryRow(
		"SELECT id, x, y, resource_type, state, amount FROM resources WHERE x=? AND y=?",
		x, y,
	).Scan(&info.ID, &info.X, &info.Y, &info.ResourceType, &state, &amount)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	if state.Valid {
		info.State = int(state.Int32)
	}
	if amount.Valid {
		info.Amount = int(amount.Int32)
	}
	
	return &info, nil
}

// BlockInfo 地块信息
type BlockInfo struct {
	X           int
	Y           int
	TerrainType int
	Height      int
}

// ResourceInfo 资源信息
type ResourceInfo struct {
	ID           int
	X            int
	Y            int
	ResourceType int
	State        int
	Amount       int
}
