// Package db 提供数据库连接和初始化功能
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB 全局数据库连接实例
var DB *sql.DB

// Config 数据库配置
type Config struct {
	DataDir  string // 数据目录，默认为 "./data"
	DBName   string // 数据库文件名，默认为 "agent_town.db"
	ReadOnly bool   // 是否只读模式
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		DataDir:  "./data",
		DBName:   "agent_town.db",
		ReadOnly: false,
	}
}

// Init 初始化数据库连接
// 自动创建数据目录和数据库文件（如果不存在）
func Init(config *Config) (*sql.DB, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 确保数据目录存在
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	dbPath := filepath.Join(config.DataDir, config.DBName)

	// SQLite 连接字符串
	// _journal_mode=WAL: 使用 WAL 模式提升并发性能
	// _busy_timeout=5000: 忙等待超时 5 秒
	// _foreign_keys=on: 启用外键约束
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on", dbPath)
	if config.ReadOnly {
		dsn += "&mode=ro"
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	// 验证连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// 创建表结构
	if err := createSchema(db); err != nil {
		return nil, fmt.Errorf("创建表结构失败: %w", err)
	}

	DB = db
	return db, nil
}

// Close 关闭数据库连接
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// WithTx 在事务中执行函数
func WithTx(fn func(*sql.Tx) error) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
