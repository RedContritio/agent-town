package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/RedContritio/agent-town/cli/internal/auth"
	"github.com/RedContritio/agent-town/cli/internal/builder"
	"github.com/RedContritio/agent-town/cli/internal/client"
	cmdpkg "github.com/RedContritio/agent-town/cli/internal/cmd"
	"github.com/RedContritio/agent-town/cli/internal/config"
	"github.com/RedContritio/agent-town/cli/internal/executor"
	"github.com/RedContritio/agent-town/cli/internal/keystore"
	"github.com/RedContritio/agent-town/cli/internal/registry"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestEnv(t *testing.T) string {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })
	return tmpDir
}

// TestFullWorkflow 测试完整流程：create → auth → sync cmd → async cmd
func TestFullWorkflow(t *testing.T) {
	setupTestEnv(t)

	// 统计 API 调用
	var registerCalled, challengeCalled, tokenCalled, commandsCalled int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/register":
			registerCalled++
			var req client.RegisterRequest
			json.NewDecoder(r.Body).Decode(&req)
			json.NewEncoder(w).Encode(client.RegisterResponse{
				AgentID: "agent-alice-001",
			})

		case "/api/v1/auth/challenge":
			challengeCalled++
			json.NewEncoder(w).Encode(client.ChallengeResponse{
				ChallengeID: "chal-001",
				Nonce:       []byte("test-nonce"),
				ExpiresAt:   9999999999,
			})

		case "/api/v1/auth/token":
			tokenCalled++
			json.NewEncoder(w).Encode(client.TokenResponse{
				Token:     "test-jwt-token",
				ExpiresAt: 9999999999,
				AgentID:   "agent-alice-001",
			})

		case "/api/v1/commands":
			commandsCalled++
			json.NewEncoder(w).Encode(client.CommandsResponse{
				Commands: []client.CommandDefinition{
					{
						Name:        "status",
						Type:        "sync",
						Endpoint:    "/api/v1/status",
						Method:      "GET",
						Description: "Get agent status",
					},
					{
						Name:        "move",
						Type:        "async",
						Endpoint:    "/api/v1/stack",
						Method:      "POST",
						Params: []client.CommandParam{
							{Name: "dx", Type: "int", Required: true},
							{Name: "dy", Type: "int", Required: true},
						},
						Description: "Move agent",
					},
				},
			})

		case "/api/v1/status":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"hp":      100,
				"stamina": 80,
			})

		case "/api/v1/stack":
			json.NewEncoder(w).Encode(client.TaskResponse{
				TaskID:        "Alice-001",
				EstimatedTime: "6s",
			})
		}
	}))
	defer server.Close()

	// Step 1: Create Agent
	err := cmdpkg.RunAgentCreate("Alice", server.URL)
	require.NoError(t, err)
	assert.Equal(t, 1, registerCalled, "register should be called once")

	// Step 2: Load config and key
	cfg, err := config.Load()
	require.NoError(t, err)

	agent, ok := cfg.GetAgent("Alice")
	require.True(t, ok)
	assert.Equal(t, "agent-alice-001", agent.AgentID)

	key, err := keystore.Load("Alice")
	require.NoError(t, err)
	assert.NotNil(t, key)

	// Step 3: Authenticate
	c := client.New(server.URL)
	authMgr := auth.NewManager(cfg)
	token, err := authMgr.EnsureToken("Alice", c)
	require.NoError(t, err)
	assert.Equal(t, "test-jwt-token", token)
	assert.Equal(t, 1, challengeCalled, "challenge should be called once")
	assert.Equal(t, 1, tokenCalled, "token should be called once")

	// Step 4: Load commands
	reg := registry.New()
	err = reg.LoadFromServer(c)
	require.NoError(t, err)
	assert.Equal(t, 1, commandsCalled, "commands should be called once")
	assert.Len(t, reg.All(), 2)

	// Step 5: Execute sync command (status)
	exec := executor.New(c)
	statusDef, ok := reg.Get("status")
	require.True(t, ok)

	result, err := exec.ExecuteSync(statusDef, nil)
	require.NoError(t, err)
	assert.Equal(t, float64(100), result["hp"])

	// Step 6: Execute async command (move)
	moveDef, ok := reg.Get("move")
	require.True(t, ok)

	resp, err := exec.ExecuteAsync(moveDef, map[string]interface{}{"dx": 3, "dy": 0})
	require.NoError(t, err)
	assert.Equal(t, "Alice-001", resp.TaskID)

	// Step 7: Verify token caching (second call should not trigger auth)
	challengeCalled = 0
	tokenCalled = 0
	token2, err := authMgr.EnsureToken("Alice", c)
	require.NoError(t, err)
	assert.Equal(t, token, token2)
	assert.Equal(t, 0, challengeCalled, "challenge should not be called (cached)")
	assert.Equal(t, 0, tokenCalled, "token should not be called (cached)")
}

// TestAgentCommands 测试本地 agent 命令
func TestAgentCommands(t *testing.T) {
	setupTestEnv(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(client.RegisterResponse{AgentID: "agent-001"})
	}))
	defer server.Close()

	// Create
	err := cmdpkg.RunAgentCreate("Bob", server.URL)
	require.NoError(t, err)

	// List
	err = cmdpkg.RunAgentList()
	require.NoError(t, err)

	// Export
	tmpFile := filepath.Join(t.TempDir(), "bob.pem")
	err = cmdpkg.RunAgentExport("Bob", tmpFile)
	require.NoError(t, err)

	data, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	assert.Contains(t, string(data), "ED25519 PRIVATE KEY")

	// Delete
	err = cmdpkg.RunAgentDelete("Bob")
	require.NoError(t, err)

	// Verify deleted
	_, err = keystore.Load("Bob")
	assert.Error(t, err)
}

// TestDynamicCommandBuilder 测试动态命令构建
func TestDynamicCommandBuilder(t *testing.T) {
	setupTestEnv(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/commands":
			json.NewEncoder(w).Encode(client.CommandsResponse{
				Commands: []client.CommandDefinition{
					{Name: "status", Type: "sync", Endpoint: "/status", Method: "GET"},
					{Name: "look", Type: "sync", Endpoint: "/look", Method: "GET"},
				},
			})
		case "/status":
			json.NewEncoder(w).Encode(map[string]interface{}{"result": "ok"})
		case "/look":
			json.NewEncoder(w).Encode(map[string]interface{}{"seen": "tree"})
		}
	}))
	defer server.Close()

	c := client.New(server.URL)
	reg := registry.New()
	err := reg.LoadFromServer(c)
	require.NoError(t, err)

	exec := executor.New(c)
	b := builder.New(reg, exec)

	// Build root with local commands
	localCmds := []*cobra.Command{cmdpkg.AgentCmd()}
	root := b.BuildRoot(localCmds)

	// Verify commands exist
	cmds := root.Commands()
	assert.GreaterOrEqual(t, len(cmds), 3) // agent, status, look

	// Find status command
	var statusCmd *cobra.Command
	for _, cmd := range cmds {
		if cmd.Name() == "status" {
			statusCmd = cmd
			break
		}
	}
	assert.NotNil(t, statusCmd)
	assert.Equal(t, "status", statusCmd.Use)
}
