package parser

import (
	"testing"

	"github.com/RedContritio/agent-town/cli/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseArgs_Move(t *testing.T) {
	// move 命令: dx, dy 两个 int 参数
	params := []client.CommandParam{
		{Name: "dx", Type: "int", Required: true},
		{Name: "dy", Type: "int", Required: true},
	}

	result, err := ParseArgs(params, []string{"3", "0"})
	require.NoError(t, err)
	assert.Equal(t, 3, result["dx"])
	assert.Equal(t, 0, result["dy"])
}

func TestParseArgs_Harvest(t *testing.T) {
	// harvest 命令: resource_id 可选 string 参数
	params := []client.CommandParam{
		{Name: "resource_id", Type: "string", Required: false},
	}

	// 提供参数
	result, err := ParseArgs(params, []string{"tree-001"})
	require.NoError(t, err)
	assert.Equal(t, "tree-001", result["resource_id"])

	// 不提供参数
	result, err = ParseArgs(params, []string{})
	require.NoError(t, err)
	assert.Empty(t, result) // 空 map，表示使用默认行为
}

func TestParseArgs_Craft(t *testing.T) {
	// craft 命令: item (string, required), count (int, optional, default: 1)
	params := []client.CommandParam{
		{Name: "item", Type: "string", Required: true},
		{Name: "count", Type: "int", Required: false, Default: "1"},
	}

	// 提供两个参数
	result, err := ParseArgs(params, []string{"wood", "5"})
	require.NoError(t, err)
	assert.Equal(t, "wood", result["item"])
	assert.Equal(t, 5, result["count"])

	// 只提供必需参数
	result, err = ParseArgs(params, []string{"stone"})
	require.NoError(t, err)
	assert.Equal(t, "stone", result["item"])
	assert.Equal(t, 1, result["count"]) // 使用默认值
}

func TestParseArgs_MissingRequired(t *testing.T) {
	params := []client.CommandParam{
		{Name: "x", Type: "int", Required: true},
		{Name: "y", Type: "int", Required: true},
	}

	_, err := ParseArgs(params, []string{"1"}) // 只提供一个
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected at least 2")
}

func TestParseArgs_InvalidType(t *testing.T) {
	params := []client.CommandParam{
		{Name: "count", Type: "int", Required: true},
	}

	_, err := ParseArgs(params, []string{"not-a-number"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected integer")
}

func TestParseArgs_ExtraArgs(t *testing.T) {
	// 提供比定义更多的参数，多余的被忽略
	params := []client.CommandParam{
		{Name: "x", Type: "int", Required: true},
	}

	result, err := ParseArgs(params, []string{"1", "2", "3"})
	require.NoError(t, err)
	assert.Equal(t, 1, result["x"])
	// "2", "3" 被忽略
}

func TestParseValue(t *testing.T) {
	// int
	v, err := parseValue("int", "42")
	require.NoError(t, err)
	assert.Equal(t, 42, v)

	// string
	v, err = parseValue("string", "hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", v)

	// bool
	v, err = parseValue("bool", "true")
	require.NoError(t, err)
	assert.Equal(t, true, v)

	v, err = parseValue("bool", "false")
	require.NoError(t, err)
	assert.Equal(t, false, v)

	// 未知类型按字符串
	v, err = parseValue("unknown", "value")
	require.NoError(t, err)
	assert.Equal(t, "value", v)
}

func TestParseValue_InvalidInt(t *testing.T) {
	_, err := parseValue("int", "abc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected integer")
}

func TestUsageString(t *testing.T) {
	params := []client.CommandParam{
		{Name: "dx", Type: "int", Required: true},
		{Name: "dy", Type: "int", Required: true},
		{Name: "speed", Type: "int", Required: false},
	}

	usage := UsageString(params)
	assert.Equal(t, " <dx> <dy> [speed]", usage)
}

func TestUsageString_Empty(t *testing.T) {
	usage := UsageString([]client.CommandParam{})
	assert.Equal(t, "", usage)
}

func TestValidate(t *testing.T) {
	params := []client.CommandParam{
		{Name: "id", Type: "string", Required: true},
		{Name: "count", Type: "int", Required: false},
	}

	// 有效
	err := Validate(params, map[string]interface{}{
		"id":    "test",
		"count": 5,
	})
	assert.NoError(t, err)

	// 类型错误
	err = Validate(params, map[string]interface{}{
		"id":    123, // 应该是 string
		"count": 5,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected string")
}

func TestValidate_MissingRequired(t *testing.T) {
	params := []client.CommandParam{
		{Name: "id", Type: "string", Required: true},
	}

	err := Validate(params, map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required")
}
