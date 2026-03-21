package engine

import (
	"log"
	"sync"

	"github.com/RedContritio/agent-town/server/internal/repository"
)

// Manager 引擎管理器（单例）
type Manager struct {
	engine *Engine
	mu     sync.RWMutex
}

var (
	managerInstance *Manager
	managerOnce     sync.Once
)

// GetManager 获取引擎管理器单例
func GetManager() *Manager {
	managerOnce.Do(func() {
		managerInstance = &Manager{}
	})
	return managerInstance
}

// Init 初始化引擎
func (m *Manager) Init(
	agentRepo *repository.AgentRepository,
	taskRepo *repository.TaskRepository,
	worldRepo *repository.WorldRepository,
	invRepo *repository.InventoryRepository,
) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.engine != nil {
		m.engine.Stop()
	}
	
	m.engine = NewEngine(agentRepo, taskRepo, worldRepo, invRepo)
	log.Println("[EngineManager] Engine initialized")
}

// Start 启动引擎
func (m *Manager) Start() {
	m.mu.RLock()
	engine := m.engine
	m.mu.RUnlock()
	
	if engine != nil {
		engine.Start()
	}
}

// Stop 停止引擎
func (m *Manager) Stop() {
	m.mu.RLock()
	engine := m.engine
	m.mu.RUnlock()
	
	if engine != nil {
		engine.Stop()
	}
}

// IsRunning 检查是否运行
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.engine == nil {
		return false
	}
	return m.engine.IsRunning()
}

// GetStats 获取统计
func (m *Manager) GetStats() EngineStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.engine == nil {
		return EngineStats{}
	}
	return m.engine.GetStats()
}

// SetCraftHandler 设置制作处理器
func (m *Manager) SetCraftHandler(handler CraftHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.engine != nil {
		m.engine.SetCraftHandler(handler)
	}
}

// ScheduleTask 调度任务
func (m *Manager) ScheduleTask(agentID int64, taskSeq int) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.engine == nil {
		return nil
	}
	return m.engine.ScheduleTask(agentID, taskSeq)
}

// GetEngine 获取引擎实例（用于测试）
func (m *Manager) GetEngine() *Engine {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.engine
}
