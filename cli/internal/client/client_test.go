package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("期望 GET 方法, 得到 %s", r.Method)
		}
		if r.URL.Path != "/test" {
			t.Errorf("期望路径 /test, 得到 %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	resp, err := client.Get("/test")
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("期望状态码 200, 得到 %d", resp.StatusCode)
	}
}

func TestClient_PostJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("期望 POST 方法, 得到 %s", r.Method)
		}

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["key"] != "value" {
			t.Errorf("期望 body.key=value, 得到 %v", body["key"])
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"result": "created"})
	}))
	defer server.Close()

	client := NewClient(server.URL)

	var result map[string]string
	err := client.PostJSON("/test", map[string]string{"key": "value"}, &result)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}

	if result["result"] != "created" {
		t.Errorf("期望 result=created, 得到 %s", result["result"])
	}
}

func TestClient_SetToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token-123" {
			t.Errorf("期望 Authorization: Bearer test-token-123, 得到 %s", auth)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	client.SetToken("test-token-123")

	_, err := client.Get("/test")
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
}

func TestClient_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	var result interface{}
	err := client.GetJSON("/test", &result)
	if err == nil {
		t.Error("期望错误，但没有")
	}
}
