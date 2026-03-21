package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RedContritio/agent-town/server/internal/model"
)

func TestCommandHandler(t *testing.T) {
	handler := NewCommandHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/commands", nil)
	w := httptest.NewRecorder()

	handler.handleGetCommands(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusOK, w.Code)
	}

	var resp model.CommandsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if len(resp.Commands) == 0 {
		t.Error("命令列表为空")
	}

	// 检查是否有 move 命令
	foundMove := false
	for _, cmd := range resp.Commands {
		if cmd.Name == "move" {
			foundMove = true
			if cmd.Type != "async" {
				t.Errorf("move 命令应为 async 类型")
			}
			if cmd.Method != "POST" {
				t.Errorf("move 命令应为 POST 方法")
			}
			break
		}
	}
	if !foundMove {
		t.Error("未找到 move 命令")
	}
}

func TestCommandHandler_InvalidMethod(t *testing.T) {
	handler := NewCommandHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/commands", nil)
	w := httptest.NewRecorder()

	handler.handleGetCommands(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("期望状态码 %d, 得到 %d", http.StatusMethodNotAllowed, w.Code)
	}
}
