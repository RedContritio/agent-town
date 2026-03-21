package model

import (
	"crypto/rand"
	"encoding/hex"
)

// Challenge 认证挑战
type Challenge struct {
	ID        string `json:"challengeId"`
	PublicKey []byte `json:"-"`
	Challenge string `json:"challenge"`
	CreatedAt int64  `json:"createdAt"`
	ExpiresAt int64  `json:"expiresAt"`
}

// IsExpired 检查挑战是否过期
func (c *Challenge) IsExpired() bool {
	return NowMillis() > c.ExpiresAt
}

// GenerateChallenge 生成随机挑战字符串
func GenerateChallenge() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// NewChallenge 创建新挑战
// ttlMs: 过期时间（毫秒）
func NewChallenge(publicKey []byte, ttlMs int64) (*Challenge, error) {
	challenge, err := GenerateChallenge()
	if err != nil {
		return nil, err
	}

	now := NowMillis()
	id, _ := GenerateChallenge() // 复用生成函数作为 ID

	return &Challenge{
		ID:        id,
		PublicKey: publicKey,
		Challenge: challenge,
		CreatedAt: now,
		ExpiresAt: now + ttlMs,
	}, nil
}
