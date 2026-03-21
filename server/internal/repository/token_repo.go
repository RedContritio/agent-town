package repository

import (
	"database/sql"
	"errors"

	"github.com/RedContritio/agent-town/server/internal/model"
	"github.com/google/uuid"
)

// TokenRepository Token 数据访问
type TokenRepository struct {
	db *sql.DB
}

// NewTokenRepository 创建 TokenRepository
func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

// Create 创建 Token
func (r *TokenRepository) Create(token *model.Token) error {
	// 如果没有指定 token 值，生成 UUID
	if token.Token == "" {
		token.Token = uuid.New().String()
	}

	var expiresAt interface{}
	if token.ExpiresAt != nil {
		expiresAt = *token.ExpiresAt
	} else {
		expiresAt = nil
	}

	_, err := r.db.Exec(
		`INSERT INTO tokens (token, agent_id, token_type, scopes, created_at, expires_at, last_used_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		token.Token, token.AgentID, token.TokenType, token.Scopes,
		token.CreatedAt, expiresAt, token.LastUsedAt,
	)
	return err
}

// GetByToken 根据 token 字符串获取
func (r *TokenRepository) GetByToken(tokenStr string) (*model.Token, error) {
	token := &model.Token{}
	var expiresAt sql.NullInt64
	var lastUsedAt sql.NullInt64

	err := r.db.QueryRow(
		`SELECT token, agent_id, token_type, scopes, created_at, expires_at, last_used_at
		 FROM tokens WHERE token=?`,
		tokenStr,
	).Scan(
		&token.Token, &token.AgentID, &token.TokenType, &token.Scopes,
		&token.CreatedAt, &expiresAt, &lastUsedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if expiresAt.Valid {
		token.ExpiresAt = &expiresAt.Int64
	}
	if lastUsedAt.Valid {
		token.LastUsedAt = &lastUsedAt.Int64
	}

	return token, nil
}

// GetByAgentID 获取 Agent 的所有 Token
func (r *TokenRepository) GetByAgentID(agentID int64) ([]*model.Token, error) {
	rows, err := r.db.Query(
		`SELECT token, agent_id, token_type, scopes, created_at, expires_at, last_used_at
		 FROM tokens WHERE agent_id=?`,
		agentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*model.Token
	for rows.Next() {
		token := &model.Token{}
		var expiresAt sql.NullInt64
		var lastUsedAt sql.NullInt64

		err := rows.Scan(
			&token.Token, &token.AgentID, &token.TokenType, &token.Scopes,
			&token.CreatedAt, &expiresAt, &lastUsedAt,
		)
		if err != nil {
			return nil, err
		}

		if expiresAt.Valid {
			token.ExpiresAt = &expiresAt.Int64
		}
		if lastUsedAt.Valid {
			token.LastUsedAt = &lastUsedAt.Int64
		}

		tokens = append(tokens, token)
	}

	return tokens, rows.Err()
}

// UpdateLastUsed 更新最后使用时间
func (r *TokenRepository) UpdateLastUsed(tokenStr string) error {
	_, err := r.db.Exec(
		"UPDATE tokens SET last_used_at=? WHERE token=?",
		model.NowMillis(), tokenStr,
	)
	return err
}

// Delete 删除 Token
func (r *TokenRepository) Delete(tokenStr string) error {
	_, err := r.db.Exec("DELETE FROM tokens WHERE token=?", tokenStr)
	return err
}

// DeleteByAgentID 删除 Agent 的所有 Token
func (r *TokenRepository) DeleteByAgentID(agentID int64) error {
	_, err := r.db.Exec("DELETE FROM tokens WHERE agent_id=?", agentID)
	return err
}

// DeleteExpired 删除过期 Token
func (r *TokenRepository) DeleteExpired() error {
	_, err := r.db.Exec(
		"DELETE FROM tokens WHERE expires_at IS NOT NULL AND expires_at<?",
		model.NowMillis(),
	)
	return err
}
