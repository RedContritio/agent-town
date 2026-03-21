package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RedContritio/agent-town/server/internal/db"
	"github.com/RedContritio/agent-town/server/internal/repository"
	"github.com/RedContritio/agent-town/server/internal/service"
	_ "modernc.org/sqlite"
)

func setupTestMiddleware(t *testing.T) (*AuthMiddleware, *service.AuthService) {
	database, err := db.Init(&db.Config{DataDir: ":memory:", DBName: ""})
	if err != nil {
		t.Fatalf("初始化数据库失败: %v", err)
	}

	agentRepo := repository.NewAgentRepository(database)
	tokenRepo := repository.NewTokenRepository(database)
	authService := service.NewAuthService(agentRepo, tokenRepo)
	middleware := NewAuthMiddleware(authService)

	return middleware, authService
}

func TestRequireAuth_Success(t *testing.T) {
	middleware, authService := setupTestMiddleware(t)

	// 创建一个测试用的 token
	// 这里我们直接使用 repository 插入一个 token
	database, _ := db.Init(&db.Config{DataDir: ":memory:", DBName: ""})
	agentRepo := repository.NewAgentRepository(database)
	tokenRepo := repository.NewTokenRepository(database)

	// 插入测试 agent
	_, err := database.Exec(
		`INSERT INTO agents (public_key, name, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		[]byte("test-key"), "TestAgent", 1234567890000, 1234567890000,
	)
	if err != nil {
		t.Fatalf("插入 agent 失败: %v", err)
	}

	// 插入测试 token
	expiresAt := int64(9999999999999)
	_, err = database.Exec(
		`INSERT INTO tokens (token, agent_id, token_type, scopes, created_at, expires_at) VALUES (?, ?, ?, ?, ?, ?)`,
		"test-token-123", 1, 0, 0x1F, 1234567890000, expiresAt,
	)
	if err != nil {
		t.Fatalf("插入 token 失败: %v", err)
	}

	// 重新创建 middleware 使用同一个数据库
	authService = service.NewAuthService(agentRepo, tokenRepo)
	middleware = NewAuthMiddleware(authService)

	handlerCalled := false
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		agentID, ok := GetAgentID(r.Context())
		if !ok {
			t.Error("无法从 context 获取 AgentID")
		}
		if agentID != 1 {
			t.Errorf("期望 AgentID=1, 得到 %d", agentID)
		}
		w.WriteHeader(http.StatusOK)
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer test-token-123")
	w := httptest.NewRecorder()

	middleware.RequireAuth(testHandler)(w, req)

	if !handlerCalled {
		t.Error("处理器未被调用")
	}
	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusOK, w.Code)
	}
}

func TestRequireAuth_MissingToken(t *testing.T) {
	middleware, _ := setupTestMiddleware(t)

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		t.Error("处理器不应被调用")
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	middleware.RequireAuth(testHandler)(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	middleware, _ := setupTestMiddleware(t)

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		t.Error("处理器不应被调用")
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	middleware.RequireAuth(testHandler)(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRequireAuthWithScopes(t *testing.T) {
	middleware, _ := setupTestMiddleware(t)

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	// 测试没有权限要求的情况（scopes=0）
	// 由于我们没有有效的 token，这个测试主要验证中间件结构正确
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid")
	w := httptest.NewRecorder()

	middleware.RequireAuthWithScopes(0)(testHandler)(w, req)

	// 应该失败，因为 token 无效
	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusUnauthorized, w.Code)
	}
}

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name     string
		auth     string
		expected string
	}{
		{
			name:     "有效 Bearer",
			auth:     "Bearer abc123",
			expected: "abc123",
		},
		{
			name:     "小写 bearer",
			auth:     "bearer abc123",
			expected: "abc123",
		},
		{
			name:     "无 Bearer",
			auth:     "abc123",
			expected: "",
		},
		{
			name:     "空",
			auth:     "",
			expected: "",
		},
		{
			name:     "Bearer 后无内容",
			auth:     "Bearer ",
			expected: "",
		},
		{
			name:     "Bearer 后多空格",
			auth:     "Bearer   abc123  ",
			expected: "abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.auth != "" {
				req.Header.Set("Authorization", tt.auth)
			}
			result := extractBearerToken(req)
			if result != tt.expected {
				t.Errorf("期望 %q, 得到 %q", tt.expected, result)
			}
		})
	}
}
