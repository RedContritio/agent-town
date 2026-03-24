package auth

import (
	"crypto/ed25519"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/RedContritio/agent-town/cli/internal/client"
	"github.com/RedContritio/agent-town/cli/internal/config"
	"github.com/RedContritio/agent-town/cli/internal/keystore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestEnv(t *testing.T) (*config.Config, string) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })

	cfg := &config.Config{
		Agents: make(map[string]*config.AgentConfig),
	}
	return cfg, tmpDir
}

func TestEnsureToken_Cached(t *testing.T) {
	cfg, _ := setupTestEnv(t)

	// 预置未过期的 Token
	cfg.Agents["Alice"] = &config.AgentConfig{
		AgentID:      "agent-001",
		ServerURL:    "http://test.com",
		Token:        "cached-token",
		TokenExpires: time.Now().Add(1 * time.Hour).Unix(),
	}

	// 创建密钥
	keystore.Create("Alice")

	mgr := NewManager(cfg)

	// Mock server 不应该被调用（因为有缓存）
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := client.New(server.URL)
	token, err := mgr.EnsureToken("Alice", c)

	require.NoError(t, err)
	assert.Equal(t, "cached-token", token)
	assert.Equal(t, 0, callCount)       // 没有调用 server
}

func TestEnsureToken_Expired(t *testing.T) {
	cfg, _ := setupTestEnv(t)

	// 预置已过期的 Token
	cfg.Agents["Alice"] = &config.AgentConfig{
		AgentID:      "agent-001",
		ServerURL:    "http://test.com",
		Token:        "expired-token",
		TokenExpires: time.Now().Add(-1 * time.Hour).Unix(),
	}

	// 创建密钥
	key, _ := keystore.Create("Alice")

	mgr := NewManager(cfg)

	// Mock server 返回 challenge 和 token
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/challenge":
			var req client.ChallengeRequest
			json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, key.PublicKeyHex(), req.PublicKey)

			json.NewEncoder(w).Encode(client.ChallengeResponse{
				ChallengeID: "chal-001",
				Nonce:       []byte("test-nonce"),
				ExpiresAt:   time.Now().Add(5 * time.Minute).Unix(),
			})

		case "/api/v1/auth/token":
			var req client.TokenRequest
			json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "chal-001", req.ChallengeID)
			// 验证签名
			assert.True(t, ed25519.Verify(key.PublicKey, []byte("test-nonce"), req.Signature))

			json.NewEncoder(w).Encode(client.TokenResponse{
				Token:     "new-token",
				ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
				AgentID:   "agent-001",
			})
		}
	}))
	defer server.Close()

	c := client.New(server.URL)
	token, err := mgr.EnsureToken("Alice", c)

	require.NoError(t, err)
	assert.Equal(t, "new-token", token)

	// 验证 Token 已保存
	savedToken, _, _ := cfg.GetToken("Alice")
	assert.Equal(t, "new-token", savedToken)
}

func TestEnsureToken_NoToken(t *testing.T) {
	cfg, _ := setupTestEnv(t)

	// 无 Token，但有 Agent 配置
	cfg.Agents["Alice"] = &config.AgentConfig{
		AgentID:   "agent-001",
		ServerURL: "http://test.com",
	}

	// 创建密钥
	key, _ := keystore.Create("Alice")

	mgr := NewManager(cfg)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/challenge":
			json.NewEncoder(w).Encode(client.ChallengeResponse{
				ChallengeID: "chal-002",
				Nonce:       []byte("nonce-xyz"),
				ExpiresAt:   time.Now().Add(5 * time.Minute).Unix(),
			})

		case "/api/v1/auth/token":
			var req client.TokenRequest
			json.NewDecoder(r.Body).Decode(&req)
			// 验证签名正确
			assert.True(t, ed25519.Verify(key.PublicKey, []byte("nonce-xyz"), req.Signature))

			json.NewEncoder(w).Encode(client.TokenResponse{
				Token:     "fresh-token",
				ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
				AgentID:   "agent-001",
			})
		}
	}))
	defer server.Close()

	c := client.New(server.URL)
	token, err := mgr.EnsureToken("Alice", c)

	require.NoError(t, err)
	assert.Equal(t, "fresh-token", token)
}

func TestEnsureToken_KeyNotFound(t *testing.T) {
	cfg, _ := setupTestEnv(t)
	cfg.Agents["Alice"] = &config.AgentConfig{
		AgentID:   "agent-001",
		ServerURL: "http://test.com",
	}

	mgr := NewManager(cfg)
	c := client.New("http://localhost:9999")

	_, err := mgr.EnsureToken("Alice", c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "load agent key")
}

func TestIsExpired(t *testing.T) {
	mgr := &Manager{}

	// 未来时间：未过期
	assert.False(t, mgr.isExpired(time.Now().Add(1*time.Hour).Unix()))

	// 过去时间：已过期
	assert.True(t, mgr.isExpired(time.Now().Add(-1*time.Hour).Unix()))

	// 即将过期（60秒内）：视为已过期（预留缓冲）
	assert.True(t, mgr.isExpired(time.Now().Add(30*time.Second).Unix()))

	// 还有 2 分钟过期：未过期
	assert.False(t, mgr.isExpired(time.Now().Add(2*time.Minute).Unix()))
}

func TestClearToken(t *testing.T) {
	cfg, _ := setupTestEnv(t)
	cfg.Agents["Alice"] = &config.AgentConfig{
		AgentID:      "agent-001",
		Token:        "to-be-cleared",
		TokenExpires: time.Now().Add(1 * time.Hour).Unix(),
	}

	mgr := NewManager(cfg)
	err := mgr.ClearToken("Alice")

	require.NoError(t, err)
	token, expires, _ := cfg.GetToken("Alice")
	assert.Empty(t, token)
	assert.Zero(t, expires)
}

func TestClearToken_AgentNotFound(t *testing.T) {
	cfg, _ := setupTestEnv(t)
	mgr := NewManager(cfg)

	err := mgr.ClearToken("NotExist")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
