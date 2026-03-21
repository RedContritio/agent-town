package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/RedContritio/agent-town/cli/internal/client"
	"github.com/RedContritio/agent-town/cli/internal/config"
)

func setupTestRegistry(t *testing.T) (*Registry, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/commands" {
			http.NotFound(w, r)
			return
		}

		resp := CommandsResponse{
			Commands: []CommandDefinition{
				{Name: "test", Type: "sync", Endpoint: "/test", Method: "GET"},
			},
			Version: "1.0",
		}
		json.NewEncoder(w).Encode(resp)
	}))

	tmpDir := t.TempDir()
	cfg := &config.Config{
		ServerURL: server.URL,
		DataDir:   tmpDir,
	}

	cli := client.NewClient(server.URL)
	registry := NewRegistry(cli, cfg)

	return registry, server
}

func TestRegistry_Discover(t *testing.T) {
	registry, server := setupTestRegistry(t)
	defer server.Close()

	err := registry.Discover(false)
	if err != nil {
		t.Fatalf("发现命令失败: %v", err)
	}

	cmd, found := registry.GetCommand("test")
	if !found {
		t.Error("未找到 test 命令")
	}
	if cmd.Type != "sync" {
		t.Errorf("期望 type=sync, 得到 %s", cmd.Type)
	}
}

func TestRegistry_Cache(t *testing.T) {
	registry, server := setupTestRegistry(t)
	defer server.Close()

	// 第一次发现
	err := registry.Discover(false)
	if err != nil {
		t.Fatalf("发现命令失败: %v", err)
	}

	// 检查缓存文件是否存在
	cachePath := registry.config.GetCommandsCachePath()
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		t.Error("缓存文件未创建")
	}

	// 创建新的 registry，从缓存加载
	cli := client.NewClient(server.URL)
	newRegistry := NewRegistry(cli, registry.config)

	// 不强制刷新，应该从缓存加载
	err = newRegistry.Discover(false)
	if err != nil {
		t.Fatalf("从缓存加载失败: %v", err)
	}

	cmd, found := newRegistry.GetCommand("test")
	if !found {
		t.Error("从缓存未找到 test 命令")
	}
	if cmd.Name != "test" {
		t.Errorf("命令名称不匹配: %s", cmd.Name)
	}
}

func TestRegistry_ClearCache(t *testing.T) {
	registry, server := setupTestRegistry(t)
	defer server.Close()

	// 发现并缓存
	registry.Discover(false)

	// 清除缓存
	err := registry.ClearCache()
	if err != nil {
		t.Fatalf("清除缓存失败: %v", err)
	}

	// 验证缓存文件不存在
	cachePath := registry.config.GetCommandsCachePath()
	if _, err := os.Stat(cachePath); !os.IsNotExist(err) {
		t.Error("缓存文件应该被删除")
	}
}

func TestRegistry_Refresh(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := CommandsResponse{
			Commands: []CommandDefinition{
				{Name: "cmd" + string(rune('0'+callCount)), Type: "sync"},
			},
			Version: "1.0",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	tmpDir := t.TempDir()
	cfg := &config.Config{
		ServerURL: server.URL,
		DataDir:   tmpDir,
	}

	cli := client.NewClient(server.URL)
	registry := NewRegistry(cli, cfg)

	// 第一次发现
	registry.Discover(false)
	if callCount != 1 {
		t.Errorf("期望服务器被调用 1 次，实际 %d 次", callCount)
	}

	// 强制刷新
	registry.Discover(true)
	if callCount != 2 {
		t.Errorf("期望服务器被调用 2 次，实际 %d 次", callCount)
	}
}

func TestRegistry_GetCommands(t *testing.T) {
	registry, server := setupTestRegistry(t)
	defer server.Close()

	registry.Discover(false)

	cmds := registry.GetCommands()
	if len(cmds) != 1 {
		t.Errorf("期望 1 个命令，实际 %d 个", len(cmds))
	}
}
