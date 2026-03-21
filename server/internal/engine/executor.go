package engine

import (
	"fmt"
	"math"

	"github.com/RedContritio/agent-town/server/internal/model"
	"github.com/RedContritio/agent-town/server/internal/repository"
)

// TaskExecutor 任务执行器
type TaskExecutor struct {
	agentRepo *repository.AgentRepository
	taskRepo  *repository.TaskRepository
	worldRepo *repository.WorldRepository
}

// NewTaskExecutor 创建任务执行器
func NewTaskExecutor(
	agentRepo *repository.AgentRepository,
	taskRepo *repository.TaskRepository,
	worldRepo *repository.WorldRepository,
) *TaskExecutor {
	return &TaskExecutor{
		agentRepo: agentRepo,
		taskRepo:  taskRepo,
		worldRepo: worldRepo,
	}
}

// ExecuteTaskStart 执行任务开始
func (e *TaskExecutor) ExecuteTaskStart(agentID int64, taskSeq int, taskType int, params map[string]interface{}) ([]Event, error) {
	// 获取任务
	task, err := e.taskRepo.GetByAgentAndSeq(agentID, taskSeq)
	if err != nil {
		return nil, fmt.Errorf("get task failed: %w", err)
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}

	now := model.NowMillis()
	var events []Event

	switch taskType {
	case model.TaskTypeMove:
		evts, err := e.executeMoveStart(agentID, taskSeq, params, now)
		if err != nil {
			return nil, err
		}
		events = append(events, evts...)

	case model.TaskTypeHarvest:
		evts, err := e.executeHarvestStart(agentID, taskSeq, params, now)
		if err != nil {
			return nil, err
		}
		events = append(events, evts...)

	case model.TaskTypeCraft:
		evts, err := e.executeCraftStart(agentID, taskSeq, params, now)
		if err != nil {
			return nil, err
		}
		events = append(events, evts...)

	case model.TaskTypeBuild:
		evts, err := e.executeBuildStart(agentID, taskSeq, params, now)
		if err != nil {
			return nil, err
		}
		events = append(events, evts...)

	default:
		// 未知任务类型，立即失败
		events = append(events, &TaskFailEvent{
			BaseEvent: BaseEvent{
				Type:    EventTypeTaskFail,
				Time:    now,
				AgentID: agentID,
				TaskSeq: taskSeq,
			},
			ErrorCode: 1,
			Reason:    "unknown task type",
		})
	}

	return events, nil
}

// ExecuteTaskStep 执行任务步骤
func (e *TaskExecutor) ExecuteTaskStep(agentID int64, taskSeq, step, totalStep int) ([]Event, error) {
	task, err := e.taskRepo.GetByAgentAndSeq(agentID, taskSeq)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}

	now := model.NowMillis()
	var events []Event

	// 检查是否是最后一步
	if step >= totalStep {
		// 任务完成
		events = append(events, &TaskCompleteEvent{
			BaseEvent: BaseEvent{
				Type:    EventTypeTaskComplete,
				Time:    now,
				AgentID: agentID,
				TaskSeq: taskSeq,
			},
			Result: map[string]interface{}{
				"steps_completed": step,
			},
		})
	} else {
		// 继续下一步
		stepDuration := e.getStepDuration(task.Type)
		events = append(events, &TaskStepEvent{
			BaseEvent: BaseEvent{
				Type:    EventTypeTaskStep,
				Time:    now + stepDuration,
				AgentID: agentID,
				TaskSeq: taskSeq,
			},
			Step:      step + 1,
			TotalStep: totalStep,
		})
	}

	return events, nil
}

// executeMoveStart 开始移动任务
func (e *TaskExecutor) executeMoveStart(agentID int64, taskSeq int, params map[string]interface{}, now int64) ([]Event, error) {
	dx := getIntParam(params, "dx", 0)
	dy := getIntParam(params, "dy", 0)
	
	// 计算距离和步数
	distance := int(math.Abs(float64(dx)) + math.Abs(float64(dy)))
	if distance == 0 {
		// 无需移动，立即完成
		return []Event{
			&TaskCompleteEvent{
				BaseEvent: BaseEvent{
					Type:    EventTypeTaskComplete,
					Time:    now,
					AgentID: agentID,
					TaskSeq: taskSeq,
				},
				Result: map[string]interface{}{"moved": 0},
			},
		}, nil
	}

	// 每格 2 秒
	stepDuration := int64(2000)
	
	// 创建第一步事件
	return []Event{
		&TaskStepEvent{
			BaseEvent: BaseEvent{
				Type:    EventTypeTaskStep,
				Time:    now + stepDuration,
				AgentID: agentID,
				TaskSeq: taskSeq,
			},
			Step:      1,
			TotalStep: distance,
		},
	}, nil
}

// executeHarvestStart 开始采集任务
func (e *TaskExecutor) executeHarvestStart(agentID int64, taskSeq int, params map[string]interface{}, now int64) ([]Event, error) {
	// 采集任务：30 秒完成
	duration := int64(30000)
	
	return []Event{
		&TaskCompleteEvent{
			BaseEvent: BaseEvent{
				Type:    EventTypeTaskComplete,
				Time:    now + duration,
				AgentID: agentID,
				TaskSeq: taskSeq,
			},
			Result: map[string]interface{}{
				"resource_id": params["resource_id"],
				"harvested":   5,
			},
		},
	}, nil
}

// executeCraftStart 开始制作任务
func (e *TaskExecutor) executeCraftStart(agentID int64, taskSeq int, params map[string]interface{}, now int64) ([]Event, error) {
	itemName := getStringParam(params, "item", "unknown")
	count := getIntParam(params, "count", 1)
	
	// 查找配方
	recipe := model.GetRecipeByName(itemName)
	if recipe == nil {
		// 未知配方，任务失败
		return []Event{
			&TaskFailEvent{
				BaseEvent: BaseEvent{
					Type:    EventTypeTaskFail,
					Time:    now,
					AgentID: agentID,
					TaskSeq: taskSeq,
				},
				ErrorCode: 1,
				Reason:    "unknown recipe: " + itemName,
			},
		}, nil
	}
	
	// 制作时间使用配方的耗时
	duration := recipe.TimeCost
	
	// 返回完成事件（材料检查在 handler 层做）
	return []Event{
		&TaskCompleteEvent{
			BaseEvent: BaseEvent{
				Type:    EventTypeTaskComplete,
				Time:    now + duration,
				AgentID: agentID,
				TaskSeq: taskSeq,
			},
			Result: map[string]interface{}{
				"item":   itemName,
				"count":  count,
				"recipe": recipe.ID,
			},
		},
	}, nil
}

// executeBuildStart 开始建造任务
func (e *TaskExecutor) executeBuildStart(agentID int64, taskSeq int, params map[string]interface{}, now int64) ([]Event, error) {
	// 建造任务：60 秒
	duration := int64(60000)
	
	return []Event{
		&TaskCompleteEvent{
			BaseEvent: BaseEvent{
				Type:    EventTypeTaskComplete,
				Time:    now + duration,
				AgentID: agentID,
				TaskSeq: taskSeq,
			},
			Result: map[string]interface{}{
				"building_type": params["type"],
				"x":             params["x"],
				"y":             params["y"],
			},
		},
	}, nil
}

// getStepDuration 获取步骤执行时间（毫秒）
func (e *TaskExecutor) getStepDuration(taskType int) int64 {
	switch taskType {
	case model.TaskTypeMove:
		return 2000 // 移动一格 2 秒
	case model.TaskTypeHarvest:
		return 5000 // 采集每阶段 5 秒
	case model.TaskTypeCraft:
		return 2000 // 制作每阶段 2 秒
	case model.TaskTypeBuild:
		return 10000 // 建造每阶段 10 秒
	default:
		return 1000
	}
}

// getIntParam 获取整数参数
func getIntParam(params map[string]interface{}, key string, defaultVal int) int {
	if v, ok := params[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		}
	}
	return defaultVal
}

// getStringParam 获取字符串参数
func getStringParam(params map[string]interface{}, key string, defaultVal string) string {
	if v, ok := params[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}
