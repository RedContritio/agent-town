package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/RedContritio/agent-town/server/internal/model"
	"github.com/RedContritio/agent-town/server/internal/repository"
)

// CraftHandler 制作处理器接口
// 使用 interface{} 返回类型避免与 service 包的类型循环依赖
type CraftHandler interface {
	CanCraft(agentID int64, recipe *model.Recipe) (bool, string)
	ExecuteCraftInterface(agentID int64, recipe *model.Recipe, count int) (interface{}, error)
	GetRecipe(name string) *model.Recipe
}

// Engine 任务执行引擎
type Engine struct {
	eventQueue    *EventQueue
	executor      *TaskExecutor
	agentRepo     *repository.AgentRepository
	taskRepo      *repository.TaskRepository
	invRepo       *repository.InventoryRepository
	craftHandler  CraftHandler
	
	running       bool
	ctx           context.Context
	cancel        context.CancelFunc
	
	// 统计信息
	stats         EngineStats
}

// EngineStats 引擎统计
type EngineStats struct {
	EventsProcessed int64
	TasksStarted    int64
	TasksCompleted  int64
	TasksFailed     int64
}

// NewEngine 创建引擎
func NewEngine(
	agentRepo *repository.AgentRepository,
	taskRepo *repository.TaskRepository,
	worldRepo *repository.WorldRepository,
	invRepo *repository.InventoryRepository,
) *Engine {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Engine{
		eventQueue:   NewEventQueue(),
		executor:     NewTaskExecutor(agentRepo, taskRepo, worldRepo),
		agentRepo:    agentRepo,
		taskRepo:     taskRepo,
		invRepo:      invRepo,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start 启动引擎（单线程事件循环）
func (e *Engine) Start() {
	if e.running {
		return
	}
	e.running = true
	
	log.Println("[Engine] Starting event loop...")
	
	// 启动定时 tick（每秒）
	go e.tickGenerator()
	
	// 主事件循环（单线程）
	go e.eventLoop()
	
	log.Println("[Engine] Event loop started")
}

// Stop 停止引擎
func (e *Engine) Stop() {
	if !e.running {
		return
	}
	e.running = false
	e.cancel()
	log.Println("[Engine] Stopped")
}

// IsRunning 检查是否运行中
func (e *Engine) IsRunning() bool {
	return e.running
}

// SetCraftHandler 设置制作处理器
func (e *Engine) SetCraftHandler(handler CraftHandler) {
	e.craftHandler = handler
}

// GetStats 获取统计
func (e *Engine) GetStats() EngineStats {
	return e.stats
}

// ScheduleTask 调度任务（外部调用入口）
func (e *Engine) ScheduleTask(agentID int64, taskSeq int) error {
	task, err := e.taskRepo.GetByAgentAndSeq(agentID, taskSeq)
	if err != nil {
		return fmt.Errorf("get task failed: %w", err)
	}
	if task == nil {
		return fmt.Errorf("task not found")
	}
	
	// 解析参数
	var params map[string]interface{}
	if task.Params != "" {
		if err := json.Unmarshal([]byte(task.Params), &params); err != nil {
			return fmt.Errorf("parse params failed: %w", err)
		}
	}
	
	// 创建任务开始事件（立即执行）
	event := &TaskStartEvent{
		BaseEvent: BaseEvent{
			Type:    EventTypeTaskStart,
			Time:    model.NowMillis(),
			AgentID: agentID,
			TaskSeq: taskSeq,
		},
		TaskType: task.Type,
		Params:   params,
	}
	
	e.eventQueue.Push(event)
	return nil
}

// tickGenerator 定时 tick 生成器
func (e *Engine) tickGenerator() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			if !e.running {
				return
			}
			
			event := &TickEvent{
				BaseEvent: BaseEvent{
					Type:    EventTypeTick,
					Time:    model.NowMillis(),
					AgentID: -1, // Tick 不属于特定 Agent
					TaskSeq: -1,
				},
			}
			e.eventQueue.Push(event)
		}
	}
}

// eventLoop 主事件循环（单线程）
func (e *Engine) eventLoop() {
	for {
		select {
		case <-e.ctx.Done():
			return
		default:
		}
		
		if !e.running {
			return
		}
		
		now := model.NowMillis()
		
		// 计算等待时间
		waitTime := e.eventQueue.WaitTime(now)
		
		// 如果有需要立即处理的事件，waitTime 为 0
		if waitTime > 0 {
			select {
			case <-e.ctx.Done():
				return
			case <-time.After(waitTime):
				// 继续处理
			case <-e.eventQueue.GetNotifier():
				// 有新事件，立即检查
			}
		}
		
		// 处理所有到期事件
		for {
			event := e.eventQueue.Peek()
			if event == nil {
				break
			}
			
			if event.GetTime() > model.NowMillis() {
				break // 下一个事件还未到时间
			}
			
			// 取出并处理事件
			e.eventQueue.Pop()
			e.handleEvent(event)
			e.stats.EventsProcessed++
		}
	}
}

// handleEvent 处理事件
func (e *Engine) handleEvent(event Event) {
	switch ev := event.(type) {
	case *TaskStartEvent:
		e.handleTaskStart(ev)
	case *TaskStepEvent:
		e.handleTaskStep(ev)
	case *TaskCompleteEvent:
		e.handleTaskComplete(ev)
	case *TaskFailEvent:
		e.handleTaskFail(ev)
	case *TickEvent:
		e.handleTick(ev)
	default:
		log.Printf("[Engine] Unknown event type: %T", event)
	}
}

// handleTaskStart 处理任务开始
func (e *Engine) handleTaskStart(ev *TaskStartEvent) {
	log.Printf("[Engine] Task start: agent=%d, task=%d, type=%d", 
		ev.AgentID, ev.TaskSeq, ev.TaskType)
	
	// 执行任务开始逻辑
	events, err := e.executor.ExecuteTaskStart(ev.AgentID, ev.TaskSeq, ev.TaskType, ev.Params)
	if err != nil {
		log.Printf("[Engine] Task start failed: %v", err)
		// 推送失败事件
		e.eventQueue.Push(&TaskFailEvent{
			BaseEvent: BaseEvent{
				Type:    EventTypeTaskFail,
				Time:    model.NowMillis(),
				AgentID: ev.AgentID,
				TaskSeq: ev.TaskSeq,
			},
			ErrorCode: 1,
			Reason:    err.Error(),
		})
		return
	}
	
	// 推送后续事件
	for _, evt := range events {
		e.eventQueue.Push(evt)
	}
	
	e.stats.TasksStarted++
}

// handleTaskStep 处理任务步骤
func (e *Engine) handleTaskStep(ev *TaskStepEvent) {
	log.Printf("[Engine] Task step: agent=%d, task=%d, step=%d/%d", 
		ev.AgentID, ev.TaskSeq, ev.Step, ev.TotalStep)
	
	// 这里可以执行具体的步骤逻辑（如移动一格）
	// 简化：直接执行下一步
	events, err := e.executor.ExecuteTaskStep(ev.AgentID, ev.TaskSeq, ev.Step, ev.TotalStep)
	if err != nil {
		log.Printf("[Engine] Task step failed: %v", err)
		return
	}
	
	for _, evt := range events {
		e.eventQueue.Push(evt)
	}
}

// handleTaskComplete 处理任务完成
func (e *Engine) handleTaskComplete(ev *TaskCompleteEvent) {
	log.Printf("[Engine] Task complete: agent=%d, task=%d", 
		ev.AgentID, ev.TaskSeq)
	
	// 先获取任务类型和参数，用于处理背包操作
	task, err := e.taskRepo.GetByAgentAndSeq(ev.AgentID, ev.TaskSeq)
	if err != nil {
		log.Printf("[Engine] Get task failed: %v", err)
		return
	}
	if task == nil {
		log.Printf("[Engine] Task not found: agent=%d, seq=%d", ev.AgentID, ev.TaskSeq)
		return
	}
	
	// 处理背包操作
	switch task.Type {
	case model.TaskTypeHarvest:
		e.processHarvestResult(ev)
	case model.TaskTypeCraft:
		e.processCraftResult(ev)
	}
	
	resultJSON, _ := json.Marshal(ev.Result)
	
	// 完成任务
	if err := e.taskRepo.CompleteTaskWithResult(ev.AgentID, ev.TaskSeq, string(resultJSON)); err != nil {
		log.Printf("[Engine] Complete task failed: %v", err)
		return
	}
	
	e.stats.TasksCompleted++
	
	// 触发栈恢复（如果这是栈顶任务）
	e.resumeNextTask(ev.AgentID)
}

// processHarvestResult 处理采集结果
func (e *Engine) processHarvestResult(ev *TaskCompleteEvent) {
	// 获取产出资源类型
	resourceID, _ := ev.Result["resource_id"].(string)
	
	// 根据资源类型决定产出
	var itemName string
	var amount int
	switch resourceID {
	case "tree-001", "tree":
		itemName = "木材"
		amount = 5
	case "mine-001", "mine":
		itemName = "石块"
		amount = 3
	default:
		itemName = "木材"
		amount = 2
	}
	
	// 添加到背包
	success, added, err := e.invRepo.AddItem(ev.AgentID, getItemTypeID(itemName), amount)
	if err != nil {
		log.Printf("[Engine] Add harvest result to inventory failed: %v", err)
		return
	}
	
	ev.Result["items_added"] = map[string]interface{}{
		"item":   itemName,
		"amount": added,
		"success": success,
	}
}

// processCraftResult 处理制作结果
func (e *Engine) processCraftResult(ev *TaskCompleteEvent) {
	itemName, _ := ev.Result["item"].(string)
	count, _ := ev.Result["count"].(int)
	if count == 0 {
		count = 1
	}
	
	// 使用 CraftHandler 执行制作
	if e.craftHandler != nil {
		recipe := e.craftHandler.GetRecipe(itemName)
		if recipe != nil {
			resp, err := e.craftHandler.ExecuteCraftInterface(ev.AgentID, recipe, count)
			if err != nil {
				log.Printf("[Engine] Craft execution failed: %v", err)
				return
			}
			
			// 使用反射获取响应字段
			if resp == nil {
				ev.Result["error"] = "craft response is nil"
				return
			}
			
			// 通过反射访问 CraftResponse 的字段
			t := reflect.TypeOf(resp)
			v := reflect.ValueOf(resp)
			
			// 如果是指针，解引用
			if t.Kind() == reflect.Ptr {
				v = v.Elem()
				t = t.Elem()
			}
			
			// 提取字段值
			result := make(map[string]interface{})
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				result[field.Name] = v.Field(i).Interface()
			}
			
			if success, ok := result["Success"].(bool); ok && success {
				ev.Result["items_added"] = map[string]interface{}{
					"item":      result["Recipe"],
					"amount":    result["Crafted"],
					"success":   true,
					"level_up":  result["LevelUp"],
					"new_level": result["NewLevel"],
				}
			} else {
				ev.Result["error"] = result["Message"]
			}
			return
		}
	}
	
	// 没有 CraftService 时，直接添加物品（降级处理）
	success, added, err := e.invRepo.AddItem(ev.AgentID, getItemTypeID(itemName), count)
	if err != nil {
		log.Printf("[Engine] Add craft result to inventory failed: %v", err)
		return
	}
	
	ev.Result["items_added"] = map[string]interface{}{
		"item":    itemName,
		"amount":  added,
		"success": success,
	}
}

// getItemTypeID 获取物品类型ID（简化版）
func getItemTypeID(name string) int {
	switch name {
	case "木材":
		return 1
	case "石块":
		return 2
	case "铁矿石":
		return 3
	case "小麦":
		return 4
	case "斧头":
		return 5
	case "镐":
		return 6
	case "木板":
		return 7
	case "铁锭":
		return 8
	default:
		return 1
	}
}

// handleTaskFail 处理任务失败
func (e *Engine) handleTaskFail(ev *TaskFailEvent) {
	log.Printf("[Engine] Task fail: agent=%d, task=%d, reason=%s", 
		ev.AgentID, ev.TaskSeq, ev.Reason)
	
	// 更新数据库
	if err := e.taskRepo.FailTaskWithReason(ev.AgentID, ev.TaskSeq, ev.ErrorCode, ev.Reason); err != nil {
		log.Printf("[Engine] Fail task failed: %v", err)
		return
	}
	
	e.stats.TasksFailed++
	
	// 触发栈恢复
	e.resumeNextTask(ev.AgentID)
}

// handleTick 处理定时 Tick
func (e *Engine) handleTick(ev *TickEvent) {
	// 可以在这里执行一些定期检查
	// 如：恢复体力、建筑损耗等
}

// resumeNextTask 恢复下一个任务（depth=1 改为 depth=0）
func (e *Engine) resumeNextTask(agentID int64) {
	// 这个逻辑在 task_service 中实现，这里通过数据库操作
	// 简化：发送一个特殊事件或直接调用 service
}
