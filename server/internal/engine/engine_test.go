package engine

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/RedContritio/agent-town/server/internal/db"
	"github.com/RedContritio/agent-town/server/internal/model"
	"github.com/RedContritio/agent-town/server/internal/repository"
	_ "modernc.org/sqlite"
)

func setupTestEngine(t *testing.T) (*Engine, *repository.AgentRepository, *repository.TaskRepository, *repository.InventoryRepository, *repository.SkillsRepository) {
	database, err := db.Init(&db.Config{DataDir: ":memory:", DBName: ""})
	if err != nil {
		t.Fatalf("初始化数据库失败: %v", err)
	}

	agentRepo := repository.NewAgentRepository(database)
	taskRepo := repository.NewTaskRepository(database)
	worldRepo := repository.NewWorldRepository(database)
	invRepo := repository.NewInventoryRepository(database)
	skillsRepo := repository.NewSkillsRepository(database)
	
	engine := NewEngine(agentRepo, taskRepo, worldRepo, invRepo)
	
	// 注意：测试中不设置 craft handler 以避免导入 service 包
	// craft handler 的测试在 service 包中进行
	
	return engine, agentRepo, taskRepo, invRepo, skillsRepo
}

func createTestAgent(t *testing.T, agentRepo *repository.AgentRepository, name string) int64 {
	agent := model.NewAgent(name, []byte("test-key-"+name))
	if err := agentRepo.Create(agent); err != nil {
		t.Fatalf("创建 Agent 失败: %v", err)
	}
	return agent.ID
}

func TestEngine_StartStop(t *testing.T) {
	engine, _, _, _, _ := setupTestEngine(t)
	
	// 测试启动
	engine.Start()
	time.Sleep(100 * time.Millisecond)
	
	if !engine.IsRunning() {
		t.Error("引擎应该处于运行状态")
	}
	
	// 测试停止
	engine.Stop()
	time.Sleep(100 * time.Millisecond)
	
	if engine.IsRunning() {
		t.Error("引擎应该已停止")
	}
}

func TestEventQueue_Basic(t *testing.T) {
	queue := NewEventQueue()
	
	now := model.NowMillis()
	
	// 添加事件
	event1 := &BaseEvent{
		Type:    EventTypeTick,
		Time:    now + 1000,
		AgentID: 1,
	}
	event2 := &BaseEvent{
		Type:    EventTypeTick,
		Time:    now + 500,
		AgentID: 2,
	}
	
	queue.Push(event1)
	queue.Push(event2)
	
	// 检查顺序（event2 应该更早）
	peek := queue.Peek()
	if peek.GetAgentID() != 2 {
		t.Errorf("期望最早事件是 agent 2, 得到 %d", peek.GetAgentID())
	}
	
	// 弹出
	pop1 := queue.Pop()
	if pop1.GetAgentID() != 2 {
		t.Errorf("期望弹出 agent 2, 得到 %d", pop1.GetAgentID())
	}
	
	pop2 := queue.Pop()
	if pop2.GetAgentID() != 1 {
		t.Errorf("期望弹出 agent 1, 得到 %d", pop2.GetAgentID())
	}
}

func TestTaskExecutor_Move(t *testing.T) {
	engine, agentRepo, _, _, _ := setupTestEngine(t)
	
	agentID := createTestAgent(t, agentRepo, "MoveTest")
	
	// 创建移动任务
	params := map[string]interface{}{
		"dx": 3,
		"dy": 0,
	}
	
	events, err := engine.executor.executeMoveStart(agentID, 1, params, model.NowMillis())
	if err != nil {
		t.Fatalf("执行移动任务失败: %v", err)
	}
	
	// 应该有步骤事件
	if len(events) != 1 {
		t.Fatalf("期望 1 个事件，得到 %d", len(events))
	}
	
	stepEvent, ok := events[0].(*TaskStepEvent)
	if !ok {
		t.Fatalf("期望 TaskStepEvent, 得到 %T", events[0])
	}
	
	if stepEvent.TotalStep != 3 {
		t.Errorf("期望总步数 3, 得到 %d", stepEvent.TotalStep)
	}
}

func TestTaskExecutor_ZeroMove(t *testing.T) {
	engine, agentRepo, _, _, _ := setupTestEngine(t)
	
	agentID := createTestAgent(t, agentRepo, "ZeroMove")
	
	// 创建 0 距离移动任务
	params := map[string]interface{}{
		"dx": 0,
		"dy": 0,
	}
	
	events, err := engine.executor.executeMoveStart(agentID, 1, params, model.NowMillis())
	if err != nil {
		t.Fatalf("执行移动任务失败: %v", err)
	}
	
	// 应该立即完成
	if len(events) != 1 {
		t.Fatalf("期望 1 个事件，得到 %d", len(events))
	}
	
	_, ok := events[0].(*TaskCompleteEvent)
	if !ok {
		t.Fatalf("期望 TaskCompleteEvent, 得到 %T", events[0])
	}
}

func TestEngine_ScheduleTask(t *testing.T) {
	engine, agentRepo, taskRepo, _, _ := setupTestEngine(t)
	
	agentID := createTestAgent(t, agentRepo, "ScheduleTest")
	
	// 先创建一个任务到数据库
	params := map[string]interface{}{"dx": 1, "dy": 0}
	paramsJSON, _ := json.Marshal(params)
	now := model.NowMillis()
	
	// 手动插入任务
	_, err := taskRepo.DB.Exec(
		`INSERT INTO tasks (agent_id, seq, type, status, params, stack_depth, created_at, started_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		agentID, 1, model.TaskTypeMove, model.TaskStatusRunning, 
		string(paramsJSON), 0, now, now,
	)
	if err != nil {
		t.Fatalf("插入任务失败: %v", err)
	}
	
	// 启动引擎
	engine.Start()
	defer engine.Stop()
	
	// 调度任务
	err = engine.ScheduleTask(agentID, 1)
	if err != nil {
		t.Errorf("调度任务失败: %v", err)
	}
	
	// 等待事件处理
	time.Sleep(500 * time.Millisecond)
	
	// 检查统计
	stats := engine.GetStats()
	if stats.EventsProcessed == 0 {
		t.Error("应该已处理事件")
	}
}
