package client

// RegisterRequest 注册请求
type RegisterRequest struct {
	PublicKey string `json:"public_key"`
	Name      string `json:"name"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	AgentID string `json:"agent_id"`
}

// ChallengeRequest 挑战请求
type ChallengeRequest struct {
	PublicKey string `json:"public_key"`
}

// ChallengeResponse 挑战响应
type ChallengeResponse struct {
	ChallengeID string `json:"challenge_id"`
	Nonce       []byte `json:"nonce"`
	ExpiresAt   int64  `json:"expires_at"` // Unix timestamp
}

// TokenRequest Token 请求
type TokenRequest struct {
	ChallengeID string `json:"challenge_id"`
	Signature   []byte `json:"signature"`
}

// TokenResponse Token 响应
type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"` // Unix timestamp
	AgentID   string `json:"agent_id"`
}

// CommandDefinition 命令定义（从服务端拉取）
type CommandDefinition struct {
	Name        string        `json:"name"`
	Type        string        `json:"type"` // sync / async
	Endpoint    string        `json:"endpoint"`
	Method      string        `json:"method"`
	Params      []CommandParam `json:"params"`
	Description string        `json:"description"`
}

// CommandParam 命令参数
type CommandParam struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // int, string, bool
	Required    bool   `json:"required"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}

// CommandsResponse 命令列表响应
type CommandsResponse struct {
	Commands []CommandDefinition `json:"commands"`
	Version  string              `json:"version"`
}

// TaskResponse 异步任务响应
type TaskResponse struct {
	TaskID        string `json:"task_id"`
	EstimatedTime string `json:"estimated_time,omitempty"`
}
