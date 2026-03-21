package model

import (
	"encoding/json"
	"fmt"
)

// Task 任务实体
type Task struct {
	AgentID      int64  `json:"-"` // AgentID 不序列化到 JSON（在任务ID中体现）
	Seq          int    `json:"seq"`
	Type         int    `json:"type"`
	Status       int    `json:"status"`
	Params       string `json:"params,omitempty"`
	StackDepth   int    `json:"stack_depth"`
	Result       string `json:"result,omitempty"`
	ErrorCode    *int   `json:"error_code,omitempty"`
	CreatedAt    int64  `json:"created_at"`
	StartedAt    *int64 `json:"started_at,omitempty"`
	CompletedAt  *int64 `json:"completed_at,omitempty"`
}

// TaskID 获取任务ID（格式：<agent-name>-<seq>）
func (t *Task) TaskID(agentName string) string {
	return fmt.Sprintf("%s-%03d", agentName, t.Seq)
}

// ParseTaskID 解析任务ID
// 返回 agentName, seq, error
func ParseTaskID(taskID string) (string, int, error) {
	// 简单解析：假设格式为 name-NNN
	var agentName string
	var seq int
	_, err := fmt.Sscanf(taskID, "%[^-]-%d", &agentName, &seq)
	if err != nil {
		return "", 0, fmt.Errorf("无效的任务ID格式: %s", taskID)
	}
	return agentName, seq, nil
}

// IsActive 检查任务是否活跃（pending, running, paused）
func (t *Task) IsActive() bool {
	return t.Status == TaskStatusPending || t.Status == TaskStatusRunning || t.Status == TaskStatusPaused
}

// ToAPIResponse 转换为 API 响应
func (t *Task) ToAPIResponse(agentName string) map[string]interface{} {
	resp := map[string]interface{}{
		"task_id":     t.TaskID(agentName),
		"type":        t.Type,
		"status":      t.Status,
		"stack_depth": t.StackDepth,
		"created_at":  t.CreatedAt,
	}

	if t.Params != "" {
		var params interface{}
		if err := json.Unmarshal([]byte(t.Params), &params); err == nil {
			resp["params"] = params
		} else {
			resp["params"] = t.Params
		}
	}

	if t.Result != "" {
		var result interface{}
		if err := json.Unmarshal([]byte(t.Result), &result); err == nil {
			resp["result"] = result
		} else {
			resp["result"] = t.Result
		}
	}

	if t.ErrorCode != nil {
		resp["error_code"] = *t.ErrorCode
	}

	if t.StartedAt != nil {
		resp["started_at"] = *t.StartedAt
	}

	if t.CompletedAt != nil {
		resp["completed_at"] = *t.CompletedAt
	}

	return resp
}

// TaskType 常量（已在 schema.go 定义，这里重复方便使用）
const (
	TaskTypeMove    = 0
	TaskTypeHarvest = 1
	TaskTypeCraft   = 2
	TaskTypeBuild   = 3
	TaskTypeCombat  = 4
)

// TaskStatus 常量（已在 schema.go 定义，这里重复方便使用）
const (
	TaskStatusPending   = 0
	TaskStatusRunning   = 1
	TaskStatusPaused    = 2
	TaskStatusCompleted = 3
	TaskStatusFailed    = 4
)

// TaskTypeName 任务类型名称映射
var TaskTypeName = map[int]string{
	TaskTypeMove:    "move",
	TaskTypeHarvest: "harvest",
	TaskTypeCraft:   "craft",
	TaskTypeBuild:   "build",
	TaskTypeCombat:  "combat",
}

// TaskStatusName 任务状态名称映射
var TaskStatusName = map[int]string{
	TaskStatusPending:   "pending",
	TaskStatusRunning:   "running",
	TaskStatusPaused:    "paused",
	TaskStatusCompleted: "completed",
	TaskStatusFailed:    "failed",
}

// GetTaskTypeName 获取任务类型名称
func GetTaskTypeName(taskType int) string {
	if name, ok := TaskTypeName[taskType]; ok {
		return name
	}
	return "unknown"
}

// GetTaskStatusName 获取任务状态名称
func GetTaskStatusName(status int) string {
	if name, ok := TaskStatusName[status]; ok {
		return name
	}
	return "unknown"
}

// NewTask 创建新任务
func NewTask(taskType int, params map[string]interface{}) (*Task, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	now := NowMillis()
	return &Task{
		Type:      taskType,
		Status:    TaskStatusPending,
		Params:    string(paramsJSON),
		CreatedAt: now,
	}, nil
}
