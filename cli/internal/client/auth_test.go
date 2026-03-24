package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthAPI_Register(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/auth/register", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req RegisterRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "ed25519:abc123", req.PublicKey)
		assert.Equal(t, "Alice", req.Name)

		json.NewEncoder(w).Encode(RegisterResponse{
			AgentID: "agent-abc-123",
		})
	}))
	defer server.Close()

	c := New(server.URL)
	resp, err := c.Auth().Register(&RegisterRequest{
		PublicKey: "ed25519:abc123",
		Name:      "Alice",
	})

	require.NoError(t, err)
	assert.Equal(t, "agent-abc-123", resp.AgentID)
}

func TestAuthAPI_Register_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error": "agent already exists"}`))
	}))
	defer server.Close()

	c := New(server.URL)
	resp, err := c.Auth().Register(&RegisterRequest{
		PublicKey: "ed25519:abc123",
		Name:      "Alice",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "HTTP 409")
}

func TestAuthAPI_Challenge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/auth/challenge", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req ChallengeRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "ed25519:pubkey123", req.PublicKey)

		json.NewEncoder(w).Encode(ChallengeResponse{
			ChallengeID: "chal-001",
			Nonce:       []byte("test-nonce-123"),
			ExpiresAt:   time.Now().Add(5 * time.Minute).Unix(),
		})
	}))
	defer server.Close()

	c := New(server.URL)
	resp, err := c.Auth().Challenge("ed25519:pubkey123")

	require.NoError(t, err)
	assert.Equal(t, "chal-001", resp.ChallengeID)
	assert.Equal(t, []byte("test-nonce-123"), resp.Nonce)
	assert.Greater(t, resp.ExpiresAt, time.Now().Unix())
}

func TestAuthAPI_Token(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/auth/token", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var req TokenRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "chal-001", req.ChallengeID)
		assert.Equal(t, []byte("signature123"), req.Signature)

		json.NewEncoder(w).Encode(TokenResponse{
			Token:     "eyJhbGciOiJIUzI1NiIs",
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			AgentID:   "agent-abc-123",
		})
	}))
	defer server.Close()

	c := New(server.URL)
	resp, err := c.Auth().Token(&TokenRequest{
		ChallengeID: "chal-001",
		Signature:   []byte("signature123"),
	})

	require.NoError(t, err)
	assert.Equal(t, "eyJhbGciOiJIUzI1NiIs", resp.Token)
	assert.Equal(t, "agent-abc-123", resp.AgentID)
}

func TestClient_Commands(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/commands", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		json.NewEncoder(w).Encode(CommandsResponse{
			Commands: []CommandDefinition{
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
					Params: []CommandParam{
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

	c := New(server.URL)
	resp, err := c.Commands()

	require.NoError(t, err)
	assert.Equal(t, "1.0", resp.Version)
	assert.Len(t, resp.Commands, 2)
	assert.Equal(t, "status", resp.Commands[0].Name)
	assert.Equal(t, "sync", resp.Commands[0].Type)
	assert.Equal(t, "move", resp.Commands[1].Name)
	assert.Equal(t, "async", resp.Commands[1].Type)
	assert.Len(t, resp.Commands[1].Params, 2)
}
