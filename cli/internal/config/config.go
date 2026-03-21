// Package config CLI 配置
package config

import (
	"os"
	"path/filepath"
)

// Config CLI 配置
type Config struct {
	ServerURL    string // 服务器 URL
	AgentName    string // 当前 Agent 名称
	Token        string // 认证 Token
	DataDir      string // 数据目录
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		ServerURL: "http://localhost:8080",
		DataDir:   filepath.Join(homeDir, ".at-cli"),
	}
}

// GetAgentsDir 获取 Agent 密钥存储目录
func (c *Config) GetAgentsDir() string {
	return filepath.Join(c.DataDir, "agents")
}

// GetCacheDir 获取缓存目录
func (c *Config) GetCacheDir() string {
	return filepath.Join(c.DataDir, "cache")
}

// GetCommandsCachePath 获取命令缓存路径
func (c *Config) GetCommandsCachePath() string {
	return filepath.Join(c.GetCacheDir(), "commands.json")
}

// EnsureDirs 确保目录存在
func (c *Config) EnsureDirs() error {
	dirs := []string{c.DataDir, c.GetAgentsDir(), c.GetCacheDir()}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}
