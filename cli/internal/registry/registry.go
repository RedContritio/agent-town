package registry

import (
	"fmt"

	"github.com/RedContritio/agent-town/cli/internal/client"
)

// Registry 命令注册表
// 决策：内存存储，CLI 每次启动时从服务端拉取
// 可选：本地缓存（减少启动时间，需要处理缓存失效）
type Registry struct {
	commands map[string]*client.CommandDefinition
	byType   map[string][]*client.CommandDefinition // sync/async 分组
}

// New 创建新的注册表
func New() *Registry {
	return &Registry{
		commands: make(map[string]*client.CommandDefinition),
		byType:   make(map[string][]*client.CommandDefinition),
	}
}

// LoadFromServer 从服务端加载命令定义
func (r *Registry) LoadFromServer(c *client.Client) error {
	resp, err := c.Commands()
	if err != nil {
		return fmt.Errorf("fetch commands: %w", err)
	}

	for i := range resp.Commands {
		cmd := &resp.Commands[i]
		r.commands[cmd.Name] = cmd
		r.byType[cmd.Type] = append(r.byType[cmd.Type], cmd)
	}

	return nil
}

// Get 获取命令定义
func (r *Registry) Get(name string) (*client.CommandDefinition, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}

// IsSync 是否是同步命令
func (r *Registry) IsSync(name string) bool {
	cmd, ok := r.commands[name]
	return ok && cmd.Type == "sync"
}

// IsAsync 是否是异步命令
func (r *Registry) IsAsync(name string) bool {
	cmd, ok := r.commands[name]
	return ok && cmd.Type == "async"
}

// All 返回所有命令
func (r *Registry) All() []*client.CommandDefinition {
	result := make([]*client.CommandDefinition, 0, len(r.commands))
	for _, cmd := range r.commands {
		result = append(result, cmd)
	}
	return result
}

// SyncCommands 返回所有同步命令
func (r *Registry) SyncCommands() []*client.CommandDefinition {
	return r.byType["sync"]
}

// AsyncCommands 返回所有异步命令
func (r *Registry) AsyncCommands() []*client.CommandDefinition {
	return r.byType["async"]
}
