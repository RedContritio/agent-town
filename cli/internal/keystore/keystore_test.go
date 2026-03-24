package keystore

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDir(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Unsetenv("HOME") })
}

func TestCreate(t *testing.T) {
	setupTestDir(t)

	key, err := Create("Alice")
	require.NoError(t, err)
	assert.NotNil(t, key)
	assert.Equal(t, "Alice", key.Name)
	assert.NotEmpty(t, key.PrivateKey)
	assert.NotEmpty(t, key.PublicKey)
	assert.True(t, strings.HasPrefix(key.AgentID, "agent-"))
	assert.True(t, strings.HasPrefix(key.PublicKeyHex(), "ed25519:"))

	// 验证文件创建
	path := keyPath("Alice")
	_, err = os.Stat(path)
	require.NoError(t, err)
}

func TestCreate_AlreadyExists(t *testing.T) {
	setupTestDir(t)

	_, err := Create("Alice")
	require.NoError(t, err)

	// 再次创建应该失败
	_, err = Create("Alice")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestLoad(t *testing.T) {
	setupTestDir(t)

	// 先创建
	created, err := Create("Alice")
	require.NoError(t, err)

	// 再加载
	loaded, err := Load("Alice")
	require.NoError(t, err)
	assert.Equal(t, created.Name, loaded.Name)
	assert.Equal(t, created.PublicKeyHex(), loaded.PublicKeyHex())
	assert.Equal(t, created.AgentID, loaded.AgentID)
}

func TestLoad_NotExist(t *testing.T) {
	setupTestDir(t)

	_, err := Load("NotExist")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDelete(t *testing.T) {
	setupTestDir(t)

	// 创建
	_, err := Create("Alice")
	require.NoError(t, err)

	// 删除
	err = Delete("Alice")
	require.NoError(t, err)

	// 验证删除
	_, err = os.Stat(keyPath("Alice"))
	assert.True(t, os.IsNotExist(err))
}

func TestDelete_NotExist(t *testing.T) {
	setupTestDir(t)

	err := Delete("NotExist")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestList(t *testing.T) {
	setupTestDir(t)

	// 空列表
	names, err := List()
	require.NoError(t, err)
	assert.Empty(t, names)

	// 创建几个
	Create("Alice")
	Create("Bob")
	Create("Charlie")

	names, err = List()
	require.NoError(t, err)
	assert.Len(t, names, 3)
	assert.Contains(t, names, "Alice")
	assert.Contains(t, names, "Bob")
	assert.Contains(t, names, "Charlie")
}

func TestExport(t *testing.T) {
	setupTestDir(t)

	_, err := Create("Alice")
	require.NoError(t, err)

	pemStr, err := Export("Alice")
	require.NoError(t, err)
	assert.Contains(t, pemStr, "BEGIN ED25519 PRIVATE KEY")
	assert.Contains(t, pemStr, "END ED25519 PRIVATE KEY")
}

func TestExport_NotExist(t *testing.T) {
	setupTestDir(t)

	_, err := Export("NotExist")
	assert.Error(t, err)
}

func TestDeriveAgentID(t *testing.T) {
	setupTestDir(t)

	// 创建两个 key，验证 AgentID 不同
	key1, _ := Create("Alice")
	key2, _ := Create("Bob")

	assert.NotEqual(t, key1.AgentID, key2.AgentID)
	assert.True(t, strings.HasPrefix(key1.AgentID, "agent-"))
	// agent- 前缀 + 32个 hex 字符 (16字节)
	assert.Equal(t, len("agent-")+32, len(key1.AgentID))
}
