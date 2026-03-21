package service

import (
	"crypto/ed25519"
	"encoding/base64"
	"testing"

	"github.com/RedContritio/agent-town/server/internal/db"
	"github.com/RedContritio/agent-town/server/internal/model"
	"github.com/RedContritio/agent-town/server/internal/repository"
	_ "modernc.org/sqlite"
)

// setupTestService 创建测试服务
func setupTestService(t *testing.T) (*AuthService, *repository.AgentRepository, *repository.TokenRepository) {
	database, err := db.Init(&db.Config{DataDir: ":memory:", DBName: ""})
	if err != nil {
		t.Fatalf("初始化数据库失败: %v", err)
	}

	agentRepo := repository.NewAgentRepository(database)
	tokenRepo := repository.NewTokenRepository(database)
	service := NewAuthService(agentRepo, tokenRepo)

	return service, agentRepo, tokenRepo
}

// generateTestKeyPair 生成测试密钥对
func generateTestKeyPair(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("生成密钥对失败: %v", err)
	}
	return publicKey, privateKey
}

func TestAuthService_Register(t *testing.T) {
	service, _, _ := setupTestService(t)

	publicKey, _ := generateTestKeyPair(t)
	publicKeyB64 := base64.StdEncoding.EncodeToString(publicKey)

	// 测试成功注册
	req := &RegisterRequest{
		PublicKey: publicKeyB64,
		Name:      "TestAgent",
	}
	resp, err := service.Register(req)
	if err != nil {
		t.Fatalf("注册失败: %v", err)
	}
	if resp.AgentID != 1 {
		t.Errorf("期望 AgentID=1, 得到 %d", resp.AgentID)
	}
	if resp.Name != "TestAgent" {
		t.Errorf("期望 Name=TestAgent, 得到 %s", resp.Name)
	}

	// 测试重复名称
	publicKey2, _ := generateTestKeyPair(t)
	req2 := &RegisterRequest{
		PublicKey: base64.StdEncoding.EncodeToString(publicKey2),
		Name:      "TestAgent",
	}
	_, err = service.Register(req2)
	if err == nil {
		t.Error("期望重复名称失败，但没有错误")
	}

	// 测试重复公钥
	req3 := &RegisterRequest{
		PublicKey: publicKeyB64,
		Name:      "AnotherName",
	}
	_, err = service.Register(req3)
	if err == nil {
		t.Error("期望重复公钥失败，但没有错误")
	}

	// 测试空名称
	req4 := &RegisterRequest{
		PublicKey: publicKeyB64,
		Name:      "",
	}
	_, err = service.Register(req4)
	if err == nil {
		t.Error("期望空名称失败，但没有错误")
	}
}

func TestAuthService_CreateChallenge(t *testing.T) {
	service, _, _ := setupTestService(t)
	publicKey, _ := generateTestKeyPair(t)
	publicKeyB64 := base64.StdEncoding.EncodeToString(publicKey)

	// 先注册
	_, err := service.Register(&RegisterRequest{
		PublicKey: publicKeyB64,
		Name:      "ChallengeTest",
	})
	if err != nil {
		t.Fatalf("注册失败: %v", err)
	}

	// 创建挑战
	req := &ChallengeRequest{PublicKey: publicKeyB64}
	resp, err := service.CreateChallenge(req)
	if err != nil {
		t.Fatalf("创建挑战失败: %v", err)
	}
	if resp.ChallengeID == "" {
		t.Error("ChallengeID 为空")
	}
	if resp.Challenge == "" {
		t.Error("Challenge 为空")
	}
	if resp.ExpiresAt <= model.NowMillis() {
		t.Error("ExpiresAt 无效")
	}

	// 测试未注册的公钥
	publicKey2, _ := generateTestKeyPair(t)
	req2 := &ChallengeRequest{
		PublicKey: base64.StdEncoding.EncodeToString(publicKey2),
	}
	_, err = service.CreateChallenge(req2)
	if err == nil {
		t.Error("期望未注册公钥失败，但没有错误")
	}
}

func TestAuthService_CreateToken(t *testing.T) {
	service, _, _ := setupTestService(t)
	publicKey, privateKey := generateTestKeyPair(t)
	publicKeyB64 := base64.StdEncoding.EncodeToString(publicKey)

	// 先注册
	_, err := service.Register(&RegisterRequest{
		PublicKey: publicKeyB64,
		Name:      "TokenTest",
	})
	if err != nil {
		t.Fatalf("注册失败: %v", err)
	}

	// 创建挑战
	challengeResp, err := service.CreateChallenge(&ChallengeRequest{PublicKey: publicKeyB64})
	if err != nil {
		t.Fatalf("创建挑战失败: %v", err)
	}

	// 签名挑战
	signature := ed25519.Sign(privateKey, []byte(challengeResp.Challenge))
	signatureB64 := base64.StdEncoding.EncodeToString(signature)

	// 获取 Token
	req := &TokenRequest{
		ChallengeID: challengeResp.ChallengeID,
		Signature:   signatureB64,
	}
	tokenResp, err := service.CreateToken(req)
	if err != nil {
		t.Fatalf("创建 Token 失败: %v", err)
	}
	if tokenResp.Token == "" {
		t.Error("Token 为空")
	}
	if tokenResp.Scopes != model.ScopeAll {
		t.Errorf("期望 Scopes=%d, 得到 %d", model.ScopeAll, tokenResp.Scopes)
	}
	if tokenResp.ExpiresAt == nil {
		t.Error("ExpiresAt 为空")
	}

	// 测试无效签名
	_, wrongPrivateKey := generateTestKeyPair(t)
	wrongSignature := ed25519.Sign(wrongPrivateKey, []byte(challengeResp.Challenge))
	req2 := &TokenRequest{
		ChallengeID: challengeResp.ChallengeID,
		Signature:   base64.StdEncoding.EncodeToString(wrongSignature),
	}
	_, err = service.CreateToken(req2)
	if err == nil {
		t.Error("期望无效签名失败，但没有错误")
	}

	// 测试无效 challenge_id
	req3 := &TokenRequest{
		ChallengeID: "invalid-id",
		Signature:   signatureB64,
	}
	_, err = service.CreateToken(req3)
	if err == nil {
		t.Error("期望无效 challenge_id 失败，但没有错误")
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	service, _, _ := setupTestService(t)
	publicKey, privateKey := generateTestKeyPair(t)
	publicKeyB64 := base64.StdEncoding.EncodeToString(publicKey)

	// 注册并获取 Token
	_, err := service.Register(&RegisterRequest{
		PublicKey: publicKeyB64,
		Name:      "ValidateTest",
	})
	if err != nil {
		t.Fatalf("注册失败: %v", err)
	}

	challengeResp, _ := service.CreateChallenge(&ChallengeRequest{PublicKey: publicKeyB64})
	signature := ed25519.Sign(privateKey, []byte(challengeResp.Challenge))
	tokenResp, _ := service.CreateToken(&TokenRequest{
		ChallengeID: challengeResp.ChallengeID,
		Signature:   base64.StdEncoding.EncodeToString(signature),
	})

	// 验证有效 Token
	token, err := service.ValidateToken(tokenResp.Token)
	if err != nil {
		t.Fatalf("验证 Token 失败: %v", err)
	}
	if token.Token != tokenResp.Token {
		t.Error("Token 不匹配")
	}

	// 验证无效 Token
	_, err = service.ValidateToken("invalid-token")
	if err == nil {
		t.Error("期望无效 Token 失败，但没有错误")
	}

	// 验证 Token 权限
	if !token.HasScope(model.ScopeRead) {
		t.Error("Token 应该有 Read 权限")
	}
	if !token.HasScope(model.ScopeWrite) {
		t.Error("Token 应该有 Write 权限")
	}
}

func TestAuthService_Logout(t *testing.T) {
	service, _, _ := setupTestService(t)
	publicKey, privateKey := generateTestKeyPair(t)
	publicKeyB64 := base64.StdEncoding.EncodeToString(publicKey)

	// 注册并获取 Token
	service.Register(&RegisterRequest{
		PublicKey: publicKeyB64,
		Name:      "LogoutTest",
	})
	challengeResp, _ := service.CreateChallenge(&ChallengeRequest{PublicKey: publicKeyB64})
	signature := ed25519.Sign(privateKey, []byte(challengeResp.Challenge))
	tokenResp, _ := service.CreateToken(&TokenRequest{
		ChallengeID: challengeResp.ChallengeID,
		Signature:   base64.StdEncoding.EncodeToString(signature),
	})

	// 登出
	err := service.Logout(tokenResp.Token)
	if err != nil {
		t.Fatalf("登出失败: %v", err)
	}

	// 验证 Token 已失效
	_, err = service.ValidateToken(tokenResp.Token)
	if err == nil {
		t.Error("期望 Token 已失效，但没有错误")
	}
}

func TestAuthService_FullFlow(t *testing.T) {
	service, _, _ := setupTestService(t)
	publicKey, privateKey := generateTestKeyPair(t)
	publicKeyB64 := base64.StdEncoding.EncodeToString(publicKey)

	// 1. 注册
	registerResp, err := service.Register(&RegisterRequest{
		PublicKey: publicKeyB64,
		Name:      "FullFlowTest",
	})
	if err != nil {
		t.Fatalf("注册失败: %v", err)
	}
	t.Logf("注册成功: AgentID=%d", registerResp.AgentID)

	// 2. 创建挑战
	challengeResp, err := service.CreateChallenge(&ChallengeRequest{PublicKey: publicKeyB64})
	if err != nil {
		t.Fatalf("创建挑战失败: %v", err)
	}
	t.Logf("挑战创建: ID=%s", challengeResp.ChallengeID)

	// 3. 签名
	signature := ed25519.Sign(privateKey, []byte(challengeResp.Challenge))
	signatureB64 := base64.StdEncoding.EncodeToString(signature)

	// 4. 获取 Token
	tokenResp, err := service.CreateToken(&TokenRequest{
		ChallengeID: challengeResp.ChallengeID,
		Signature:   signatureB64,
	})
	if err != nil {
		t.Fatalf("获取 Token 失败: %v", err)
	}
	t.Logf("Token 获取: %s...", tokenResp.Token[:8])

	// 5. 验证 Token
	token, err := service.ValidateToken(tokenResp.Token)
	if err != nil {
		t.Fatalf("验证 Token 失败: %v", err)
	}
	if token.AgentID != registerResp.AgentID {
		t.Error("Token 的 AgentID 不匹配")
	}

	// 6. 登出
	err = service.Logout(tokenResp.Token)
	if err != nil {
		t.Fatalf("登出失败: %v", err)
	}

	// 7. 验证已登出
	_, err = service.ValidateToken(tokenResp.Token)
	if err == nil {
		t.Error("期望 Token 已失效")
	}
}
