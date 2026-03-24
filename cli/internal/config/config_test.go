package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_NotExist(t *testing.T) {
	// 使用临时目录隔离测试
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	cfg, err := Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Agents)
	assert.Empty(t, cfg.Agents)
}

func TestLoad_Exist(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	// 创建测试配置文件
	configContent := `
agents:
  Alice:
    agent_id: agent-abc123
    server_url: http://localhost:8080
    token: eyJhbGciOiJIUzI1NiIs
    token_expires: 1710000000
  Bob:
    agent_id: agent-def456
    server_url: http://localhost:8081
current_agent: Alice
`
	configDir := filepath.Join(tmpDir, ".at-cli")
	os.MkdirAll(configDir, 0700)
	os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configContent), 0600)

	cfg, err := Load()
	require.NoError(t, err)

	// 验证加载的数据
	assert.Equal(t, "Alice", cfg.CurrentAgent)
	
	alice, ok := cfg.GetAgent("Alice")
	require.True(t, ok)
	assert.Equal(t, "agent-abc123", alice.AgentID)
	assert.Equal(t, "http://localhost:8080", alice.ServerURL)
	assert.Equal(t, "eyJhbGciOiJIUzI1NiIs", alice.Token)
	assert.Equal(t, int64(1710000000), alice.TokenExpires)

	bob, ok := cfg.GetAgent("Bob")
	require.True(t, ok)
	assert.Equal(t, "agent-def456", bob.AgentID)
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	cfg := &Config{
		Agents: map[string]*AgentConfig{
			"TestAgent": {
				AgentID:   "agent-test-001",
				ServerURL: "http://test.com",
				Token:     "test-token",
			},
		},
		CurrentAgent: "TestAgent",
	}

	err := cfg.Save()
	require.NoError(t, err)

	// 验证文件存在
	configFile := filepath.Join(tmpDir, ".at-cli", "config.yaml")
	_, err = os.Stat(configFile)
	require.NoError(t, err)

	// 重新加载验证
	loaded, err := Load()
	require.NoError(t, err)
	
	agent, ok := loaded.GetAgent("TestAgent")
	require.True(t, ok)
	assert.Equal(t, "agent-test-001", agent.AgentID)
	assert.Equal(t, "http://test.com", agent.ServerURL)
	assert.Equal(t, "test-token", agent.Token)
}

func TestConfig_SaveToken(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	cfg := &Config{
		Agents: map[string]*AgentConfig{
			"Alice": {
				AgentID:   "agent-001",
				ServerURL: "http://localhost:8080",
			},
		},
	}

	err := cfg.SaveToken("Alice", "new-token", 1711111111)
	require.NoError(t, err)

	token, expires, ok := cfg.GetToken("Alice")
	require.True(t, ok)
	assert.Equal(t, "new-token", token)
	assert.Equal(t, int64(1711111111), expires)
}

func TestConfig_SaveToken_AgentNotFound(t *testing.T) {
	cfg := &Config{
		Agents: make(map[string]*AgentConfig),
	}

	err := cfg.SaveToken("NotExist", "token", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestConfig_DeleteAgent(t *testing.T) {
	cfg := &Config{
		Agents: map[string]*AgentConfig{
			"Alice": {AgentID: "agent-001"},
			"Bob":   {AgentID: "agent-002"},
		},
		CurrentAgent: "Alice",
	}

	cfg.DeleteAgent("Alice")

	_, ok := cfg.GetAgent("Alice")
	assert.False(t, ok)

	// CurrentAgent 应该被清空
	assert.Empty(t, cfg.CurrentAgent)
}
