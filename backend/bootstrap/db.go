package bootstrap

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/pudongping/wx-graph-crawl/backend/configs"
	"github.com/pudongping/wx-graph-crawl/backend/utils"
)

var (
	DB   *sqlx.DB
	once sync.Once
)

func initDB(dbPath string, cfg *configs.Config) (db *sqlx.DB, err error) {
	if err = utils.MkdirIfNotExist(filepath.Dir(dbPath)); err != nil {
		return nil, errors.Wrap(err, "创建数据库目录失败")
	}

	// 连接数据库
	db, err = sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, errors.Wrap(err, "连接数据库失败")
	}

	// 初始化所有数据表
	if err = initTables(db); err != nil {
		return nil, errors.Wrap(err, "初始化数据表失败")
	}

	// 设置连接池参数
	db.SetMaxOpenConns(cfg.DB.MaxOpenConnections)               // 设置最大打开连接数
	db.SetMaxIdleConns(cfg.DB.MaxIdleConnections)               // 设置最大空闲连接数
	db.SetConnMaxLifetime(time.Duration(cfg.DB.MaxLifeSeconds)) // 设置连接的最大生命周期，0表示不限制

	return db, nil
}

// InitDB 初始化数据库连接
func InitDB(dbPath string, cfg *configs.Config) (db *sqlx.DB, err error) {
	once.Do(func() {
		db, err = initDB(dbPath, cfg)
	})

	DB = db
	return DB, err
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// initTables 初始化所有数据表
func initTables(db *sqlx.DB) error {
	var err error
	defer func() {
		if err != nil {
			_ = db.Close() // 关闭数据库连接
		}
	}()

	// 创建系统配置表
	if err = createSystemConfigsTable(db); err != nil {
		return errors.Wrap(err, "创建系统配置表失败")
	}

	return nil
}

// createSystemConfigsTable 创建系统配置表
func createSystemConfigsTable(db *sqlx.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS system_configs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT NOT NULL UNIQUE,
			content TEXT NOT NULL DEFAULT '',
			version INTEGER NOT NULL DEFAULT 1,
			created_at INTEGER NOT NULL DEFAULT 0,
			updated_at INTEGER NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS idx_version_updated_at ON system_configs (version, updated_at);
	`)
	return err
}
