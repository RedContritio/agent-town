package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// Formatter 输出格式化接口
type Formatter interface {
	Format(data interface{}, w io.Writer) error
}

// TableFormatter 表格格式
type TableFormatter struct{}

// NewTable 创建表格格式化器
func NewTable() *TableFormatter {
	return &TableFormatter{}
}

// Format 格式化为表格
// 决策：简单 key-value 表格，适合大多数 sync 命令
// 可选：根据数据类型动态选择布局（列表、树形等）
func (f *TableFormatter) Format(data interface{}, w io.Writer) error {
	// 尝试转换为 map
	m, ok := data.(map[string]interface{})
	if !ok {
		// 如果不是 map，直接打印
		fmt.Fprintln(w, data)
		return nil
	}

	if len(m) == 0 {
		fmt.Fprintln(w, "(empty result)")
		return nil
	}

	// 使用 tabwriter 对齐
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	for k, v := range m {
		fmt.Fprintf(tw, "%s:\t%v\n", k, formatValue(v))
	}
	return tw.Flush()
}

// formatValue 格式化单个值
func formatValue(v interface{}) string {
	switch val := v.(type) {
	case []interface{}:
		// 数组格式化
		items := make([]string, len(val))
		for i, item := range val {
			items[i] = fmt.Sprintf("%v", item)
		}
		return strings.Join(items, ", ")
	case map[string]interface{}:
		// 嵌套 map 简单表示
		return "{...}"
	default:
		return fmt.Sprintf("%v", val)
	}
}

// JSONFormatter JSON 格式
type JSONFormatter struct {
	Indent bool
}

// NewJSON 创建 JSON 格式化器
func NewJSON(indent bool) *JSONFormatter {
	return &JSONFormatter{Indent: indent}
}

// Format 格式化为 JSON
func (f *JSONFormatter) Format(data interface{}, w io.Writer) error {
	encoder := json.NewEncoder(w)
	if f.Indent {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(data)
}

// FormatSyncResult 格式化同步命令结果（便捷函数）
func FormatSyncResult(data map[string]interface{}, useJSON bool) string {
	var b strings.Builder
	var f Formatter
	if useJSON {
		f = NewJSON(true)
	} else {
		f = NewTable()
	}
	f.Format(data, &b)
	return b.String()
}

// FormatAsyncResult 格式化异步任务结果
func FormatAsyncResult(taskID, cmdName string, params map[string]interface{}, estTime string) string {
	paramsStr := formatParams(params)
	if estTime != "" {
		return fmt.Sprintf("%s: %s%s (est: %s)", taskID, cmdName, paramsStr, estTime)
	}
	return fmt.Sprintf("%s: %s%s", taskID, cmdName, paramsStr)
}

// formatParams 格式化参数
func formatParams(params map[string]interface{}) string {
	if len(params) == 0 {
		return ""
	}
	var parts []string
	for _, v := range params {
		parts = append(parts, fmt.Sprintf("%v", v))
	}
	return " " + strings.Join(parts, " ")
}
