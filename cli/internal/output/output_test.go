package output

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTableFormatter_Format(t *testing.T) {
	f := NewTable()

	data := map[string]interface{}{
		"hp":      100,
		"stamina": 80,
		"name":    "Alice",
	}

	var b strings.Builder
	err := f.Format(data, &b)
	require.NoError(t, err)

	output := b.String()
	assert.Contains(t, output, "hp:")
	assert.Contains(t, output, "100")
	assert.Contains(t, output, "stamina:")
	assert.Contains(t, output, "80")
}

func TestTableFormatter_Format_Empty(t *testing.T) {
	f := NewTable()

	var b strings.Builder
	err := f.Format(map[string]interface{}{}, &b)
	require.NoError(t, err)
	assert.Contains(t, b.String(), "empty")
}

func TestTableFormatter_Format_NotMap(t *testing.T) {
	f := NewTable()

	var b strings.Builder
	err := f.Format("simple string", &b)
	require.NoError(t, err)
	assert.Contains(t, b.String(), "simple string")
}

func TestTableFormatter_Format_Array(t *testing.T) {
	f := NewTable()

	data := map[string]interface{}{
		"items": []interface{}{"wood", "stone", "food"},
	}

	var b strings.Builder
	err := f.Format(data, &b)
	require.NoError(t, err)
	assert.Contains(t, b.String(), "wood, stone, food")
}

func TestJSONFormatter_Format(t *testing.T) {
	f := NewJSON(true)

	data := map[string]interface{}{
		"hp":   100,
		"name": "Alice",
	}

	var b strings.Builder
	err := f.Format(data, &b)
	require.NoError(t, err)

	output := b.String()
	assert.Contains(t, output, "\"hp\":")
	assert.Contains(t, output, "100")
	assert.Contains(t, output, "{\n") // 有缩进
}

func TestJSONFormatter_Format_NoIndent(t *testing.T) {
	f := NewJSON(false)

	data := map[string]interface{}{
		"hp": 100,
	}

	var b strings.Builder
	err := f.Format(data, &b)
	require.NoError(t, err)

	// 没有缩进，一行
	output := strings.TrimSpace(b.String())
	assert.True(t, strings.HasPrefix(output, "{"))
	assert.True(t, strings.HasSuffix(output, "}"))
	assert.NotContains(t, output, "\n")
}

func TestFormatSyncResult_JSON(t *testing.T) {
	data := map[string]interface{}{"hp": 100}
	result := FormatSyncResult(data, true)
	assert.Contains(t, result, "\"hp\"")
}

func TestFormatSyncResult_Table(t *testing.T) {
	data := map[string]interface{}{"hp": 100}
	result := FormatSyncResult(data, false)
	assert.Contains(t, result, "hp:")
}

func TestFormatAsyncResult(t *testing.T) {
	params := map[string]interface{}{"dx": 3, "dy": 0}
	result := FormatAsyncResult("Alice-001", "move", params, "6s")
	assert.Equal(t, "Alice-001: move 3 0 (est: 6s)", result)
}

func TestFormatAsyncResult_NoEstTime(t *testing.T) {
	params := map[string]interface{}{"resource_id": "tree-001"}
	result := FormatAsyncResult("Alice-002", "harvest", params, "")
	assert.Equal(t, "Alice-002: harvest tree-001", result)
}

func TestFormatAsyncResult_NoParams(t *testing.T) {
	result := FormatAsyncResult("Alice-003", "look", nil, "1s")
	assert.Equal(t, "Alice-003: look (est: 1s)", result)
}

func TestFormatValue(t *testing.T) {
	assert.Equal(t, "test", formatValue("test"))
	assert.Equal(t, "123", formatValue(123))
	assert.Equal(t, "true", formatValue(true))
	assert.Equal(t, "a, b, c", formatValue([]interface{}{"a", "b", "c"}))
	assert.Equal(t, "{...}", formatValue(map[string]interface{}{"k": "v"}))
}

func TestFormatParams(t *testing.T) {
	assert.Equal(t, "", formatParams(nil))
	assert.Equal(t, "", formatParams(map[string]interface{}{}))
	// 注意：map 遍历顺序不定，所以只检查包含
	result := formatParams(map[string]interface{}{"x": 1, "y": 2})
	assert.Contains(t, result, "1")
	assert.Contains(t, result, "2")
}
