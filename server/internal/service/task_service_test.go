package service

import (
	"testing"

	"github.com/RedContritio/agent-town/server/internal/db"
	"github.com/RedContritio/agent-town/server/internal/model"
	"github.com/RedContritio/agent-town/server/internal/repository"
	_ "modernc.org/sqlite"
)

func setupTaskTest(t *testing.T) (*TaskService, *repository.AgentRepository) {
	database, err := db.Init(&db.Config{DataDir: ":memory:", DBName: ""})
	if err != nil {
		t.Fatalf("初始化数据库失败: %v", err)
	}

	agentRepo := repository.NewAgentRepository(database)
	taskRepo := repository.NewTaskRepository(database)
	invRepo := repository.NewInventoryRepository(database)
	skillsRepo := repository.NewSkillsRepository(database)
	craftService := NewCraftService(invRepo, skillsRepo)
	taskService := NewTaskService(taskRepo, agentRepo, craftService)

	return taskService, agentRepo
}

func createTestAgent(t *testing.T, agentRepo *repository.AgentRepository, name string) int64 {
	agent := model.NewAgent(name, []byte("test-key-"+name))
	if err := agentRepo.Create(agent); err != nil {
		t.Fatalf("创建 Agent 失败: %v", err)
	}
	return agent.ID
}

func TestCreateTask(t *testing.T) {
	taskService, agentRepo := setupTaskTest(t)
	agentID := createTestAgent(t, agentRepo, "TaskTest")

	// 创建第一个任务
	req := &CreateTaskRequest{
		Type: model.TaskTypeMove,
		Params: map[string]interface{}{
			"dx": 3,
			"dy": 0,
		},
	}
	resp, err := taskService.CreateTask(agentID, req)
	if err != nil {
		t.Fatalf("创建任务失败: %v", err)
	}
	if resp.TaskID != "TaskTest-001" {
		t.Errorf("期望 TaskID=TaskTest-001, 得到 %s", resp.TaskID)
	}
	if resp.Status != model.TaskStatusRunning {
		t.Errorf("期望 Status=running, 得到 %d", resp.Status)
	}

	// 创建第二个任务（第一个应该被暂停）
	req2 := &CreateTaskRequest{
		Type: model.TaskTypeHarvest,
		Params: map[string]interface{}{
			"resource_id": "tree-001",
		},
	}
	resp2, err := taskService.CreateTask(agentID, req2)
	if err != nil {
		t.Fatalf("创建第二个任务失败: %v", err)
	}
	if resp2.TaskID != "TaskTest-002" {
		t.Errorf("期望 TaskID=TaskTest-002, 得到 %s", resp2.TaskID)
	}

	// 验证栈
	stack, err := taskService.GetStack(agentID)
	if err != nil {
		t.Fatalf("获取任务栈失败: %v", err)
	}
	if len(stack) != 2 {
		t.Fatalf("期望栈长度=2, 得到 %d", len(stack))
	}

	// 栈顶（depth=0）应该是第二个任务
	if stack[0].Seq != 2 {
		t.Errorf("期望栈顶 Seq=2, 得到 %d", stack[0].Seq)
	}
	if stack[0].StackDepth != 0 {
		t.Errorf("期望栈顶 Depth=0, 得到 %d", stack[0].StackDepth)
	}
	if stack[0].Status != model.TaskStatusRunning {
		t.Errorf("期望栈顶 Status=running, 得到 %d", stack[0].Status)
	}

	// 第一个任务应该被暂停（depth=1）
	if stack[1].Seq != 1 {
		t.Errorf("期望第二个任务 Seq=1, 得到 %d", stack[1].Seq)
	}
	if stack[1].StackDepth != 1 {
		t.Errorf("期望第二个任务 Depth=1, 得到 %d", stack[1].StackDepth)
	}
	if stack[1].Status != model.TaskStatusPaused {
		t.Errorf("期望第二个任务 Status=paused, 得到 %d", stack[1].Status)
	}
}

func TestCompleteTask(t *testing.T) {
	taskService, agentRepo := setupTaskTest(t)
	agentID := createTestAgent(t, agentRepo, "CompleteTest")

	// 创建两个任务
	taskService.CreateTask(agentID, &CreateTaskRequest{
		Type:   model.TaskTypeMove,
		Params: map[string]interface{}{"dx": 1},
	})
	taskService.CreateTask(agentID, &CreateTaskRequest{
		Type:   model.TaskTypeHarvest,
		Params: map[string]interface{}{"resource_id": "tree-001"},
	})

	// 完成栈顶任务（seq=2）
	err := taskService.CompleteTask(agentID, 2, `{"harvested": 5}`)
	if err != nil {
		t.Fatalf("完成任务失败: %v", err)
	}

	// 验证第一个任务已恢复
	stack, _ := taskService.GetStack(agentID)
	if len(stack) != 1 {
		t.Fatalf("期望栈长度=1, 得到 %d", len(stack))
	}
	if stack[0].Seq != 1 {
		t.Errorf("期望剩余任务 Seq=1, 得到 %d", stack[0].Seq)
	}
	if stack[0].StackDepth != 0 {
		t.Errorf("期望剩余任务 Depth=0, 得到 %d", stack[0].StackDepth)
	}
	if stack[0].Status != model.TaskStatusRunning {
		t.Errorf("期望剩余任务 Status=running, 得到 %d", stack[0].Status)
	}
}

func TestDropTask(t *testing.T) {
	taskService, agentRepo := setupTaskTest(t)
	agentID := createTestAgent(t, agentRepo, "DropTest")

	// 创建三个任务
	taskService.CreateTask(agentID, &CreateTaskRequest{
		Type:   model.TaskTypeMove,
		Params: map[string]interface{}{"dx": 1},
	})
	taskService.CreateTask(agentID, &CreateTaskRequest{
		Type:   model.TaskTypeHarvest,
		Params: map[string]interface{}{"resource_id": "tree-001"},
	})
	taskService.CreateTask(agentID, &CreateTaskRequest{
		Type:   model.TaskTypeCraft,
		Params: map[string]interface{}{"item": "axe"},
	})

	// 放弃栈顶任务（seq=3）
	err := taskService.DropTask(agentID, 3)
	if err != nil {
		t.Fatalf("放弃任务失败: %v", err)
	}

	// 验证栈
	stack, _ := taskService.GetStack(agentID)
	if len(stack) != 2 {
		t.Fatalf("期望栈长度=2, 得到 %d", len(stack))
	}

	// seq=2 应该现在在栈顶
	if stack[0].Seq != 2 {
		t.Errorf("期望栈顶 Seq=2, 得到 %d", stack[0].Seq)
	}
	if stack[0].StackDepth != 0 {
		t.Errorf("期望栈顶 Depth=0, 得到 %d", stack[0].StackDepth)
	}
}

func TestPauseAndResumeTask(t *testing.T) {
	taskService, agentRepo := setupTaskTest(t)
	agentID := createTestAgent(t, agentRepo, "PauseTest")

	// 创建任务
	taskService.CreateTask(agentID, &CreateTaskRequest{
		Type:   model.TaskTypeMove,
		Params: map[string]interface{}{"dx": 1},
	})

	// 暂停任务
	err := taskService.PauseTask(agentID, 1)
	if err != nil {
		t.Fatalf("暂停任务失败: %v", err)
	}

	task, _ := taskService.GetTask(agentID, 1)
	if task.Status != model.TaskStatusPaused {
		t.Errorf("期望 Status=paused, 得到 %d", task.Status)
	}

	// 恢复任务
	err = taskService.ResumeTask(agentID, 1)
	if err != nil {
		t.Fatalf("恢复任务失败: %v", err)
	}

	task, _ = taskService.GetTask(agentID, 1)
	if task.Status != model.TaskStatusRunning {
		t.Errorf("期望 Status=running, 得到 %d", task.Status)
	}
}

func TestTaskTypeName(t *testing.T) {
	taskService, _ := setupTaskTest(t)

	tests := []struct {
		typ      int
		expected string
	}{
		{model.TaskTypeMove, "move"},
		{model.TaskTypeHarvest, "harvest"},
		{model.TaskTypeCraft, "craft"},
		{model.TaskTypeBuild, "build"},
		{model.TaskTypeCombat, "combat"},
		{999, "unknown"},
	}

	for _, tt := range tests {
		name := taskService.GetTaskTypeName(tt.typ)
		if name != tt.expected {
			t.Errorf("GetTaskTypeName(%d) = %s, 期望 %s", tt.typ, name, tt.expected)
		}
	}
}

func TestTaskStatusName(t *testing.T) {
	taskService, _ := setupTaskTest(t)

	tests := []struct {
		status   int
		expected string
	}{
		{model.TaskStatusPending, "pending"},
		{model.TaskStatusRunning, "running"},
		{model.TaskStatusPaused, "paused"},
		{model.TaskStatusCompleted, "completed"},
		{model.TaskStatusFailed, "failed"},
		{999, "unknown"},
	}

	for _, tt := range tests {
		name := taskService.GetTaskStatusName(tt.status)
		if name != tt.expected {
			t.Errorf("GetTaskStatusName(%d) = %s, 期望 %s", tt.status, name, tt.expected)
		}
	}
}
