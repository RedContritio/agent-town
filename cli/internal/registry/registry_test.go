package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RedContritio/agent-town/cli/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	r := New()
	assert.NotNil(t, r)
	assert.NotNil(t, r.commands)
	assert.NotNil(t, r.byType)
}

func TestLoadFromServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/commands", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		json.NewEncoder(w).Encode(client.CommandsResponse{
			Commands: []client.CommandDefinition{
				{
					Name:        "status",
					Type:        "sync",
					Endpoint:    "/api/v1/status",
					Method:      "GET",
					Description: "Get status",
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
			Version: "1.0",
		})
	}))
	defer server.Close()

	r := New()
	c := client.New(server.URL)
	err := r.LoadFromServer(c)

	require.NoError(t, err)
	assert.Len(t, r.commands, 2)

	// 验证 status
	status, ok := r.Get("status")
	require.True(t, ok)
	assert.Equal(t, "sync", status.Type)
	assert.True(t, r.IsSync("status"))
	assert.False(t, r.IsAsync("status"))

	// 验证 move
	move, ok := r.Get("move")
	require.True(t, ok)
	assert.Equal(t, "async", move.Type)
	assert.Equal(t, 2, len(move.Params))
	assert.True(t, r.IsAsync("move"))
	assert.False(t, r.IsSync("move"))
}

func TestLoadFromServer_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	r := New()
	c := client.New(server.URL)
	err := r.LoadFromServer(c)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fetch commands")
}

func TestAll(t *testing.T) {
	r := New()
	r.commands["status"] = &client.CommandDefinition{Name: "status", Type: "sync"}
	r.commands["move"] = &client.CommandDefinition{Name: "move", Type: "async"}

	all := r.All()
	assert.Len(t, all, 2)
}

func TestSyncAsyncCommands(t *testing.T) {
	r := New()
	r.byType["sync"] = []*client.CommandDefinition{
		{Name: "status", Type: "sync"},
		{Name: "look", Type: "sync"},
	}
	r.byType["async"] = []*client.CommandDefinition{
		{Name: "move", Type: "async"},
	}

	syncs := r.SyncCommands()
	asyncs := r.AsyncCommands()

	assert.Len(t, syncs, 2)
	assert.Len(t, asyncs, 1)
}

func TestGet_NotFound(t *testing.T) {
	r := New()
	cmd, ok := r.Get("notexist")
	assert.False(t, ok)
	assert.Nil(t, cmd)

	assert.False(t, r.IsSync("notexist"))
	assert.False(t, r.IsAsync("notexist"))
}
