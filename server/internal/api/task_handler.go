package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/RedContritio/agent-town/server/internal/middleware"
	"github.com/RedContritio/agent-town/server/internal/model"
	"github.com/RedContritio/agent-town/server/internal/service"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	taskService    *service.TaskService
	agentService   *service.AgentService
	authMiddleware *middleware.AuthMiddleware
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(
	taskService *service.TaskService,
	agentService *service.AgentService,
	authMiddleware *middleware.AuthMiddleware,
) *TaskHandler {
	return &TaskHandler{
		taskService:    taskService,
		agentService:   agentService,
		authMiddleware: authMiddleware,
	}
}

// RegisterRoutes 注册路由
func (h *TaskHandler) RegisterRoutes(mux *http.ServeMux) {
	// 任务栈管理 - 与 CLI 命令 `stack` 对齐
	mux.HandleFunc("/api/v1/stack", h.authMiddleware.RequireAuth(h.handleStack))
	mux.HandleFunc("/api/v1/stack/", h.authMiddleware.RequireAuth(h.handleTaskDetail))
}

// handleStack 处理任务列表和创建 (对应 CLI: stack)
func (h *TaskHandler) handleStack(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetStack(w, r)
	case http.MethodPost:
		h.handleCreateTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetStack 获取任务栈
func (h *TaskHandler) handleGetStack(w http.ResponseWriter, r *http.Request) {
	agentID, ok := middleware.GetAgentID(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "无法获取 AgentID", "")
		return
	}

	// 获取 Agent 名称用于生成 task_id
	agent, err := h.agentService.GetAgent(agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "获取 Agent 失败", err.Error())
		return
	}
	if agent == nil {
		writeError(w, http.StatusNotFound, "Agent 不存在", "")
		return
	}

	stack, err := h.taskService.GetStack(agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "获取任务栈失败", err.Error())
		return
	}

	// 转换为响应格式
	var stackResp []map[string]interface{}
	for _, task := range stack {
		stackResp = append(stackResp, task.ToAPIResponse(agent.Name))
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"agent_id": agentID,
		"stack":    stackResp,
	})
}

// handleCreateTask 创建任务
func (h *TaskHandler) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	agentID, ok := middleware.GetAgentID(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "无法获取 AgentID", "")
		return
	}

	var req service.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误", err.Error())
		return
	}

	resp, err := h.taskService.CreateTask(agentID, &req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "创建任务失败", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// handleTaskDetail 处理单个任务操作
func (h *TaskHandler) handleTaskDetail(w http.ResponseWriter, r *http.Request) {
	// 解析 URL 路径获取 task_id
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/stack/")
	if path == "" {
		writeError(w, http.StatusBadRequest, "缺少任务ID", "")
		return
	}

	// 提取 task_id（去除后缀如 /pause, /resume）
	parts := strings.Split(path, "/")
	taskID := parts[0]

	// 解析 agentName 和 seq
	agentName, seq, err := model.ParseTaskID(taskID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "无效的任务ID", err.Error())
		return
	}

	// 获取当前 agent
	agentID, ok := middleware.GetAgentID(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "无法获取 AgentID", "")
		return
	}

	agent, err := h.agentService.GetAgent(agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "获取 Agent 失败", err.Error())
		return
	}
	if agent == nil {
		writeError(w, http.StatusNotFound, "Agent 不存在", "")
		return
	}

	// 验证任务属于当前 Agent
	if agent.Name != agentName {
		writeError(w, http.StatusForbidden, "无权访问此任务", "")
		return
	}

	// 根据路径和方法处理
	if len(parts) > 1 {
		// 有后缀操作
		switch parts[1] {
		case "pause":
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			h.handlePauseTask(w, r, agentID, seq)
		case "resume":
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			h.handleResumeTask(w, r, agentID, seq)
		default:
			writeError(w, http.StatusNotFound, "未知操作", "")
		}
		return
	}

	// 无后缀，处理 GET 和 DELETE
	switch r.Method {
	case http.MethodGet:
		h.handleGetTask(w, r, agentID, seq)
	case http.MethodDelete:
		h.handleDropTask(w, r, agentID, seq)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetTask 获取单个任务详情
func (h *TaskHandler) handleGetTask(w http.ResponseWriter, r *http.Request, agentID int64, seq int) {
	agent, _ := h.agentService.GetAgent(agentID)

	task, err := h.taskService.GetTask(agentID, seq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "获取任务失败", err.Error())
		return
	}
	if task == nil {
		writeError(w, http.StatusNotFound, "任务不存在", "")
		return
	}

	writeJSON(w, http.StatusOK, task.ToAPIResponse(agent.Name))
}

// handleDropTask 放弃任务
func (h *TaskHandler) handleDropTask(w http.ResponseWriter, r *http.Request, agentID int64, seq int) {
	if err := h.taskService.DropTask(agentID, seq); err != nil {
		writeError(w, http.StatusBadRequest, "放弃任务失败", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "任务已放弃",
	})
}

// handlePauseTask 暂停任务
func (h *TaskHandler) handlePauseTask(w http.ResponseWriter, r *http.Request, agentID int64, seq int) {
	if err := h.taskService.PauseTask(agentID, seq); err != nil {
		writeError(w, http.StatusBadRequest, "暂停任务失败", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "任务已暂停",
	})
}

// handleResumeTask 恢复任务
func (h *TaskHandler) handleResumeTask(w http.ResponseWriter, r *http.Request, agentID int64, seq int) {
	if err := h.taskService.ResumeTask(agentID, seq); err != nil {
		writeError(w, http.StatusBadRequest, "恢复任务失败", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "任务已恢复",
	})
}

// parseTaskID 解析任务ID（简单版本）
func parseTaskID(taskID string) (string, int, error) {
	// 格式: name-NNN
	parts := strings.Split(taskID, "-")
	if len(parts) < 2 {
		return "", 0, errors.New("invalid task id format")
	}

	agentName := strings.Join(parts[:len(parts)-1], "-")
	seq, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return "", 0, err
	}

	return agentName, seq, nil
}
