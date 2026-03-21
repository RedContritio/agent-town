package api

import (
	"net/http"

	"github.com/RedContritio/agent-town/server/internal/middleware"
	"github.com/RedContritio/agent-town/server/internal/service"
)

// VisionHandler 视野处理器
type VisionHandler struct {
	visionService *service.VisionService
	authMiddleware *middleware.AuthMiddleware
}

// NewVisionHandler 创建视野处理器
func NewVisionHandler(visionService *service.VisionService, authMiddleware *middleware.AuthMiddleware) *VisionHandler {
	return &VisionHandler{
		visionService:  visionService,
		authMiddleware: authMiddleware,
	}
}

// RegisterRoutes 注册路由
func (h *VisionHandler) RegisterRoutes(mux *http.ServeMux) {
	// 与 CLI 命令对齐
	mux.HandleFunc("/api/v1/look", h.authMiddleware.RequireAuth(h.handleLook))
	mux.HandleFunc("/api/v1/scan", h.authMiddleware.RequireAuth(h.handleScan))
}

// handleLook 观察周围 (对应 CLI: look)
func (h *VisionHandler) handleLook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentID, ok := middleware.GetAgentID(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "无法获取 AgentID", "")
		return
	}

	// 获取 Agent 位置
	// TODO: 从 Agent 状态中获取位置
	// 简化：暂时使用 (0,0)
	posX, posY := 0, 0

	req := &service.LookRequest{
		AgentID: agentID,
		PosX:    posX,
		PosY:    posY,
	}

	resp, err := h.visionService.Look(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "观察失败", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleScan 扫描脚下 (对应 CLI: scan)
func (h *VisionHandler) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentID, ok := middleware.GetAgentID(r.Context())
	if !ok {
		writeError(w, http.StatusInternalServerError, "无法获取 AgentID", "")
		return
	}

	// 获取 Agent 位置
	// TODO: 从 Agent 状态中获取位置
	posX, posY := 0, 0

	resp, err := h.visionService.Scan(agentID, posX, posY)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "扫描失败", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
