package executor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RedContritio/agent-town/cli/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteSync(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/status", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"hp":      100,
			"stamina": 80,
		})
	}))
	defer server.Close()

	c := client.New(server.URL)
	e := New(c)

	def := &client.CommandDefinition{
		Name:     "status",
		Type:     "sync",
		Endpoint: "/api/v1/status",
		Method:   "GET",
	}

	result, err := e.ExecuteSync(def, nil)
	require.NoError(t, err)
	assert.Equal(t, float64(100), result["hp"])      // JSON 数字解析为 float64
	assert.Equal(t, float64(80), result["stamina"])
}

func TestExecuteSync_WithParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/scan", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, float64(10), body["x"])
		assert.Equal(t, float64(20), body["y"])

		json.NewEncoder(w).Encode(map[string]interface{}{
			"resources": []string{"tree-001"},
		})
	}))
	defer server.Close()

	c := client.New(server.URL)
	e := New(c)

	def := &client.CommandDefinition{
		Name:     "scan",
		Type:     "sync",
		Endpoint: "/api/v1/scan",
		Method:   "POST",
		Params: []client.CommandParam{
			{Name: "x", Type: "int", Required: true},
			{Name: "y", Type: "int", Required: true},
		},
	}

	params := map[string]interface{}{"x": 10, "y": 20}
	result, err := e.ExecuteSync(def, params)
	require.NoError(t, err)
	assert.NotNil(t, result["resources"])
}

func TestExecuteAsync(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/stack", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, float64(3), body["dx"])
		assert.Equal(t, float64(0), body["dy"])

		json.NewEncoder(w).Encode(client.TaskResponse{
			TaskID:        "Alice-001",
			EstimatedTime: "6s",
		})
	}))
	defer server.Close()

	c := client.New(server.URL)
	e := New(c)

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

	params := map[string]interface{}{"dx": 3, "dy": 0}
	result, err := e.ExecuteAsync(def, params)
	require.NoError(t, err)
	assert.Equal(t, "Alice-001", result.TaskID)
	assert.Equal(t, "6s", result.EstimatedTime)
}

func TestExecuteSync_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid token"}`))
	}))
	defer server.Close()

	c := client.New(server.URL)
	e := New(c)

	def := &client.CommandDefinition{
		Name:     "status",
		Endpoint: "/api/v1/status",
		Method:   "GET",
	}

	_, err := e.ExecuteSync(def, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "execute status")
}

func TestFormatAsyncResult(t *testing.T) {
	def := &client.CommandDefinition{Name: "move"}
	params := map[string]interface{}{"dx": 3, "dy": 0}
	resp := &client.TaskResponse{
		TaskID:        "Alice-001",
		EstimatedTime: "6s",
	}

	result := FormatAsyncResult(def, params, resp)
	assert.Equal(t, "Alice-001: move 3 0 (est: 6s)", result)
}

func TestFormatAsyncResult_NoEstTime(t *testing.T) {
	def := &client.CommandDefinition{Name: "harvest"}
	params := map[string]interface{}{"resource_id": "tree-001"}
	resp := &client.TaskResponse{
		TaskID: "Alice-002",
	}

	result := FormatAsyncResult(def, params, resp)
	assert.Equal(t, "Alice-002: harvest tree-001", result)
}

func TestFormatAsyncResult_NoParams(t *testing.T) {
	def := &client.CommandDefinition{Name: "look"}
	resp := &client.TaskResponse{
		TaskID:        "Alice-003",
		EstimatedTime: "1s",
	}

	result := FormatAsyncResult(def, nil, resp)
	assert.Equal(t, "Alice-003: look (est: 1s)", result)
}

func TestBuildRequestBody(t *testing.T) {
	// 有参数
	body := buildRequestBody(map[string]interface{}{"x": 1, "y": 2})
	assert.Equal(t, map[string]interface{}{"x": 1, "y": 2}, body)

	// 无参数
	body = buildRequestBody(nil)
	assert.Nil(t, body)

	// 空 map
	body = buildRequestBody(map[string]interface{}{})
	assert.Nil(t, body)
}
