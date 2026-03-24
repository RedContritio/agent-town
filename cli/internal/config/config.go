package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config 是 CLI 的主配置
// 决策：使用 YAML 格式，简单可读。可选 JSON（更快）或 SQLite（复杂查询）
type Config struct {
	Agents       map[string]*AgentConfig `yaml:"agents"`
	CurrentAgent string                  `yaml:"current_agent,omitempty"`
}

// AgentConfig 存储单个 Agent 的配置
type AgentConfig struct {
	AgentID      string `yaml:"agent_id"`
	ServerURL    string `yaml:"server_url"`
	Token        string `yaml:"token,omitempty"`
	TokenExpires int64  `yaml:"token_expires,omitempty"` // Unix timestamp
}

// 配置目录和文件路径
func configDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".at-cli")
}

func configPath() string {
	return filepath.Join(configDir(), "config.yaml")
}

// Load 从配置文件加载配置
// 如果不存在则返回空配置（不报错，由调用者决定是否创建）
func Load() (*Config, error) {
	path := configPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Agents: make(map[string]*AgentConfig),
			}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Agents == nil {
		cfg.Agents = make(map[string]*AgentConfig)
	}

	return &cfg, nil
}

// Save 保存配置到文件
func (c *Config) Save() error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	path := configPath()
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// GetAgent 获取指定 Agent 的配置
func (c *Config) GetAgent(name string) (*AgentConfig, bool) {
	agent, ok := c.Agents[name]
	return agent, ok
}

// SetAgent 设置/更新 Agent 配置
func (c *Config) SetAgent(name string, agent *AgentConfig) {
	c.Agents[name] = agent
}

// DeleteAgent 删除 Agent 配置
func (c *Config) DeleteAgent(name string) {
	delete(c.Agents, name)
	if c.CurrentAgent == name {
		c.CurrentAgent = ""
	}
}

// SaveToken 保存 Agent 的 Token
func (c *Config) SaveToken(agentName, token string, expires int64) error {
	agent, ok := c.Agents[agentName]
	if !ok {
		return fmt.Errorf("agent %q not found", agentName)
	}
	agent.Token = token
	agent.TokenExpires = expires
	return c.Save()
}

// GetToken 获取 Agent 的 Token
func (c *Config) GetToken(agentName string) (string, int64, bool) {
	agent, ok := c.Agents[agentName]
	if !ok {
		return "", 0, false
	}
	return agent.Token, agent.TokenExpires, agent.Token != ""
}
