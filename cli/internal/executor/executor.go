package executor

import (
	"fmt"

	"github.com/RedContritio/agent-town/cli/internal/client"
	"github.com/RedContritio/agent-town/cli/internal/parser"
)

// Executor 命令执行器
type Executor struct {
	Client *client.Client
}

// New 创建执行器
func New(c *client.Client) *Executor {
	return &Executor{Client: c}
}

// ExecuteSync 执行同步命令
// 等待服务端响应并返回完整结果
func (e *Executor) ExecuteSync(
	def *client.CommandDefinition,
	params map[string]interface{},
) (map[string]interface{}, error) {
	// 验证参数
	if err := parser.Validate(def.Params, params); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	// 构建请求体
	body := buildRequestBody(params)

	// 发送请求
	var result map[string]interface{}
	if err := e.Client.Do(def.Method, def.Endpoint, body, &result); err != nil {
		return nil, fmt.Errorf("execute %s: %w", def.Name, err)
	}

	return result, nil
}

// ExecuteAsync 执行异步命令
// 立即返回 task_id，不等待完成
func (e *Executor) ExecuteAsync(
	def *client.CommandDefinition,
	params map[string]interface{},
) (*client.TaskResponse, error) {
	// 验证参数
	if err := parser.Validate(def.Params, params); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}

	// 构建请求体
	body := buildRequestBody(params)

	// 发送请求
	var result client.TaskResponse
	if err := e.Client.Do(def.Method, def.Endpoint, body, &result); err != nil {
		return nil, fmt.Errorf("execute %s: %w", def.Name, err)
	}

	return &result, nil
}

// buildRequestBody 将参数构建为请求体
// 决策：所有参数放入一个 map 作为请求体
// 可选：根据 endpoint 模板替换路径参数（如 /agents/{id}）
func buildRequestBody(params map[string]interface{}) map[string]interface{} {
	if len(params) == 0 {
		return nil
	}
	return params
}

// FormatAsyncResult 格式化异步命令输出
// 格式: "{taskID}: {cmd} {params} (est: {time})"
func FormatAsyncResult(def *client.CommandDefinition, params map[string]interface{}, resp *client.TaskResponse) string {
	paramsStr := formatParams(params)
	if resp.EstimatedTime != "" {
		return fmt.Sprintf("%s: %s%s (est: %s)", resp.TaskID, def.Name, paramsStr, resp.EstimatedTime)
	}
	return fmt.Sprintf("%s: %s%s", resp.TaskID, def.Name, paramsStr)
}

// formatParams 格式化参数为字符串
func formatParams(params map[string]interface{}) string {
	if len(params) == 0 {
		return ""
	}

	result := ""
	for _, v := range params {
		result += fmt.Sprintf(" %v", v)
	}
	return result
}
