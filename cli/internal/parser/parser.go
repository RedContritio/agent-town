package parser

import (
	"fmt"
	"strconv"

	"github.com/RedContritio/agent-town/cli/internal/client"
)

// ParseArgs 根据命令定义解析位置参数
// 决策：位置参数按 params 定义的顺序解析，简单直观
// 可选：使用 flag 风格 --name value（更灵活但冗长）
func ParseArgs(params []client.CommandParam, args []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 检查必需参数
	requiredCount := 0
	for _, p := range params {
		if p.Required {
			requiredCount++
		}
	}

	if len(args) < requiredCount {
		return nil, fmt.Errorf("expected at least %d arguments, got %d", requiredCount, len(args))
	}

	// 按顺序解析参数
	for i, param := range params {
		var value interface{}
		var err error

		if i < len(args) {
			// 提供了参数值
			value, err = parseValue(param.Type, args[i])
			if err != nil {
				return nil, fmt.Errorf("parse param %q: %w", param.Name, err)
			}
		} else if param.Default != "" {
			// 使用默认值
			value, err = parseValue(param.Type, param.Default)
			if err != nil {
				return nil, fmt.Errorf("parse default for %q: %w", param.Name, err)
			}
		} else if param.Required {
			return nil, fmt.Errorf("missing required param %q", param.Name)
		} else {
			// 可选参数无默认值：跳过
			continue
		}

		result[param.Name] = value
	}

	return result, nil
}

// parseValue 根据类型解析值
func parseValue(paramType, raw string) (interface{}, error) {
	switch paramType {
	case "int":
		v, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("expected integer, got %q", raw)
		}
		return v, nil

	case "string":
		return raw, nil

	case "bool":
		// bool 类型通过 flag 形式 --param 传递，值为 true
		// 这里如果传了值，解析为 true（除非是 false/f/0/no）
		switch raw {
		case "false", "f", "0", "no":
			return false, nil
		default:
			return true, nil
		}

	default:
		return raw, nil // 未知类型按字符串处理
	}
}

// UsageString 生成使用说明字符串
func UsageString(params []client.CommandParam) string {
	result := ""
	for _, p := range params {
		if p.Required {
			result += fmt.Sprintf(" <%s>", p.Name)
		} else {
			result += fmt.Sprintf(" [%s]", p.Name)
		}
	}
	return result
}

// Validate 验证参数值是否符合类型要求
func Validate(params []client.CommandParam, values map[string]interface{}) error {
	for _, p := range params {
		value, ok := values[p.Name]
		if !ok {
			if p.Required {
				return fmt.Errorf("missing required param %q", p.Name)
			}
			continue
		}

		if err := validateType(p.Type, value); err != nil {
			return fmt.Errorf("param %q: %w", p.Name, err)
		}
	}

	return nil
}

func validateType(expectedType string, value interface{}) error {
	switch expectedType {
	case "int":
		if _, ok := value.(int); !ok {
			return fmt.Errorf("expected int, got %T", value)
		}
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected bool, got %T", value)
		}
	}
	return nil
}
