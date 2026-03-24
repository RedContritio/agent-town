package keystore

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

// AgentKey 存储 Agent 的密钥对
// 决策：使用 ED25519（现代、快速、密钥短）。可选 RSA（兼容旧系统）
type AgentKey struct {
	Name       string
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
	AgentID    string // 从公钥派生
}

// PublicKeyHex 返回公钥的 hex 编码
func (k *AgentKey) PublicKeyHex() string {
	return fmt.Sprintf("ed25519:%x", k.PublicKey)
}

// keystoreDir 返回密钥存储目录
func keystoreDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".at-cli", "agents")
}

// keyPath 返回指定 agent 的密钥文件路径
func keyPath(name string) string {
	return filepath.Join(keystoreDir(), fmt.Sprintf("%s.pem", name))
}

// Create 创建新的 Agent 密钥对
func Create(name string) (*AgentKey, error) {
	dir := keystoreDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create keystore dir: %w", err)
	}

	// 检查是否已存在
	path := keyPath(name)
	if _, err := os.Stat(path); err == nil {
		return nil, fmt.Errorf("agent %q already exists", name)
	}

	// 生成 ED25519 密钥对
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}

	key := &AgentKey{
		Name:       name,
		PrivateKey: privKey,
		PublicKey:  privKey.Public().(ed25519.PublicKey),
		AgentID:    deriveAgentID(privKey.Public().(ed25519.PublicKey)),
	}

	// 保存到文件
	if err := key.Save(); err != nil {
		return nil, err
	}

	return key, nil
}

// Load 加载指定 Agent 的密钥
func Load(name string) (*AgentKey, error) {
	path := keyPath(name)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("agent %q not found", name)
		}
		return nil, fmt.Errorf("read key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM format")
	}

	// ED25519 私钥就是 64 字节种子
	if len(block.Bytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid key size")
	}

	privKey := ed25519.PrivateKey(block.Bytes)

	return &AgentKey{
		Name:       name,
		PrivateKey: privKey,
		PublicKey:  privKey.Public().(ed25519.PublicKey),
		AgentID:    deriveAgentID(privKey.Public().(ed25519.PublicKey)),
	}, nil
}

// Save 保存密钥到文件
func (k *AgentKey) Save() error {
	block := &pem.Block{
		Type:  "ED25519 PRIVATE KEY",
		Bytes: k.PrivateKey,
	}

	path := keyPath(k.Name)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("create key file: %w", err)
	}
	defer file.Close()

	if err := pem.Encode(file, block); err != nil {
		return fmt.Errorf("encode key: %w", err)
	}

	return nil
}

// Delete 删除 Agent 密钥
func Delete(name string) error {
	path := keyPath(name)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("agent %q not found", name)
		}
		return fmt.Errorf("delete key file: %w", err)
	}
	return nil
}

// List 列出所有 Agent 名称
func List() ([]string, error) {
	dir := keystoreDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("read keystore: %w", err)
	}

	names := []string{}
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".pem" {
			name := entry.Name()[:len(entry.Name())-4] // 去掉 .pem
			names = append(names, name)
		}
	}

	return names, nil
}

// Export 导出密钥为 PEM 字符串
func Export(name string) (string, error) {
	key, err := Load(name)
	if err != nil {
		return "", err
	}

	block := &pem.Block{
		Type:  "ED25519 PRIVATE KEY",
		Bytes: key.PrivateKey,
	}

	return string(pem.EncodeToMemory(block)), nil
}

// deriveAgentID 从公钥派生 AgentID
// 决策：使用前 16 字节 hex。可选 base58（更短）或完整 hex（更长）
func deriveAgentID(pubKey ed25519.PublicKey) string {
	return fmt.Sprintf("agent-%x", pubKey[:16])
}
