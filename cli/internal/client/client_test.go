package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	c := New("http://localhost:8080")
	assert.Equal(t, "http://localhost:8080", c.BaseURL)
	assert.NotNil(t, c.HTTPClient)
}

func TestSetToken(t *testing.T) {
	c := New("http://localhost:8080")
	c.SetToken("test-token")
	assert.Equal(t, "test-token", c.token)
}

func TestDo_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/test", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// 验证请求体
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "value", body["key"])

		// 返回响应
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"result": "ok"})
	}))
	defer server.Close()

	c := New(server.URL)
	reqBody := map[string]string{"key": "value"}
	var result map[string]string

	err := c.Do("POST", "/test", reqBody, &result)
	require.NoError(t, err)
	assert.Equal(t, "ok", result["result"])
}

func TestDo_WithToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer my-token", authHeader)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := New(server.URL)
	c.SetToken("my-token")

	err := c.Get("/test", nil)
	require.NoError(t, err)
}

func TestDo_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer server.Close()

	c := New(server.URL)
	err := c.Get("/test", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 400")
}

func TestDo_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	c := New(server.URL)
	var result map[string]string
	err := c.Get("/test", &result)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}
