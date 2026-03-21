package commands

// CommandDefinition 命令定义
type CommandDefinition struct {
	Name        string         `json:"name"`
	Type        string         `json:"type"` // sync / async
	Endpoint    string         `json:"endpoint"`
	Method      string         `json:"method"`
	Params      []CommandParam `json:"params"`
	Description string         `json:"description"`
}

// CommandParam 命令参数
type CommandParam struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // int, string, bool, etc.
	Required    bool   `json:"required"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}

// CommandsResponse 命令列表响应
type CommandsResponse struct {
	Commands []CommandDefinition `json:"commands"`
	Version  string              `json:"version"`
}
