package model

// Token Token 实体
type Token struct {
	Token       string `json:"token"`
	AgentID     int64  `json:"agentId"`
	TokenType   int    `json:"tokenType"`
	Scopes      int    `json:"scopes"`
	CreatedAt   int64  `json:"createdAt"`
	ExpiresAt   *int64 `json:"expiresAt,omitempty"`
	LastUsedAt  *int64 `json:"lastUsedAt,omitempty"`
}

// IsExpired 检查 Token 是否过期
func (t *Token) IsExpired() bool {
	if t.ExpiresAt == nil {
		return false // 永不过期
	}
	return NowMillis() > *t.ExpiresAt
}

// UpdateLastUsed 更新最后使用时间
func (t *Token) UpdateLastUsed() {
	now := NowMillis()
	t.LastUsedAt = &now
}

// HasScope 检查是否有指定权限
func (t *Token) HasScope(scope int) bool {
	return t.Scopes&scope != 0
}

// TokenType 常量
const (
	TokenTypeCLI = 0
	TokenTypeWeb = 1
)

// Scope 权限位
const (
	ScopeRead   = 0x1  // 1
	ScopeWrite  = 0x2  // 2
	ScopeCombat = 0x4  // 4
	ScopeTrade  = 0x8  // 8
	ScopeTodo   = 0x10 // 16
	ScopeAll    = ScopeRead | ScopeWrite | ScopeCombat | ScopeTrade | ScopeTodo
)

// NewToken 创建新 Token
func NewToken(agentID int64, tokenType int, scopes int, expiresAt *int64) *Token {
	return &Token{
		AgentID:   agentID,
		TokenType: tokenType,
		Scopes:    scopes,
		CreatedAt: NowMillis(),
		ExpiresAt: expiresAt,
	}
}
