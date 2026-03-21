package api

import (
	"net/http"

	"github.com/RedContritio/agent-town/server/internal/middleware"
	"github.com/RedContritio/agent-town/server/internal/service"
)

// AgentHandler Agent 处理器
type AgentHandler struct {
	agentService *service.AgentService
	authMiddleware *middleware.AuthMiddleware
}

// NewAgentHandler 创建 Agent 处理器
func NewAgentHandler(agentService *service.AgentService, authMiddleware *middleware.AuthMiddleware) *AgentHandler {
	return &AgentHandler{
		agentService:   agentService,
		authMiddleware: authMiddleware,
	}
}

// RegisterRoutes 注册路由
func (h *AgentHandler) RegisterRoutes(mux *http.ServeMux) {
	// 需要认证的端点 - 与 CLI 命令对齐
	mux.HandleFunc("/api/v1/status", h.authMiddleware.RequireAuth(h.handleGetStatus))
}

// handleGetStatus 获取当前 Agent 状态 (对应 CLI: status)
func (h *AgentHandler) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentID, ok := middleware.GetAgentID(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "无法获取 AgentID", "")
		return
	}

	status, err := h.agentService.GetAgentStatus(agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "获取状态失败", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, status)
}
