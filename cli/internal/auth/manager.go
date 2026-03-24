package auth

import (
	"crypto/ed25519"
	"fmt"
	"time"

	"github.com/RedContritio/agent-town/cli/internal/client"
	"github.com/RedContritio/agent-town/cli/internal/config"
	"github.com/RedContritio/agent-town/cli/internal/keystore"
)

// Manager 管理认证流程
// 决策：Token 过期检查使用本地时间对比，简单高效。
// 可选：每次请求前调用服务端验证（更安全但增加延迟）
type Manager struct {
	Config *config.Config
}

// NewManager 创建认证管理器
func NewManager(cfg *config.Config) *Manager {
	return &Manager{Config: cfg}
}

// EnsureToken 确保有有效的 Token
// 如果有未过期的 Token 直接返回，否则执行完整认证流程
func (m *Manager) EnsureToken(agentName string, cli *client.Client) (string, error) {
	// 检查现有 Token
	token, expires, hasToken := m.Config.GetToken(agentName)
	if hasToken && !m.isExpired(expires) {
		cli.SetToken(token)
		return token, nil
	}

	// 需要重新认证
	return m.authenticate(agentName, cli)
}

// isExpired 检查 Token 是否过期
// 预留 60 秒缓冲，避免边缘情况
func (m *Manager) isExpired(expires int64) bool {
	return time.Now().Unix() >= expires-60
}

// authenticate 执行完整的 challenge-token 认证流程
func (m *Manager) authenticate(agentName string, cli *client.Client) (string, error) {
	// 1. 加载 Agent 密钥
	key, err := keystore.Load(agentName)
	if err != nil {
		return "", fmt.Errorf("load agent key: %w", err)
	}

	// 2. 获取 challenge
	challenge, err := cli.Auth().Challenge(key.PublicKeyHex())
	if err != nil {
		return "", fmt.Errorf("get challenge: %w", err)
	}

	// 3. 用私钥签名 nonce
	signature := ed25519.Sign(key.PrivateKey, challenge.Nonce)

	// 4. 获取 Token
	tokenResp, err := cli.Auth().Token(&client.TokenRequest{
		ChallengeID: challenge.ChallengeID,
		Signature:   signature,
	})
	if err != nil {
		return "", fmt.Errorf("get token: %w", err)
	}

	// 5. 保存 Token
	if err := m.Config.SaveToken(agentName, tokenResp.Token, tokenResp.ExpiresAt); err != nil {
		return "", fmt.Errorf("save token: %w", err)
	}

	// 6. 设置到客户端
	cli.SetToken(tokenResp.Token)

	return tokenResp.Token, nil
}

// ClearToken 清除指定 Agent 的 Token（用于登出或 Token 无效时）
func (m *Manager) ClearToken(agentName string) error {
	agent, ok := m.Config.GetAgent(agentName)
	if !ok {
		return fmt.Errorf("agent %q not found", agentName)
	}
	agent.Token = ""
	agent.TokenExpires = 0
	return m.Config.Save()
}
