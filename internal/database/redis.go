package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"HYH-Blog-Gin/internal/config"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient 创建并返回一个 Redis 客户端。
// 如果 Redis 启用了认证但配置中未提供密码，会返回更友好的错误提示。
func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	// 构建 Redis 地址，如 "host:port"
	addr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)

	// 创建 Redis 客户端实例
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 在短时间内检测 Redis 是否可用，避免阻塞启动流程
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Ping 用于检查连接是否成功；若失败则关闭客户端并返回错误
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		// 如果服务端要求 AUTH 而配置中没有提供 password，给出更明确的提示
		if strings.Contains(strings.ToLower(err.Error()), "noauth") || strings.Contains(strings.ToLower(err.Error()), "authentication required") {
			if cfg.Redis.Password == "" {
				return nil, fmt.Errorf("redis ping failed: %v; server requires AUTH but REDIS_PASSWORD is empty - set REDIS_PASSWORD in .env or Redis configuration", err)
			}
		}
		return nil, err
	}

	// 成功创建并验证连接后返回客户端
	return client, nil
}
