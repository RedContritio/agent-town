package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/RedContritio/agent-town/server/internal/engine"
	"github.com/RedContritio/agent-town/server/internal/model"
	"github.com/RedContritio/agent-town/server/internal/repository"
)

// TaskService 任务服务
type TaskService struct {
	taskRepo   *repository.TaskRepository
	agentRepo  *repository.AgentRepository
	craftService *CraftService
}

// NewTaskService 创建任务服务
func NewTaskService(taskRepo *repository.TaskRepository, agentRepo *repository.AgentRepository, craftService *CraftService) *TaskService {
	return &TaskService{
		taskRepo:     taskRepo,
		agentRepo:    agentRepo,
		craftService: craftService,
	}
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	Type   int                 `json:"type"`
	Params map[string]interface{} `json:"params"`
}

// CreateTaskResponse 创建任务响应
type CreateTaskResponse struct {
	TaskID  string `json:"task_id"`
	Type    int    `json:"type"`
	Status  int    `json:"status"`
	Message string `json:"message,omitempty"`
}

// CreateTask 创建任务（压栈）
// 1. 所有现有活跃任务 stack_depth + 1
// 2. 新任务 stack_depth = 0, status = running
func (s *TaskService) CreateTask(agentID int64, req *CreateTaskRequest) (*CreateTaskResponse, error) {
	// 验证 Agent 存在
	agent, err := s.agentRepo.GetByID(agentID)
	if err != nil {
		return nil, fmt.Errorf("查询 Agent 失败: %w", err)
	}
	if agent == nil {
		return nil, errors.New("agent 不存在")
	}

	// 对于制作任务，检查配方和材料
	if req.Type == model.TaskTypeCraft && s.craftService != nil {
		itemName := ""
		if v, ok := req.Params["item"]; ok {
			itemName = fmt.Sprintf("%v", v)
		}
		
		recipe, err := s.craftService.ValidateRecipe(itemName)
		if err != nil {
			return nil, err
		}
		
		canCraft, reason := s.craftService.CanCraft(agentID, recipe)
		if !canCraft {
			return nil, errors.New(reason)
		}
	}

	// 获取下一个序列号
	seq, err := s.taskRepo.GetNextSeq(agentID)
	if err != nil {
		return nil, fmt.Errorf("获取序列号失败: %w", err)
	}

	// 序列化参数
	paramsJSON, err := json.Marshal(req.Params)
	if err != nil {
		return nil, fmt.Errorf("参数序列化失败: %w", err)
	}

	now := model.NowMillis()

	// 在事务中执行
	err = s.executeWithTx(func(tx *sql.Tx) error {
		// 1. 暂停当前栈顶任务（depth=0）
		_, err := tx.Exec(
			"UPDATE tasks SET status=? WHERE agent_id=? AND stack_depth=0 AND status=?",
			model.TaskStatusPaused, agentID, model.TaskStatusRunning,
		)
		if err != nil {
			return err
		}

		// 2. 增加所有活跃任务的栈深度
		if err := s.taskRepo.IncrementStackDepth(tx, agentID); err != nil {
			return err
		}

		// 3. 创建新任务（栈顶）
		task := &model.Task{
			AgentID:    agentID,
			Seq:        seq,
			Type:       req.Type,
			Status:     model.TaskStatusRunning, // 新任务直接运行
			Params:     string(paramsJSON),
			StackDepth: 0, // 栈顶
			CreatedAt:  now,
			StartedAt:  &now,
		}

		return s.taskRepo.Create(tx, task)
	})

	if err != nil {
		return nil, fmt.Errorf("创建任务失败: %w", err)
	}

	// 调度任务到引擎执行
	if err := engine.GetManager().ScheduleTask(agentID, seq); err != nil {
		// 调度失败不影响任务创建，只是不会自动执行
		// 后续可以通过手动恢复触发执行
		fmt.Printf("Schedule task failed: %v\n", err)
	}

	return &CreateTaskResponse{
		TaskID:  fmt.Sprintf("%s-%03d", agent.Name, seq),
		Type:    req.Type,
		Status:  model.TaskStatusRunning,
		Message: "任务已创建并开始执行",
	}, nil
}

// GetStack 获取任务栈
func (s *TaskService) GetStack(agentID int64) ([]*model.Task, error) {
	// 验证 Agent 存在
	agent, err := s.agentRepo.GetByID(agentID)
	if err != nil {
		return nil, fmt.Errorf("查询 Agent 失败: %w", err)
	}
	if agent == nil {
		return nil, errors.New("agent 不存在")
	}

	return s.taskRepo.GetStack(agentID)
}

// GetTask 获取单个任务
func (s *TaskService) GetTask(agentID int64, seq int) (*model.Task, error) {
	return s.taskRepo.GetByAgentAndSeq(agentID, seq)
}

// DropTask 放弃任务
func (s *TaskService) DropTask(agentID int64, seq int) error {
	task, err := s.taskRepo.GetByAgentAndSeq(agentID, seq)
	if err != nil {
		return fmt.Errorf("查询任务失败: %w", err)
	}
	if task == nil {
		return errors.New("任务不存在")
	}

	// 只能在事务中删除
	return s.executeWithTx(func(tx *sql.Tx) error {
		// 获取当前任务的深度
		oldDepth := task.StackDepth

		// 删除任务
		if err := s.taskRepo.Delete(tx, agentID, seq); err != nil {
			return err
		}

		// 如果删除的是栈顶任务（depth=0），需要恢复下一个任务
		if oldDepth == 0 {
			// 将 depth=1 的任务改为 depth=0 并恢复运行
			// 这里我们需要找到 depth=1 的任务
			rows, err := tx.Query(
				"SELECT seq FROM tasks WHERE agent_id=? AND stack_depth=1 AND status=?",
				agentID, model.TaskStatusPaused,
			)
			if err != nil {
				return err
			}
			defer rows.Close()

			for rows.Next() {
				var nextSeq int
				if err := rows.Scan(&nextSeq); err != nil {
					return err
				}
				// 更新为栈顶并恢复运行
				if err := s.taskRepo.UpdateStackDepth(tx, agentID, nextSeq, 0); err != nil {
					return err
				}
				if err := s.taskRepo.UpdateStatus(tx, agentID, nextSeq, model.TaskStatusRunning); err != nil {
					return err
				}
			}
		}

		// 所有比删除任务深的任务都需要减 1
		_, err := tx.Exec(
			"UPDATE tasks SET stack_depth = stack_depth - 1 WHERE agent_id=? AND stack_depth > ?",
			agentID, oldDepth,
		)

		return err
	})
}

// CompleteTask 完成任务
func (s *TaskService) CompleteTask(agentID int64, seq int, result string) error {
	task, err := s.taskRepo.GetByAgentAndSeq(agentID, seq)
	if err != nil {
		return fmt.Errorf("查询任务失败: %w", err)
	}
	if task == nil {
		return errors.New("任务不存在")
	}

	return s.executeWithTx(func(tx *sql.Tx) error {
		// 标记当前任务完成
		if err := s.taskRepo.CompleteTask(tx, agentID, seq, result); err != nil {
			return err
		}

		// 恢复下一个任务（depth=1 改为 depth=0 并运行）
		rows, err := tx.Query(
			"SELECT seq FROM tasks WHERE agent_id=? AND stack_depth=1 AND status=?",
			agentID, model.TaskStatusPaused,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var nextSeq int
			if err := rows.Scan(&nextSeq); err != nil {
				return err
			}
			// 更新为栈顶
			if err := s.taskRepo.UpdateStackDepth(tx, agentID, nextSeq, 0); err != nil {
				return err
			}
			// 恢复运行
			if err := s.taskRepo.UpdateStatus(tx, agentID, nextSeq, model.TaskStatusRunning); err != nil {
				return err
			}
		}

		// 调整其他任务的深度
		_, err = tx.Exec(
			"UPDATE tasks SET stack_depth = stack_depth - 1 WHERE agent_id=? AND stack_depth > 0",
			agentID,
		)

		return err
	})
}

// PauseTask 暂停任务
func (s *TaskService) PauseTask(agentID int64, seq int) error {
	task, err := s.taskRepo.GetByAgentAndSeq(agentID, seq)
	if err != nil {
		return fmt.Errorf("查询任务失败: %w", err)
	}
	if task == nil {
		return errors.New("任务不存在")
	}
	if task.Status != model.TaskStatusRunning {
		return errors.New("只能暂停运行中的任务")
	}

	return s.executeWithTx(func(tx *sql.Tx) error {
		return s.taskRepo.UpdateStatus(tx, agentID, seq, model.TaskStatusPaused)
	})
}

// ResumeTask 恢复任务
func (s *TaskService) ResumeTask(agentID int64, seq int) error {
	task, err := s.taskRepo.GetByAgentAndSeq(agentID, seq)
	if err != nil {
		return fmt.Errorf("查询任务失败: %w", err)
	}
	if task == nil {
		return errors.New("任务不存在")
	}
	if task.Status != model.TaskStatusPaused {
		return errors.New("只能恢复已暂停的任务")
	}

	// 如果恢复的是栈顶任务，直接恢复
	if task.StackDepth == 0 {
		return s.executeWithTx(func(tx *sql.Tx) error {
			return s.taskRepo.UpdateStatus(tx, agentID, seq, model.TaskStatusRunning)
		})
	}

	// 如果恢复的不是栈顶任务，需要将它移到栈顶
	// 这会暂停当前栈顶任务
	return s.executeWithTx(func(tx *sql.Tx) error {
		oldDepth := task.StackDepth

		// 1. 暂停当前栈顶任务
		_, err = tx.Exec(
			"UPDATE tasks SET status=? WHERE agent_id=? AND stack_depth=0 AND status=?",
			model.TaskStatusPaused, agentID, model.TaskStatusRunning,
		)
		if err != nil {
			return err
		}

		// 2. 增加所有比目标任务浅的任务的深度
		_, err = tx.Exec(
			"UPDATE tasks SET stack_depth = stack_depth + 1 WHERE agent_id=? AND stack_depth < ?",
			agentID, oldDepth,
		)
		if err != nil {
			return err
		}

		// 3. 将目标任务设为栈顶并恢复
		if err := s.taskRepo.UpdateStackDepth(tx, agentID, seq, 0); err != nil {
			return err
		}
		return s.taskRepo.UpdateStatus(tx, agentID, seq, model.TaskStatusRunning)
	})
}

// GetTaskTypeName 获取任务类型名称
func (s *TaskService) GetTaskTypeName(taskType int) string {
	return model.GetTaskTypeName(taskType)
}

// GetTaskStatusName 获取任务状态名称
func (s *TaskService) GetTaskStatusName(status int) string {
	return model.GetTaskStatusName(status)
}

// executeWithTx 在事务中执行函数
func (s *TaskService) executeWithTx(fn func(*sql.Tx) error) error {
	tx, err := s.taskRepo.DB.Begin()
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
