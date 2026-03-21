package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/RedContritio/agent-town/cli/internal/client"
)

// Executor 命令执行器
type Executor struct {
	client   *client.Client
	registry *Registry
}

// NewExecutor 创建命令执行器
func NewExecutor(cli *client.Client, registry *Registry) *Executor {
	return &Executor{
		client:   cli,
		registry: registry,
	}
}

// Execute 执行命令
func (e *Executor) Execute(cmdName string, args []string) error {
	// 确保命令已加载
	if err := e.registry.Discover(false); err != nil {
		return fmt.Errorf("命令发现失败，请检查服务器连接: %w", err)
	}

	// 查找命令
	cmd, found := e.registry.GetCommand(cmdName)
	if !found {
		return fmt.Errorf("未知命令: %s", cmdName)
	}

	// 解析参数
	params, err := e.parseParams(cmd, args)
	if err != nil {
		return err
	}

	// 构建请求体
	var body interface{}
	if cmd.Name == "move" || cmd.Name == "harvest" || cmd.Name == "craft" || cmd.Name == "build" {
		// 任务提交需要包装
		body = map[string]interface{}{
			"type":   getTaskType(cmd.Name),
			"params": params,
		}
	} else {
		body = params
	}

	// 替换 URL 中的参数
	endpoint := cmd.Endpoint
	for key, value := range params {
		placeholder := "{" + key + "}"
		if strings.Contains(endpoint, placeholder) {
			endpoint = strings.ReplaceAll(endpoint, placeholder, fmt.Sprintf("%v", value))
			delete(params, key)
		}
	}

	// 执行请求
	var result interface{}
	if err := e.client.DoJSON(cmd.Method, endpoint, body, &result); err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}

	// 输出结果
	return e.outputResult(cmd, result)
}

// parseParams 解析命令行参数
func (e *Executor) parseParams(cmd *CommandDefinition, args []string) (map[string]interface{}, error) {
	params := make(map[string]interface{})

	// 处理位置参数（简单实现：按顺序分配）
	posIdx := 0
	for _, param := range cmd.Params {
		if param.Required && posIdx < len(args) {
			value, err := convertType(args[posIdx], param.Type)
			if err != nil {
				return nil, fmt.Errorf("参数 %s 类型错误: %w", param.Name, err)
			}
			params[param.Name] = value
			posIdx++
		} else if param.Required {
			return nil, fmt.Errorf("缺少必需参数: %s", param.Name)
		} else if param.Default != "" {
			value, _ := convertType(param.Default, param.Type)
			params[param.Name] = value
		}
	}

	// 处理剩余参数（键值对形式 --key value）
	for posIdx < len(args) {
		arg := args[posIdx]
		if strings.HasPrefix(arg, "--") {
			key := strings.TrimPrefix(arg, "--")
			posIdx++
			if posIdx >= len(args) {
				return nil, fmt.Errorf("参数 %s 缺少值", key)
			}
			value := args[posIdx]

			// 查找参数定义
			for _, param := range cmd.Params {
				if param.Name == key {
					converted, err := convertType(value, param.Type)
					if err != nil {
						return nil, fmt.Errorf("参数 %s 类型错误: %w", key, err)
					}
					params[key] = converted
					break
				}
			}
		}
		posIdx++
	}

	return params, nil
}

// outputResult 输出结果
func (e *Executor) outputResult(cmd *CommandDefinition, result interface{}) error {
	// 根据命令类型格式化输出
	switch cmd.Name {
	case "stack":
		return e.outputStack(result)
	case "status", "me":
		return e.outputStatus(result)
	default:
		// 默认 JSON 输出
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}
}

// outputStack 输出任务栈
func (e *Executor) outputStack(result interface{}) error {
	data, ok := result.(map[string]interface{})
	if !ok {
		return jsonOutput(result)
	}

	stack, ok := data["stack"].([]interface{})
	if !ok || len(stack) == 0 {
		fmt.Println("任务栈为空")
		return nil
	}

	fmt.Println("DEPTH  TASK              STATUS      TYPE")
	for _, item := range stack {
		task, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		depth := int(task["stack_depth"].(float64))
		taskID := task["task_id"].(string)
		status := getTaskStatusName(int(task["status"].(float64)))
		taskType := getTaskTypeName(int(task["type"].(float64)))

		fmt.Printf("%-6d %-17s %-11s %s\n", depth, taskID, status, taskType)
	}
	return nil
}

// outputStatus 输出状态
func (e *Executor) outputStatus(result interface{}) error {
	data, ok := result.(map[string]interface{})
	if !ok {
		return jsonOutput(result)
	}

	fmt.Printf("Agent: %s\n", data["name"])
	fmt.Printf("Position: (%v, %v)\n", data["position"].(map[string]interface{})["x"], data["position"].(map[string]interface{})["y"])
	fmt.Printf("HP: %v/%v\n", data["hp"], data["maxHp"])
	fmt.Printf("Stamina: %v/%v\n", data["stamina"], data["maxStamina"])
	fmt.Printf("Balance: %v\n", data["balance"])
	return nil
}

// jsonOutput 默认 JSON 输出
func jsonOutput(result interface{}) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// convertType 类型转换
func convertType(value string, targetType string) (interface{}, error) {
	switch targetType {
	case "int":
		var i int
		_, err := fmt.Sscanf(value, "%d", &i)
		return i, err
	case "bool":
		return value == "true" || value == "yes" || value == "1", nil
	case "string":
		return value, nil
	default:
		return value, nil
	}
}

// getTaskType 获取任务类型数值
func getTaskType(name string) int {
	switch name {
	case "move":
		return 0
	case "harvest":
		return 1
	case "craft":
		return 2
	case "build":
		return 3
	case "combat":
		return 4
	default:
		return -1
	}
}

// getTaskTypeName 获取任务类型名称
func getTaskTypeName(taskType int) string {
	names := map[int]string{
		0: "move",
		1: "harvest",
		2: "craft",
		3: "build",
		4: "combat",
	}
	if name, ok := names[taskType]; ok {
		return name
	}
	return "unknown"
}

// getTaskStatusName 获取任务状态名称
func getTaskStatusName(status int) string {
	names := map[int]string{
		0: "pending",
		1: "running",
		2: "paused",
		3: "completed",
		4: "failed",
	}
	if name, ok := names[status]; ok {
		return name
	}
	return "unknown"
}

// PrintHelp 打印帮助信息
func (e *Executor) PrintHelp() {
	if err := e.registry.Discover(false); err != nil {
		fmt.Fprintf(os.Stderr, "无法获取命令列表: %v\n", err)
		return
	}

	fmt.Println("可用命令:")
	fmt.Println()

	syncCmds := []CommandDefinition{}
	asyncCmds := []CommandDefinition{}

	for _, cmd := range e.registry.GetCommands() {
		if cmd.Type == "sync" {
			syncCmds = append(syncCmds, cmd)
		} else {
			asyncCmds = append(asyncCmds, cmd)
		}
	}

	fmt.Println("同步命令 (立即返回):")
	for _, cmd := range syncCmds {
		fmt.Printf("  %-15s %s\n", cmd.Name, cmd.Description)
	}
	fmt.Println()

	fmt.Println("异步命令 (压入任务栈):")
	for _, cmd := range asyncCmds {
		fmt.Printf("  %-15s %s\n", cmd.Name, cmd.Description)
	}
}
