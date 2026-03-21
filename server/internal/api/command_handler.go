package api

import (
	"net/http"

	"github.com/RedContritio/agent-town/server/internal/model"
)

// CommandHandler 命令处理器
type CommandHandler struct {
	commands []model.CommandDefinition
}

// NewCommandHandler 创建命令处理器
func NewCommandHandler() *CommandHandler {
	return &CommandHandler{
		commands: model.DefaultCommands(),
	}
}

// RegisterRoutes 注册路由
func (h *CommandHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/commands", h.handleGetCommands)
}

// handleGetCommands 获取命令列表
func (h *CommandHandler) handleGetCommands(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := model.CommandsResponse{
		Commands: h.commands,
		Version:  "1.0",
	}

	writeJSON(w, http.StatusOK, resp)
}
