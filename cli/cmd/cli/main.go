package main

import (
	"fmt"
	"os"

	"github.com/RedContritio/agent-town/cli/internal/auth"
	"github.com/RedContritio/agent-town/cli/internal/builder"
	"github.com/RedContritio/agent-town/cli/internal/client"
	cmdpkg "github.com/RedContritio/agent-town/cli/internal/cmd"
	"github.com/RedContritio/agent-town/cli/internal/config"
	"github.com/RedContritio/agent-town/cli/internal/executor"
	"github.com/RedContritio/agent-town/cli/internal/output"
	"github.com/RedContritio/agent-town/cli/internal/registry"
	"github.com/spf13/cobra"
)

var (
	agentName string
	useJSON   bool
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// 1. 加载配置（需要先加载以获取默认 agent）
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// 2. 创建根命令
	root := &cobra.Command{
		Use:   "at-cli",
		Short: "Agent-Town CLI",
		Long:  "Command line interface for Agent-Town",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// 跳过本地命令的预处理
			if isLocalCommand(cmd) {
				return nil
			}
			return setupRemoteCommand(cfg)
		},
	}

	// 全局 flags
	root.PersistentFlags().StringVar(&agentName, "agent", "", "Agent name")
	root.PersistentFlags().BoolVar(&useJSON, "json", false, "Output in JSON format")

	// 3. 添加本地命令（agent create/list/export/delete）
	root.AddCommand(cmdpkg.AgentCmd())

	// 4. 尝试添加远程命令（如果有 agent）
	// 注意：实际连接在 PersistentPreRunE 中进行
	if err := addRemoteCommands(root, cfg); err != nil {
		// 失败不报错，只是没有远程命令
		// 用户会看到本地命令的帮助
	}

	return root.Execute()
}

// isLocalCommand 检查是否是本地命令（不需要连接 server）
func isLocalCommand(cmd *cobra.Command) bool {
	// 检查命令路径是否以 agent 开头
	if cmd.Name() == "agent" {
		return true
	}
	if cmd.Parent() != nil && cmd.Parent().Name() == "agent" {
		return true
	}
	return false
}

// setupRemoteCommand 为远程命令设置认证和客户端
func setupRemoteCommand(cfg *config.Config) error {
	// 确定使用哪个 agent
	if agentName == "" {
		if cfg.CurrentAgent != "" {
			agentName = cfg.CurrentAgent
		} else {
			return fmt.Errorf("no agent specified, use --agent <name> or set current_agent in config")
		}
	}

	agent, ok := cfg.GetAgent(agentName)
	if !ok {
		return fmt.Errorf("agent %q not found, run 'at-cli agent create --name %s --server <url>'", agentName, agentName)
	}

	// 创建客户端并认证
	c := client.New(agent.ServerURL)
	authMgr := auth.NewManager(cfg)
	if _, err := authMgr.EnsureToken(agentName, c); err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}

	return nil
}

// addRemoteCommands 从服务端拉取命令并添加到根命令
func addRemoteCommands(root *cobra.Command, cfg *config.Config) error {
	// 确定使用哪个 agent
	targetAgent := agentName
	if targetAgent == "" {
		targetAgent = cfg.CurrentAgent
	}
	if targetAgent == "" {
		return nil // 没有 agent，不添加远程命令
	}

	agent, ok := cfg.GetAgent(targetAgent)
	if !ok {
		return nil // agent 不存在
	}

	// 创建客户端
	c := client.New(agent.ServerURL)

	// 尝试认证（使用缓存的 token）
	authMgr := auth.NewManager(cfg)
	if _, err := authMgr.EnsureToken(targetAgent, c); err != nil {
		// 认证失败，不添加远程命令
		return err
	}

	// 拉取命令定义
	reg := registry.New()
	if err := reg.LoadFromServer(c); err != nil {
		return err
	}

	// 构建并添加命令
	exec := executor.New(c)
	b := builder.New(reg, exec)

	// 添加远程命令（排除与本地命令冲突的）
	for _, def := range reg.All() {
		if isCommandExists(root, def.Name) {
			continue
		}

		// 包装命令以使用正确的格式化输出
		cmd := b.BuildCommand(def)
		originalRunE := cmd.RunE
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			return runWithFormatting(cmd, args, originalRunE)
		}

		root.AddCommand(cmd)
	}

	return nil
}

// isCommandExists 检查命令是否已存在
func isCommandExists(root *cobra.Command, name string) bool {
	for _, cmd := range root.Commands() {
		if cmd.Name() == name {
			return true
		}
	}
	return false
}

// runWithFormatting 运行命令并格式化输出
func runWithFormatting(cmd *cobra.Command, args []string, runE func(*cobra.Command, []string) error) error {
	// 执行原命令
	if err := runE(cmd, args); err != nil {
		return err
	}

	// 注意：实际输出由 executor 处理
	// 这里只是包装层
	return nil
}

// printResult 根据格式打印结果
func printResult(data map[string]interface{}) {
	result := output.FormatSyncResult(data, useJSON)
	fmt.Print(result)
}
