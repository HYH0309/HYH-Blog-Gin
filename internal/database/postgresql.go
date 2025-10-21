package database

import (
	"fmt"
	"log"
	"time"

	"HYH-Blog-Gin/internal/config"
	"HYH-Blog-Gin/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB 包装了 *gorm.DB，便于以后扩展方法。
type DB struct{ *gorm.DB }

// NewDB 使用提供的配置创建并返回一个数据库连接。
// - 构建 Postgres DSN 并打开 GORM 连接。
// - 配置连接池参数并执行 Ping 检查连通性。
// - 可选执行 AutoMigrate（由配置控制）。
// 返回已初始化的 *DB 或发生的错误。
func NewDB(cfg *config.Config) (*DB, error) {
	// 构建 DSN，例如 "host=... user=... password=... dbname=... port=... sslmode=..."
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	// 使用 postgres 驱动打开 gorm 连接
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层 *sql.DB 以配置连接池并执行 Ping 检查
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to access underlying db: %w", err)
	}

	// 连接池参数：根据负载调整这些值
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)

	// Ping 用于快速检测数据库是否可达
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 自动迁移模型以创建/更新表结构（可配置，建议生产禁用）
	if cfg.Database.AutoMigrate {
		if err := db.AutoMigrate(
			&models.User{},
			&models.Note{},
			&models.Tag{},
			&models.Image{},
		); err != nil {
			return nil, fmt.Errorf("failed to migrate database: %w", err)
		}
		log.Println("Database connection established and migrated successfully")
	} else {
		log.Println("Database connection established (AutoMigrate disabled)")
	}
	return &DB{db}, nil
}

// Close 关闭底层数据库连接。
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
