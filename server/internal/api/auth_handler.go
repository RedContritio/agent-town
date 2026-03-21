// Package api HTTP 处理器
package api

import (
	"encoding/json"
	"net/http"

	"github.com/RedContritio/agent-town/server/internal/service"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterRoutes 注册路由
func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/auth/register", h.handleRegister)
	mux.HandleFunc("/api/v1/auth/challenge", h.handleChallenge)
	mux.HandleFunc("/api/v1/auth/token", h.handleToken)
	mux.HandleFunc("/api/v1/auth/logout", h.handleLogout)
}

// handleRegister 处理注册请求
func (h *AuthHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req service.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误", err.Error())
		return
	}

	resp, err := h.authService.Register(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "注册失败", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

// handleChallenge 处理挑战请求
func (h *AuthHandler) handleChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req service.ChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误", err.Error())
		return
	}

	resp, err := h.authService.CreateChallenge(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, "创建挑战失败", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleToken 处理 Token 请求
func (h *AuthHandler) handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req service.TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误", err.Error())
		return
	}

	resp, err := h.authService.CreateToken(&req)
	if err != nil {
		// 根据错误类型返回不同状态码
		status := http.StatusBadRequest
		if err.Error() == "signature 验证失败" {
			status = http.StatusUnauthorized
		}
		writeError(w, status, "获取 Token 失败", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// handleLogout 处理登出请求
func (h *AuthHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从 Header 获取 Token
	token := extractBearerToken(r)
	if token == "" {
		writeError(w, http.StatusUnauthorized, "未提供 Token", "")
		return
	}

	if err := h.authService.Logout(token); err != nil {
		writeError(w, http.StatusInternalServerError, "登出失败", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "登出成功"})
}

// 辅助函数

// APIError API 错误响应
type APIError struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err string, message string) {
	writeJSON(w, status, APIError{Error: err, Message: message})
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return ""
}
