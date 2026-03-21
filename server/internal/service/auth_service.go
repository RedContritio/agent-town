// Package service 业务逻辑层
package service

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/RedContritio/agent-town/server/internal/model"
	"github.com/RedContritio/agent-town/server/internal/repository"
)

// AuthService 认证服务
type AuthService struct {
	agentRepo  *repository.AgentRepository
	tokenRepo  *repository.TokenRepository
	challenges map[string]*model.Challenge // challenge_id -> Challenge
	challengeMu sync.RWMutex
	challengeTTL int64 // 挑战过期时间（毫秒）
}

// NewAuthService 创建认证服务
func NewAuthService(agentRepo *repository.AgentRepository, tokenRepo *repository.TokenRepository) *AuthService {
	service := &AuthService{
		agentRepo:    agentRepo,
		tokenRepo:    tokenRepo,
		challenges:   make(map[string]*model.Challenge),
		challengeTTL: 5 * 60 * 1000, // 5 分钟
	}
	// 启动清理 goroutine
	go service.cleanupLoop()
	return service
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	PublicKey string `json:"public_key"`
	Name      string `json:"name"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	AgentID int64  `json:"agent_id"`
	Name    string `json:"name"`
}

// Register 注册新 Agent
func (s *AuthService) Register(req *RegisterRequest) (*RegisterResponse, error) {
	// 验证参数
	if req.Name == "" {
		return nil, errors.New("name 不能为空")
	}
	if req.PublicKey == "" {
		return nil, errors.New("public_key 不能为空")
	}

	// 解码公钥
	publicKey, err := base64.StdEncoding.DecodeString(req.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("public_key 格式错误: %w", err)
	}

	// 验证公钥长度（ed25519 公钥为 32 字节）
	if len(publicKey) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("public_key 长度错误，期望 %d 字节，实际 %d 字节", 
			ed25519.PublicKeySize, len(publicKey))
	}

	// 检查名称是否已存在
	exists, err := s.agentRepo.ExistsByName(req.Name)
	if err != nil {
		return nil, fmt.Errorf("检查名称失败: %w", err)
	}
	if exists {
		return nil, errors.New("name 已存在")
	}

	// 检查公钥是否已存在
	exists, err = s.agentRepo.ExistsByPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("检查公钥失败: %w", err)
	}
	if exists {
		return nil, errors.New("public_key 已注册")
	}

	// 创建 Agent
	agent := model.NewAgent(req.Name, publicKey)
	if err := s.agentRepo.Create(agent); err != nil {
		return nil, fmt.Errorf("创建 Agent 失败: %w", err)
	}

	return &RegisterResponse{
		AgentID: agent.ID,
		Name:    agent.Name,
	}, nil
}

// ChallengeRequest 挑战请求
type ChallengeRequest struct {
	PublicKey string `json:"public_key"`
}

// ChallengeResponse 挑战响应
type ChallengeResponse struct {
	ChallengeID string `json:"challenge_id"`
	Challenge   string `json:"challenge"`
	ExpiresAt   int64  `json:"expires_at"`
}

// CreateChallenge 创建认证挑战
func (s *AuthService) CreateChallenge(req *ChallengeRequest) (*ChallengeResponse, error) {
	if req.PublicKey == "" {
		return nil, errors.New("public_key 不能为空")
	}

	// 解码公钥
	publicKey, err := base64.StdEncoding.DecodeString(req.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("public_key 格式错误: %w", err)
	}

	// 验证 Agent 存在
	agent, err := s.agentRepo.GetByPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("查询 Agent 失败: %w", err)
	}
	if agent == nil {
		return nil, errors.New("public_key 未注册")
	}

	// 创建挑战
	challenge, err := model.NewChallenge(publicKey, s.challengeTTL)
	if err != nil {
		return nil, fmt.Errorf("创建挑战失败: %w", err)
	}

	// 存储挑战
	s.challengeMu.Lock()
	s.challenges[challenge.ID] = challenge
	s.challengeMu.Unlock()

	return &ChallengeResponse{
		ChallengeID: challenge.ID,
		Challenge:   challenge.Challenge,
		ExpiresAt:   challenge.ExpiresAt,
	}, nil
}

// TokenRequest Token 请求
type TokenRequest struct {
	ChallengeID string `json:"challenge_id"`
	Signature   string `json:"signature"`
}

// TokenResponse Token 响应
type TokenResponse struct {
	Token     string `json:"token"`
	TokenType int    `json:"token_type"`
	Scopes    int    `json:"scopes"`
	ExpiresAt *int64 `json:"expires_at,omitempty"`
}

// CreateToken 创建 Token
func (s *AuthService) CreateToken(req *TokenRequest) (*TokenResponse, error) {
	if req.ChallengeID == "" {
		return nil, errors.New("challenge_id 不能为空")
	}
	if req.Signature == "" {
		return nil, errors.New("signature 不能为空")
	}

	// 获取挑战
	s.challengeMu.RLock()
	challenge, ok := s.challenges[req.ChallengeID]
	s.challengeMu.RUnlock()

	if !ok {
		return nil, errors.New("challenge_id 无效或已过期")
	}

	// 检查挑战是否过期
	if challenge.IsExpired() {
		// 清理过期挑战
		s.challengeMu.Lock()
		delete(s.challenges, req.ChallengeID)
		s.challengeMu.Unlock()
		return nil, errors.New("challenge 已过期")
	}

	// 解码签名
	signature, err := base64.StdEncoding.DecodeString(req.Signature)
	if err != nil {
		return nil, fmt.Errorf("signature 格式错误: %w", err)
	}

	// 验证签名
	if !ed25519.Verify(challenge.PublicKey, []byte(challenge.Challenge), signature) {
		return nil, errors.New("signature 验证失败")
	}

	// 获取 Agent
	agent, err := s.agentRepo.GetByPublicKey(challenge.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("查询 Agent 失败: %w", err)
	}
	if agent == nil {
		return nil, errors.New("Agent 不存在")
	}

	// 清理已使用的挑战
	s.challengeMu.Lock()
	delete(s.challenges, req.ChallengeID)
	s.challengeMu.Unlock()

	// 创建 Token（30 天有效期）
	expiresAt := model.NowMillis() + 30*24*60*60*1000
	token := model.NewToken(agent.ID, model.TokenTypeCLI, model.ScopeAll, &expiresAt)
	if err := s.tokenRepo.Create(token); err != nil {
		return nil, fmt.Errorf("创建 Token 失败: %w", err)
	}

	return &TokenResponse{
		Token:     token.Token,
		TokenType: token.TokenType,
		Scopes:    token.Scopes,
		ExpiresAt: token.ExpiresAt,
	}, nil
}

// ValidateToken 验证 Token
func (s *AuthService) ValidateToken(tokenStr string) (*model.Token, error) {
	token, err := s.tokenRepo.GetByToken(tokenStr)
	if err != nil {
		return nil, fmt.Errorf("查询 Token 失败: %w", err)
	}
	if token == nil {
		return nil, errors.New("token 无效")
	}
	if token.IsExpired() {
		return nil, errors.New("token 已过期")
	}

	// 更新最后使用时间
	s.tokenRepo.UpdateLastUsed(tokenStr)
	token.UpdateLastUsed()

	return token, nil
}

// Logout 登出（删除 Token）
func (s *AuthService) Logout(tokenStr string) error {
	return s.tokenRepo.Delete(tokenStr)
}

// cleanupLoop 定期清理过期挑战
func (s *AuthService) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.challengeMu.Lock()
		now := model.NowMillis()
		for id, ch := range s.challenges {
			if now > ch.ExpiresAt {
				delete(s.challenges, id)
			}
		}
		s.challengeMu.Unlock()
	}
}
