package api

import (
	"net/http"

	"github.com/RedContritio/agent-town/server/internal/middleware"
	"github.com/RedContritio/agent-town/server/internal/service"
)

// InventoryHandler 背包处理器
type InventoryHandler struct {
	invService *service.InventoryService
	authMiddleware *middleware.AuthMiddleware
}

// NewInventoryHandler 创建背包处理器
func NewInventoryHandler(invService *service.InventoryService, authMiddleware *middleware.AuthMiddleware) *InventoryHandler {
	return &InventoryHandler{
		invService:     invService,
		authMiddleware: authMiddleware,
	}
}

// RegisterRoutes 注册路由
func (h *InventoryHandler) RegisterRoutes(mux *http.ServeMux) {
	// 与 CLI 命令 `inventory` 对齐
	mux.HandleFunc("/api/v1/inventory", h.authMiddleware.RequireAuth(h.handleInventory))
}

// handleInventory 获取背包内容 (对应 CLI: inventory)
func (h *InventoryHandler) handleInventory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentID, ok := middleware.GetAgentID(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "无法获取 AgentID", "")
		return
	}

	resp, err := h.invService.GetInventory(agentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "获取背包失败", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
