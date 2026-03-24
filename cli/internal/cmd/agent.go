package cmd

import (
	"fmt"
	"os"

	"github.com/RedContritio/agent-town/cli/internal/client"
	"github.com/RedContritio/agent-town/cli/internal/config"
	"github.com/RedContritio/agent-town/cli/internal/keystore"
	"github.com/spf13/cobra"
)

// AgentCmd 返回 agent 根命令
func AgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Manage local agents",
		Long:  "Create, list, export, and delete local agent identities",
	}

	cmd.AddCommand(
		agentCreateCmd(),
		agentListCmd(),
		agentDeleteCmd(),
		agentExportCmd(),
	)

	return cmd
}

func agentCreateCmd() *cobra.Command {
	var name, server string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new agent",
		Long:  "Generate a new key pair and register with the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunAgentCreate(name, server)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Agent name (required)")
	cmd.Flags().StringVar(&server, "server", "", "Server URL (required)")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("server")

	return cmd
}

func agentListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List local agents",
		Long:  "Display all locally stored agent identities",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunAgentList()
		},
	}
}

func agentDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete an agent",
		Long:  "Remove an agent's local key and configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunAgentDelete(args[0])
		},
	}
}

func agentExportCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "export [name]",
		Short: "Export agent key",
		Long:  "Export an agent's private key to a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunAgentExport(args[0], output)
		},
	}

	cmd.Flags().StringVar(&output, "output", "", "Output file path (default: stdout)")

	return cmd
}

// runAgentCreate 执行 create 命令
func RunAgentCreate(name, server string) error {
	// 检查 agent 是否已存在
	names, err := keystore.List()
	if err != nil {
		return fmt.Errorf("list agents: %w", err)
	}
	for _, n := range names {
		if n == name {
			return fmt.Errorf("agent %q already exists locally", name)
		}
	}

	// 1. 本地生成密钥
	key, err := keystore.Create(name)
	if err != nil {
		return fmt.Errorf("create key: %w", err)
	}

	// 2. 连接服务端注册
	c := client.New(server)
	resp, err := c.Auth().Register(&client.RegisterRequest{
		PublicKey: key.PublicKeyHex(),
		Name:      name,
	})
	if err != nil {
		// 回滚：删除本地密钥
		keystore.Delete(name)
		return fmt.Errorf("register with server: %w", err)
	}

	// 3. 保存配置
	cfg, err := config.Load()
	if err != nil {
		// 回滚
		keystore.Delete(name)
		return fmt.Errorf("load config: %w", err)
	}

	cfg.SetAgent(name, &config.AgentConfig{
		AgentID:   resp.AgentID,
		ServerURL: server,
	})

	if err := cfg.Save(); err != nil {
		// 回滚
		keystore.Delete(name)
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("Agent %q created successfully\n", name)
	fmt.Printf("Agent ID: %s\n", resp.AgentID)
	fmt.Printf("Server: %s\n", server)
	return nil
}

// runAgentList 执行 list 命令
func RunAgentList() error {
	names, err := keystore.List()
	if err != nil {
		return fmt.Errorf("list agents: %w", err)
	}

	if len(names) == 0 {
		fmt.Println("No agents found")
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// 简单的表格输出
	fmt.Printf("%-15s %-20s %-30s\n", "NAME", "AGENT ID", "SERVER")
	fmt.Println("----------------------------------------------------------------------")

	for _, name := range names {
		if agent, ok := cfg.GetAgent(name); ok {
			fmt.Printf("%-15s %-20s %-30s\n", name, agent.AgentID, agent.ServerURL)
		} else {
			fmt.Printf("%-15s %-20s %-30s\n", name, "(unknown)", "(unknown)")
		}
	}

	return nil
}

// runAgentDelete 执行 delete 命令
func RunAgentDelete(name string) error {
	// 确认存在
	_, err := keystore.Load(name)
	if err != nil {
		return fmt.Errorf("agent not found: %w", err)
	}

	// 删除密钥
	if err := keystore.Delete(name); err != nil {
		return fmt.Errorf("delete key: %w", err)
	}

	// 删除配置
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	cfg.DeleteAgent(name)
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	fmt.Printf("Agent %q deleted\n", name)
	return nil
}

// runAgentExport 执行 export 命令
func RunAgentExport(name, output string) error {
	pemStr, err := keystore.Export(name)
	if err != nil {
		return fmt.Errorf("export key: %w", err)
	}

	if output == "" {
		// 输出到 stdout
		fmt.Print(pemStr)
	} else {
		// 写入文件
		if err := os.WriteFile(output, []byte(pemStr), 0600); err != nil {
			return fmt.Errorf("write file: %w", err)
		}
		fmt.Printf("Key exported to %s\n", output)
	}

	return nil
}
