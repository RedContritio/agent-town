package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/RedContritio/agent-town/cli/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestEnv(t *testing.T) string {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })
	return tmpDir
}

func TestRunAgentCreate_Success(t *testing.T) {
	setupTestEnv(t)

	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/auth/register", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req client.RegisterRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "Alice", req.Name)
		assert.NotEmpty(t, req.PublicKey)

		json.NewEncoder(w).Encode(client.RegisterResponse{
			AgentID: "agent-alice-001",
		})
	}))
	defer server.Close()

	err := RunAgentCreate("Alice", server.URL)
	require.NoError(t, err)

	// 验证密钥文件创建
	keyPath := filepath.Join(os.Getenv("HOME"), ".at-cli", "agents", "Alice.pem")
	_, err = os.Stat(keyPath)
	require.NoError(t, err)

	// 验证配置文件
	configPath := filepath.Join(os.Getenv("HOME"), ".at-cli", "config.yaml")
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "agent-alice-001")
	assert.Contains(t, string(data), server.URL)
}

func TestRunAgentCreate_AlreadyExists(t *testing.T) {
	setupTestEnv(t)

	// 先创建一个
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(client.RegisterResponse{AgentID: "agent-001"})
	}))
	defer server.Close()

	RunAgentCreate("Alice", server.URL)

	// 再创建同名应该失败
	err := RunAgentCreate("Alice", server.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestRunAgentCreate_ServerError(t *testing.T) {
	setupTestEnv(t)

	// Mock server 返回错误
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error": "name taken"}`))
	}))
	defer server.Close()

	err := RunAgentCreate("Alice", server.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "register with server")

	// 验证密钥被回滚删除
	keyPath := filepath.Join(os.Getenv("HOME"), ".at-cli", "agents", "Alice.pem")
	_, err = os.Stat(keyPath)
	assert.True(t, os.IsNotExist(err))
}

func TestRunAgentList(t *testing.T) {
	setupTestEnv(t)

	// 先创建两个 agent
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(client.RegisterResponse{AgentID: "agent-001"})
	}))
	defer server.Close()

	RunAgentCreate("Alice", server.URL)
	RunAgentCreate("Bob", server.URL)

	// 测试 list
	err := RunAgentList()
	require.NoError(t, err)
}

func TestRunAgentList_Empty(t *testing.T) {
	setupTestEnv(t)

	err := RunAgentList()
	require.NoError(t, err)
}

func TestRunAgentDelete(t *testing.T) {
	setupTestEnv(t)

	// 先创建
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(client.RegisterResponse{AgentID: "agent-001"})
	}))
	defer server.Close()

	RunAgentCreate("Alice", server.URL)

	// 删除
	err := RunAgentDelete("Alice")
	require.NoError(t, err)

	// 验证删除
	keyPath := filepath.Join(os.Getenv("HOME"), ".at-cli", "agents", "Alice.pem")
	_, err = os.Stat(keyPath)
	assert.True(t, os.IsNotExist(err))
}

func TestRunAgentDelete_NotFound(t *testing.T) {
	setupTestEnv(t)

	err := RunAgentDelete("NotExist")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRunAgentExport(t *testing.T) {
	setupTestEnv(t)

	// 先创建
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(client.RegisterResponse{AgentID: "agent-001"})
	}))
	defer server.Close()

	RunAgentCreate("Alice", server.URL)

	// 导出到文件
	tmpFile := filepath.Join(t.TempDir(), "alice-key.pem")
	err := RunAgentExport("Alice", tmpFile)
	require.NoError(t, err)

	// 验证文件内容
	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "BEGIN ED25519 PRIVATE KEY")
}

func TestRunAgentExport_NotFound(t *testing.T) {
	setupTestEnv(t)

	err := RunAgentExport("NotExist", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
