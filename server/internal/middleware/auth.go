// Package middleware HTTP 中间件
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/RedContritio/agent-town/server/internal/model"
	"github.com/RedContritio/agent-town/server/internal/service"
)

// ContextKey 用于存储在 context 中的键类型
type ContextKey string

const (
	// ContextKeyAgentID 存储 AgentID 的 context key
	ContextKeyAgentID ContextKey = "agent_id"
	// ContextKeyToken 存储 Token 的 context key
	ContextKeyToken ContextKey = "token"
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	authService *service.AuthService
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

// RequireAuth 返回需要认证的处理器包装器
func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return m.RequireAuthWithScopes(0)(next)
}

// RequireAuthWithScopes 返回需要特定权限的处理器包装器
// scopes: 需要的权限位（0 表示只需要有效 Token，不检查具体权限）
func (m *AuthMiddleware) RequireAuthWithScopes(scopes int) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractBearerToken(r)
			if tokenStr == "" {
				writeAuthError(w, "缺少认证 Token")
				return
			}

			token, err := m.authService.ValidateToken(tokenStr)
			if err != nil {
				writeAuthError(w, "无效的 Token: "+err.Error())
				return
			}

			// 检查权限
			if scopes != 0 && !token.HasScope(scopes) {
				writeAuthError(w, "权限不足")
				return
			}

			// 将 AgentID 和 Token 注入 context
			ctx := context.WithValue(r.Context(), ContextKeyAgentID, token.AgentID)
			ctx = context.WithValue(ctx, ContextKeyToken, token)

			next(w, r.WithContext(ctx))
		}
	}
}

// GetAgentID 从 context 获取 AgentID
func GetAgentID(ctx context.Context) (int64, bool) {
	agentID, ok := ctx.Value(ContextKeyAgentID).(int64)
	return agentID, ok
}

// GetToken 从 context 获取 Token
func GetToken(ctx context.Context) (*model.Token, bool) {
	token, ok := ctx.Value(ContextKeyToken).(*model.Token)
	return token, ok
}

// extractBearerToken 从请求头提取 Bearer Token
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

func writeAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"` + message + `"}`))
}
