package builder

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RedContritio/agent-town/cli/internal/client"
	"github.com/RedContritio/agent-town/cli/internal/executor"
	"github.com/RedContritio/agent-town/cli/internal/registry"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCommand(t *testing.T) {
	r := registry.New()
	e := &executor.Executor{}
	b := New(r, e)

	def := &client.CommandDefinition{
		Name:        "status",
		Type:        "sync",
		Description: "Get agent status",
		Params:      []client.CommandParam{},
	}

	cmd := b.BuildCommand(def)
	assert.Equal(t, "status", cmd.Use)
	assert.Equal(t, "Get agent status", cmd.Short)
	assert.NotNil(t, cmd.RunE)
}

func TestBuildCommand_WithParams(t *testing.T) {
	r := registry.New()
	e := &executor.Executor{}
	b := New(r, e)

	def := &client.CommandDefinition{
		Name:        "move",
		Type:        "async",
		Description: "Move agent",
		Params: []client.CommandParam{
			{Name: "dx", Type: "int", Required: true},
			{Name: "dy", Type: "int", Required: true},
		},
	}

	cmd := b.BuildCommand(def)
	assert.Equal(t, "move <dx> <dy>", cmd.Use)
}

func TestBuildRoot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(client.CommandsResponse{
			Commands: []client.CommandDefinition{
				{Name: "status", Type: "sync", Endpoint: "/status", Method: "GET"},
				{Name: "move", Type: "async", Endpoint: "/move", Method: "POST"},
			},
		})
	}))
	defer server.Close()

	// 准备 registry
	r := registry.New()
	c := client.New(server.URL)
	err := r.LoadFromServer(c)
	require.NoError(t, err)

	// 本地命令
	localCmd := &cobra.Command{Use: "agent"}

	e := executor.New(c)
	b := New(r, e)

	root := b.BuildRoot([]*cobra.Command{localCmd})

	// 验证有命令
	cmds := root.Commands()
	assert.GreaterOrEqual(t, len(cmds), 1)

	// 验证本地命令存在
	found := false
	for _, cmd := range cmds {
		if cmd.Name() == "agent" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestIsConflict(t *testing.T) {
	localCmds := []*cobra.Command{
		{Use: "agent"},
		{Use: "help"},
	}

	assert.True(t, isConflict("agent", localCmds))
	assert.True(t, isConflict("help", localCmds))
	assert.False(t, isConflict("status", localCmds))
	assert.False(t, isConflict("move", localCmds))
}

func TestRunCommand_Sync(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"hp": 100,
		})
	}))
	defer server.Close()

	c := client.New(server.URL)
	e := executor.New(c)
	b := New(nil, e)

	def := &client.CommandDefinition{
		Name:     "status",
		Type:     "sync",
		Endpoint: "/api/v1/status",
		Method:   "GET",
	}

	err := b.runCommand(def, []string{})
	require.NoError(t, err)
}

func TestRunCommand_Async(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(client.TaskResponse{
			TaskID:        "Alice-001",
			EstimatedTime: "6s",
		})
	}))
	defer server.Close()

	c := client.New(server.URL)
	e := executor.New(c)
	b := New(nil, e)

	def := &client.CommandDefinition{
		Name:     "move",
		Type:     "async",
		Endpoint: "/api/v1/stack",
		Method:   "POST",
		Params: []client.CommandParam{
			{Name: "dx", Type: "int", Required: true},
			{Name: "dy", Type: "int", Required: true},
		},
	}

	err := b.runCommand(def, []string{"3", "0"})
	require.NoError(t, err)
}

func TestRunCommand_ParseError(t *testing.T) {
	e := &executor.Executor{}
	b := New(nil, e)

	def := &client.CommandDefinition{
		Name: "move",
		Type: "async",
		Params: []client.CommandParam{
			{Name: "dx", Type: "int", Required: true},
		},
	}

	err := b.runCommand(def, []string{"not-a-number"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse args")
}

func TestBuildHelpString(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(client.CommandsResponse{
			Commands: []client.CommandDefinition{
				{Name: "status", Type: "sync", Description: "Get status"},
				{Name: "move", Type: "async", Description: "Move agent"},
			},
		})
	}))
	defer server.Close()

	r := registry.New()
	c := client.New(server.URL)
	r.LoadFromServer(c)

	help := BuildHelpString(r)
	assert.Contains(t, help, "[Sync Commands]")
	assert.Contains(t, help, "[Async Commands]")
	assert.Contains(t, help, "status")
	assert.Contains(t, help, "move")
}
