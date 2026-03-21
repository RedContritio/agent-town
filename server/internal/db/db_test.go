package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

// setupTestDB 创建测试数据库（内存模式）
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("打开内存数据库失败: %v", err)
	}

	if err := createSchema(db); err != nil {
		t.Fatalf("创建表结构失败: %v", err)
	}

	return db
}

// TestCreateSchema 测试表结构创建
func TestCreateSchema(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 验证表是否存在
	tables := []string{
		"agents", "tokens", "tasks", "inventory",
		"skills", "item_types", "land_ownership", "land_tiles",
		"resources", "buildings",
	}

	for _, table := range tables {
		var count int
		err := db.QueryRow(
			"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?",
			table,
		).Scan(&count)
		if err != nil {
			t.Errorf("检查表 %s 失败: %v", table, err)
			continue
		}
		if count != 1 {
			t.Errorf("表 %s 不存在", table)
		}
	}
}

// TestIndexes 测试索引创建
func TestIndexes(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	indexes := []string{
		"idx_tasks_agent_status",
		"idx_tokens_agent",
		"idx_agents_pos",
		"idx_resources_pos",
		"idx_buildings_owner",
		"idx_buildings_pos",
	}

	for _, idx := range indexes {
		var count int
		err := db.QueryRow(
			"SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?",
			idx,
		).Scan(&count)
		if err != nil {
			t.Errorf("检查索引 %s 失败: %v", idx, err)
			continue
		}
		if count != 1 {
			t.Errorf("索引 %s 不存在", idx)
		}
	}
}

// TestAgentCRUD 测试 Agent 基本 CRUD
func TestAgentCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 插入
	result, err := db.Exec(
		`INSERT INTO agents (public_key, name, balance, created_at, updated_at) 
		 VALUES (?, ?, ?, ?, ?)`,
		[]byte("test-public-key"), "Alice", 100, 1234567890000, 1234567890000,
	)
	if err != nil {
		t.Fatalf("插入 Agent 失败: %v", err)
	}

	id, _ := result.LastInsertId()
	if id != 1 {
		t.Errorf("期望 ID=1, 得到 %d", id)
	}

	// 查询
	var name string
	var balance int
	err = db.QueryRow(
		"SELECT name, balance FROM agents WHERE id=?",
		id,
	).Scan(&name, &balance)
	if err != nil {
		t.Fatalf("查询 Agent 失败: %v", err)
	}
	if name != "Alice" {
		t.Errorf("期望 name=Alice, 得到 %s", name)
	}
	if balance != 100 {
		t.Errorf("期望 balance=100, 得到 %d", balance)
	}

	// 更新
	_, err = db.Exec(
		"UPDATE agents SET balance=? WHERE id=?",
		200, id,
	)
	if err != nil {
		t.Fatalf("更新 Agent 失败: %v", err)
	}

	// 删除
	_, err = db.Exec("DELETE FROM agents WHERE id=?", id)
	if err != nil {
		t.Fatalf("删除 Agent 失败: %v", err)
	}

	// 验证删除
	err = db.QueryRow("SELECT name FROM agents WHERE id=?", id).Scan(&name)
	if err != sql.ErrNoRows {
		t.Errorf("期望记录不存在，得到 %v", err)
	}
}

// TestTokenCRUD 测试 Token CRUD
func TestTokenCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 先插入 Agent（外键约束）
	_, err := db.Exec(
		`INSERT INTO agents (public_key, name, created_at, updated_at) 
		 VALUES (?, ?, ?, ?)`,
		[]byte("test-key"), "Bob", 1234567890000, 1234567890000,
	)
	if err != nil {
		t.Fatalf("插入 Agent 失败: %v", err)
	}

	// 插入 Token
	_, err = db.Exec(
		`INSERT INTO tokens (token, agent_id, token_type, scopes, created_at, expires_at) 
		 VALUES (?, ?, ?, ?, ?, ?)`,
		"test-token-uuid", 1, TokenTypeCLI, ScopeRead|ScopeWrite, 1234567890000, 1234567890000+3600000,
	)
	if err != nil {
		t.Fatalf("插入 Token 失败: %v", err)
	}

	// 查询
	var agentID int
	var scopes int
	err = db.QueryRow(
		"SELECT agent_id, scopes FROM tokens WHERE token=?",
		"test-token-uuid",
	).Scan(&agentID, &scopes)
	if err != nil {
		t.Fatalf("查询 Token 失败: %v", err)
	}
	if agentID != 1 {
		t.Errorf("期望 agent_id=1, 得到 %d", agentID)
	}
	if scopes != ScopeRead|ScopeWrite {
		t.Errorf("期望 scopes=%d, 得到 %d", ScopeRead|ScopeWrite, scopes)
	}

	// 测试 HasScope
	if !HasScope(scopes, ScopeRead) {
		t.Error("期望有 Read 权限")
	}
	if !HasScope(scopes, ScopeWrite) {
		t.Error("期望有 Write 权限")
	}
	if HasScope(scopes, ScopeCombat) {
		t.Error("期望没有 Combat 权限")
	}
}

// TestTaskCRUD 测试 Task CRUD
func TestTaskCRUD(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 先插入 Agent
	_, err := db.Exec(
		`INSERT INTO agents (public_key, name, created_at, updated_at) 
		 VALUES (?, ?, ?, ?)`,
		[]byte("test-key"), "Charlie", 1234567890000, 1234567890000,
	)
	if err != nil {
		t.Fatalf("插入 Agent 失败: %v", err)
	}

	// 插入任务
	_, err = db.Exec(
		`INSERT INTO tasks (agent_id, seq, type, status, params, stack_depth, created_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		1, 1, TaskTypeMove, TaskStatusRunning, `{"dx":3,"dy":0}`, 0, 1234567890000,
	)
	if err != nil {
		t.Fatalf("插入 Task 失败: %v", err)
	}

	// 查询
	var taskType, status, depth int
	err = db.QueryRow(
		"SELECT type, status, stack_depth FROM tasks WHERE agent_id=? AND seq=?",
		1, 1,
	).Scan(&taskType, &status, &depth)
	if err != nil {
		t.Fatalf("查询 Task 失败: %v", err)
	}
	if taskType != TaskTypeMove {
		t.Errorf("期望 type=%d, 得到 %d", TaskTypeMove, taskType)
	}
	if status != TaskStatusRunning {
		t.Errorf("期望 status=%d, 得到 %d", TaskStatusRunning, status)
	}
}

// TestForeignKey 测试外键约束
func TestForeignKey(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 确保外键已启用
	var fkEnabled int
	err := db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled)
	if err != nil {
		t.Fatalf("检查外键状态失败: %v", err)
	}
	if fkEnabled != 1 {
		t.Log("外键未启用，尝试启用")
		_, err = db.Exec("PRAGMA foreign_keys = ON")
		if err != nil {
			t.Fatalf("启用外键失败: %v", err)
		}
	}

	// 尝试插入无 Agent 的 Token（应该失败）
	_, err = db.Exec(
		`INSERT INTO tokens (token, agent_id, scopes, created_at) 
		 VALUES (?, ?, ?, ?)`,
		"invalid-token", 999, ScopeRead, 1234567890000,
	)
	if err == nil {
		t.Error("期望外键约束失败，但没有错误")
	}
}

// TestInitWithFile 测试文件数据库初始化
func TestInitWithFile(t *testing.T) {
	// 临时目录
	tmpDir := t.TempDir()

	config := &Config{
		DataDir: tmpDir,
		DBName:  "test.db",
	}

	db, err := Init(config)
	if err != nil {
		t.Fatalf("初始化数据库失败: %v", err)
	}
	defer db.Close()
	defer os.RemoveAll(tmpDir)

	// 验证文件存在
	dbPath := filepath.Join(tmpDir, "test.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("数据库文件未创建")
	}

	// 验证可以写入
	_, err = db.Exec(
		`INSERT INTO agents (public_key, name, created_at, updated_at) 
		 VALUES (?, ?, ?, ?)`,
		[]byte("file-test"), "FileAgent", 1234567890000, 1234567890000,
	)
	if err != nil {
		t.Fatalf("文件数据库写入失败: %v", err)
	}
}

// TestScopeOperations 测试权限位操作
func TestScopeOperations(t *testing.T) {
	scopes := 0

	// 添加权限
	scopes = AddScope(scopes, ScopeRead)
	if !HasScope(scopes, ScopeRead) {
		t.Error("添加 Read 权限失败")
	}

	scopes = AddScope(scopes, ScopeWrite)
	if !HasScope(scopes, ScopeWrite) {
		t.Error("添加 Write 权限失败")
	}

	// 验证多个权限
	if scopes != ScopeRead|ScopeWrite {
		t.Errorf("期望 scopes=%d, 得到 %d", ScopeRead|ScopeWrite, scopes)
	}

	// 移除权限
	scopes = RemoveScope(scopes, ScopeRead)
	if HasScope(scopes, ScopeRead) {
		t.Error("移除 Read 权限失败")
	}
	if !HasScope(scopes, ScopeWrite) {
		t.Error("Write 权限不应被移除")
	}
}

// TestUniqueConstraints 测试唯一约束
func TestUniqueConstraints(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// 插入第一个 Agent
	_, err := db.Exec(
		`INSERT INTO agents (public_key, name, created_at, updated_at) 
		 VALUES (?, ?, ?, ?)`,
		[]byte("unique-key-1"), "UniqueAgent", 1234567890000, 1234567890000,
	)
	if err != nil {
		t.Fatalf("插入第一个 Agent 失败: %v", err)
	}

	// 尝试插入同名 Agent（应该失败）
	_, err = db.Exec(
		`INSERT INTO agents (public_key, name, created_at, updated_at) 
		 VALUES (?, ?, ?, ?)`,
		[]byte("unique-key-2"), "UniqueAgent", 1234567890000, 1234567890000,
	)
	if err == nil {
		t.Error("期望唯一约束失败（name），但没有错误")
	}

	// 尝试插入相同公钥 Agent（应该失败）
	_, err = db.Exec(
		`INSERT INTO agents (public_key, name, created_at, updated_at) 
		 VALUES (?, ?, ?, ?)`,
		[]byte("unique-key-1"), "AnotherAgent", 1234567890000, 1234567890000,
	)
	if err == nil {
		t.Error("期望唯一约束失败（public_key），但没有错误")
	}
}
