package builder

import (
	"fmt"
	"strings"

	"github.com/RedContritio/agent-town/cli/internal/client"
	"github.com/RedContritio/agent-town/cli/internal/executor"
	"github.com/RedContritio/agent-town/cli/internal/parser"
	"github.com/RedContritio/agent-town/cli/internal/registry"
	"github.com/spf13/cobra"
)

// Builder 动态命令构建器
type Builder struct {
	Registry *registry.Registry
	Executor *executor.Executor
}

// New 创建构建器
func New(r *registry.Registry, e *executor.Executor) *Builder {
	return &Builder{
		Registry: r,
		Executor: e,
	}
}

// BuildCommand 为单个命令定义构建 cobra.Command
func (b *Builder) BuildCommand(def *client.CommandDefinition) *cobra.Command {
	use := def.Name + parser.UsageString(def.Params)

	cmd := &cobra.Command{
		Use:   use,
		Short: def.Description,
		Long:  fmt.Sprintf("%s (%s)", def.Description, def.Type),
		RunE: func(cmd *cobra.Command, args []string) error {
			return b.runCommand(def, args)
		},
	}

	return cmd
}

// runCommand 执行命令
func (b *Builder) runCommand(def *client.CommandDefinition, args []string) error {
	// 解析参数
	params, err := parser.ParseArgs(def.Params, args)
	if err != nil {
		return fmt.Errorf("parse args: %w", err)
	}

	// 根据类型执行
	if def.Type == "sync" {
		result, err := b.Executor.ExecuteSync(def, params)
		if err != nil {
			return err
		}
		// 简单打印结果（后续会替换为格式化输出）
		return printSyncResult(result)
	} else {
		resp, err := b.Executor.ExecuteAsync(def, params)
		if err != nil {
			return err
		}
		// 打印异步任务信息
		fmt.Println(executor.FormatAsyncResult(def, params, resp))
		return nil
	}
}

// printSyncResult 打印同步命令结果（临时实现）
func printSyncResult(result map[string]interface{}) error {
	for k, v := range result {
		fmt.Printf("%s: %v\n", k, v)
	}
	return nil
}

// BuildRoot 构建完整的命令树根节点
// 包含本地命令（agent）和从服务端拉取的远程命令
func (b *Builder) BuildRoot(localCmds []*cobra.Command) *cobra.Command {
	root := &cobra.Command{
		Use:   "at-cli",
		Short: "Agent-Town CLI",
		Long:  "Command line interface for Agent-Town",
	}

	// 添加本地命令
	for _, cmd := range localCmds {
		root.AddCommand(cmd)
	}

	// 从 registry 添加远程命令
	for _, def := range b.Registry.All() {
		// 跳过与本地命令冲突的
		if isConflict(def.Name, localCmds) {
			continue
		}
		root.AddCommand(b.BuildCommand(def))
	}

	return root
}

// isConflict 检查命令名是否与本地命令冲突
func isConflict(name string, localCmds []*cobra.Command) bool {
	for _, cmd := range localCmds {
		if cmd.Name() == name {
			return true
		}
		// 检查子命令
		for _, sub := range cmd.Commands() {
			if sub.Name() == name {
				return true
			}
		}
	}
	return false
}

// BuildHelpString 构建帮助文本
func BuildHelpString(r *registry.Registry) string {
	var sb strings.Builder

	sb.WriteString("\nAvailable Commands:\\n")

	if syncs := r.SyncCommands(); len(syncs) > 0 {
		sb.WriteString("\n  [Sync Commands]:\n")
		for _, cmd := range syncs {
			sb.WriteString(fmt.Sprintf("    %-15s %s\n", cmd.Name, cmd.Description))
		}
	}

	if asyncs := r.AsyncCommands(); len(asyncs) > 0 {
		sb.WriteString("\n  [Async Commands]:\n")
		for _, cmd := range asyncs {
			sb.WriteString(fmt.Sprintf("    %-15s %s\n", cmd.Name, cmd.Description))
		}
	}

	return sb.String()
}
