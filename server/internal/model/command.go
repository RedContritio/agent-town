package model

// CommandDefinition 命令定义
type CommandDefinition struct {
	Name        string              `json:"name"`
	Type        string              `json:"type"` // sync / async
	Endpoint    string              `json:"endpoint"`
	Method      string              `json:"method"`
	Params      []CommandParam      `json:"params"`
	Description string              `json:"description"`
}

// CommandParam 命令参数
type CommandParam struct {
	Name     string `json:"name"`
	Type     string `json:"type"`     // int, string, bool, etc.
	Required bool   `json:"required"`
	Default  string `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}

// CommandsResponse 命令列表响应
type CommandsResponse struct {
	Commands []CommandDefinition `json:"commands"`
	Version  string              `json:"version"`
}

// DefaultCommands 返回默认命令列表
func DefaultCommands() []CommandDefinition {
	return []CommandDefinition{
		// 状态查询（同步）
		{
			Name:        "status",
			Type:        "sync",
			Endpoint:    "/api/v1/status",
			Method:      "GET",
			Params:      []CommandParam{},
			Description: "获取自身状态",
		},
		// 视野系统（同步）
		{
			Name:        "look",
			Type:        "sync",
			Endpoint:    "/api/v1/look",
			Method:      "GET",
			Params:      []CommandParam{},
			Description: "观察周围（视野范围）",
		},
		{
			Name:        "scan",
			Type:        "sync",
			Endpoint:    "/api/v1/scan",
			Method:      "GET",
			Params:      []CommandParam{},
			Description: "扫描脚下（当前chunk）",
		},
		{
			Name:        "inventory",
			Type:        "sync",
			Endpoint:    "/api/v1/inventory",
			Method:      "GET",
			Params:      []CommandParam{},
			Description: "查看背包",
		},
		// 任务栈管理（同步）
		{
			Name:        "stack",
			Type:        "sync",
			Endpoint:    "/api/v1/stack",
			Method:      "GET",
			Params:      []CommandParam{},
			Description: "查看任务栈",
		},
		// 任务提交（异步）
		{
			Name:     "move",
			Type:     "async",
			Endpoint: "/api/v1/stack",
			Method:   "POST",
			Params: []CommandParam{
				{Name: "dx", Type: "int", Required: true, Description: "X方向移动距离"},
				{Name: "dy", Type: "int", Required: true, Description: "Y方向移动距离"},
			},
			Description: "移动",
		},
		{
			Name:     "harvest",
			Type:     "async",
			Endpoint: "/api/v1/stack",
			Method:   "POST",
			Params: []CommandParam{
				{Name: "resource_id", Type: "string", Required: false, Description: "资源ID（可选，默认为最近资源）"},
			},
			Description: "采集资源",
		},
		{
			Name:     "craft",
			Type:     "async",
			Endpoint: "/api/v1/stack",
			Method:   "POST",
			Params: []CommandParam{
				{Name: "item", Type: "string", Required: true, Description: "物品名称"},
				{Name: "count", Type: "int", Required: false, Default: "1", Description: "数量"},
			},
			Description: "制作物品（暂不消耗材料）",
		},
		{
			Name:     "build",
			Type:     "async",
			Endpoint: "/api/v1/stack",
			Method:   "POST",
			Params: []CommandParam{
				{Name: "type", Type: "string", Required: true, Description: "建筑类型"},
				{Name: "x", Type: "int", Required: true, Description: "X坐标"},
				{Name: "y", Type: "int", Required: true, Description: "Y坐标"},
				{Name: "name", Type: "string", Required: false, Description: "建筑名称"},
			},
			Description: "建造建筑",
		},
		// 任务控制（同步）
		{
			Name:        "pause",
			Type:        "sync",
			Endpoint:    "/api/v1/stack/{task_id}/pause",
			Method:      "POST",
			Params: []CommandParam{
				{Name: "task_id", Type: "string", Required: true, Description: "任务ID"},
			},
			Description: "暂停任务",
		},
		{
			Name:        "resume",
			Type:        "sync",
			Endpoint:    "/api/v1/stack/{task_id}/resume",
			Method:      "POST",
			Params: []CommandParam{
				{Name: "task_id", Type: "string", Required: true, Description: "任务ID"},
			},
			Description: "恢复任务",
		},
		{
			Name:        "drop",
			Type:        "sync",
			Endpoint:    "/api/v1/stack/{task_id}",
			Method:      "DELETE",
			Params: []CommandParam{
				{Name: "task_id", Type: "string", Required: true, Description: "任务ID"},
			},
			Description: "放弃任务",
		},
	}
}
