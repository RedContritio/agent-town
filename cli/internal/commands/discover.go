// Package commands 命令发现与执行
package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/RedContritio/agent-town/cli/internal/client"
	"github.com/RedContritio/agent-town/cli/internal/config"
)

const (
	// CacheTTL 缓存有效期
	CacheTTL = 1 * time.Hour
)

// Registry 命令注册表
type Registry struct {
	commands  []CommandDefinition
	version   string
	cachedAt  time.Time
	client    *client.Client
	config    *config.Config
}

// NewRegistry 创建命令注册表
func NewRegistry(cli *client.Client, cfg *config.Config) *Registry {
	return &Registry{
		client: cli,
		config: cfg,
	}
}

// Discover 发现并缓存命令
func (r *Registry) Discover(forceRefresh bool) error {
	// 检查缓存
	if !forceRefresh && r.isCacheValid() {
		if err := r.loadFromCache(); err == nil {
			return nil
		}
		// 缓存加载失败，继续从服务器获取
	}

	// 从服务器获取
	var resp CommandsResponse
	if err := r.client.GetJSON("/api/v1/commands", &resp); err != nil {
		return fmt.Errorf("获取命令列表失败: %w", err)
	}

	r.commands = resp.Commands
	r.version = resp.Version
	r.cachedAt = time.Now()

	// 保存到缓存
	if err := r.saveToCache(); err != nil {
		// 缓存保存失败不影响使用
		fmt.Fprintf(os.Stderr, "警告: 保存命令缓存失败: %v\n", err)
	}

	return nil
}

// GetCommand 获取命令定义
func (r *Registry) GetCommand(name string) (*CommandDefinition, bool) {
	for _, cmd := range r.commands {
		if cmd.Name == name {
			return &cmd, true
		}
	}
	return nil, false
}

// GetCommands 获取所有命令
func (r *Registry) GetCommands() []CommandDefinition {
	return r.commands
}

// isCacheValid 检查缓存是否有效
func (r *Registry) isCacheValid() bool {
	if r.cachedAt.IsZero() {
		// 尝试从文件加载
		if err := r.loadFromCache(); err != nil {
			return false
		}
	}
	return time.Since(r.cachedAt) < CacheTTL
}

// loadFromCache 从缓存加载
func (r *Registry) loadFromCache() error {
	cachePath := r.config.GetCommandsCachePath()
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return err
	}

	var cache struct {
		Commands []CommandDefinition `json:"commands"`
		Version  string              `json:"version"`
		CachedAt time.Time           `json:"cached_at"`
	}

	if err := json.Unmarshal(data, &cache); err != nil {
		return err
	}

	r.commands = cache.Commands
	r.version = cache.Version
	r.cachedAt = cache.CachedAt
	return nil
}

// saveToCache 保存到缓存
func (r *Registry) saveToCache() error {
	if err := r.config.EnsureDirs(); err != nil {
		return err
	}

	cachePath := r.config.GetCommandsCachePath()
	cache := struct {
		Commands []CommandDefinition `json:"commands"`
		Version  string              `json:"version"`
		CachedAt time.Time           `json:"cached_at"`
	}{
		Commands: r.commands,
		Version:  r.version,
		CachedAt: r.cachedAt,
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cachePath, data, 0644)
}

// ClearCache 清除缓存
func (r *Registry) ClearCache() error {
	r.cachedAt = time.Time{}
	r.commands = nil
	cachePath := r.config.GetCommandsCachePath()
	if err := os.Remove(cachePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
