package api

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RedContritio/agent-town/server/internal/db"
	"github.com/RedContritio/agent-town/server/internal/repository"
	"github.com/RedContritio/agent-town/server/internal/service"
	_ "modernc.org/sqlite"
)

// setupTestHandler 创建测试处理器
func setupTestHandler(t *testing.T) (*AuthHandler, ed25519.PublicKey, ed25519.PrivateKey) {
	database, err := db.Init(&db.Config{DataDir: ":memory:", DBName: ""})
	if err != nil {
		t.Fatalf("初始化数据库失败: %v", err)
	}

	agentRepo := repository.NewAgentRepository(database)
	tokenRepo := repository.NewTokenRepository(database)
	authService := service.NewAuthService(agentRepo, tokenRepo)
	handler := NewAuthHandler(authService)

	publicKey, privateKey, _ := ed25519.GenerateKey(nil)
	return handler, publicKey, privateKey
}

func TestAuthHandler_Register(t *testing.T) {
	handler, publicKey, _ := setupTestHandler(t)

	// 创建请求
	reqBody := map[string]string{
		"public_key": base64.StdEncoding.EncodeToString(publicKey),
		"name":       "HandlerTest",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// 调用处理器
	handler.handleRegister(w, req)

	// 验证响应
	if w.Code != http.StatusCreated {
		t.Errorf("期望状态码 %d, 得到 %d, 响应: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if resp["agent_id"] == nil {
		t.Error("响应缺少 agent_id")
	}
	if resp["name"] != "HandlerTest" {
		t.Errorf("期望 name=HandlerTest, 得到 %v", resp["name"])
	}
}

func TestAuthHandler_Register_InvalidMethod(t *testing.T) {
	handler, _, _ := setupTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/register", nil)
	w := httptest.NewRecorder()

	handler.handleRegister(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestAuthHandler_Register_InvalidBody(t *testing.T) {
	handler, _, _ := setupTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()

	handler.handleRegister(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthHandler_Challenge(t *testing.T) {
	handler, publicKey, _ := setupTestHandler(t)
	publicKeyB64 := base64.StdEncoding.EncodeToString(publicKey)

	// 先注册
	handler.handleRegister(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/v1/auth/register",
		bytes.NewReader(mustMarshal(map[string]string{
			"public_key": publicKeyB64,
			"name":       "ChallengeHandlerTest",
		}))))

	// 创建挑战请求
	reqBody := map[string]string{"public_key": publicKeyB64}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/challenge", bytes.NewReader(mustMarshal(reqBody)))
	w := httptest.NewRecorder()

	handler.handleChallenge(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 得到 %d, 响应: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if resp["challenge_id"] == nil {
		t.Error("响应缺少 challenge_id")
	}
	if resp["challenge"] == nil {
		t.Error("响应缺少 challenge")
	}
}

func TestAuthHandler_Token(t *testing.T) {
	handler, publicKey, privateKey := setupTestHandler(t)
	publicKeyB64 := base64.StdEncoding.EncodeToString(publicKey)

	// 注册
	handler.handleRegister(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/v1/auth/register",
		bytes.NewReader(mustMarshal(map[string]string{
			"public_key": publicKeyB64,
			"name":       "TokenHandlerTest",
		}))))

	// 创建挑战
	challengeRecorder := httptest.NewRecorder()
	handler.handleChallenge(challengeRecorder, httptest.NewRequest(http.MethodPost, "/api/v1/auth/challenge",
		bytes.NewReader(mustMarshal(map[string]string{"public_key": publicKeyB64}))))

	var challengeResp map[string]interface{}
	json.Unmarshal(challengeRecorder.Body.Bytes(), &challengeResp)
	challenge := challengeResp["challenge"].(string)
	challengeID := challengeResp["challenge_id"].(string)

	// 签名
	signature := ed25519.Sign(privateKey, []byte(challenge))

	// 获取 Token
	reqBody := map[string]string{
		"challenge_id": challengeID,
		"signature":    base64.StdEncoding.EncodeToString(signature),
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/token", bytes.NewReader(mustMarshal(reqBody)))
	w := httptest.NewRecorder()

	handler.handleToken(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 得到 %d, 响应: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var tokenResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &tokenResp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if tokenResp["token"] == nil {
		t.Error("响应缺少 token")
	}
}

func TestAuthHandler_Token_InvalidSignature(t *testing.T) {
	handler, publicKey, _ := setupTestHandler(t)
	publicKeyB64 := base64.StdEncoding.EncodeToString(publicKey)

	// 注册
	handler.handleRegister(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/v1/auth/register",
		bytes.NewReader(mustMarshal(map[string]string{
			"public_key": publicKeyB64,
			"name":       "InvalidSigTest",
		}))))

	// 创建挑战
	challengeRecorder := httptest.NewRecorder()
	handler.handleChallenge(challengeRecorder, httptest.NewRequest(http.MethodPost, "/api/v1/auth/challenge",
		bytes.NewReader(mustMarshal(map[string]string{"public_key": publicKeyB64}))))

	var challengeResp map[string]interface{}
	json.Unmarshal(challengeRecorder.Body.Bytes(), &challengeResp)
	challengeID := challengeResp["challenge_id"].(string)

	// 使用无效签名
	_, wrongPrivateKey, _ := ed25519.GenerateKey(nil)
	wrongSignature := ed25519.Sign(wrongPrivateKey, []byte("wrong"))

	reqBody := map[string]string{
		"challenge_id": challengeID,
		"signature":    base64.StdEncoding.EncodeToString(wrongSignature),
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/token", bytes.NewReader(mustMarshal(reqBody)))
	w := httptest.NewRecorder()

	handler.handleToken(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	handler, publicKey, privateKey := setupTestHandler(t)
	publicKeyB64 := base64.StdEncoding.EncodeToString(publicKey)

	// 注册 + 获取 Token
	handler.handleRegister(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/v1/auth/register",
		bytes.NewReader(mustMarshal(map[string]string{
			"public_key": publicKeyB64,
			"name":       "LogoutTest",
		}))))

	challengeRecorder := httptest.NewRecorder()
	handler.handleChallenge(challengeRecorder, httptest.NewRequest(http.MethodPost, "/api/v1/auth/challenge",
		bytes.NewReader(mustMarshal(map[string]string{"public_key": publicKeyB64}))))

	var challengeResp map[string]interface{}
	json.Unmarshal(challengeRecorder.Body.Bytes(), &challengeResp)
	challenge := challengeResp["challenge"].(string)
	challengeID := challengeResp["challenge_id"].(string)

	signature := ed25519.Sign(privateKey, []byte(challenge))

	tokenRecorder := httptest.NewRecorder()
	handler.handleToken(tokenRecorder, httptest.NewRequest(http.MethodPost, "/api/v1/auth/token",
		bytes.NewReader(mustMarshal(map[string]string{
			"challenge_id": challengeID,
			"signature":    base64.StdEncoding.EncodeToString(signature),
		}))))

	var tokenResp map[string]interface{}
	json.Unmarshal(tokenRecorder.Body.Bytes(), &tokenResp)
	token := tokenResp["token"].(string)

	// 登出
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.handleLogout(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 得到 %d, 响应: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestAuthHandler_Logout_NoToken(t *testing.T) {
	handler, _, _ := setupTestHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/auth/logout", nil)
	w := httptest.NewRecorder()

	handler.handleLogout(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusUnauthorized, w.Code)
	}
}

// mustMarshal JSON 编码辅助函数
func mustMarshal(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
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
