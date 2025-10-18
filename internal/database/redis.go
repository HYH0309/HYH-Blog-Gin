package database

import (
	"context"
	"fmt"
	"time"

	"HYH-Blog-Gin/internal/config"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient 创建并返回一个 Redis 客户端。
// 参数:
//   - cfg: 应用配置，包含 Redis 的主机、端口、密码和 DB 编号。
//
// 行为:
//   - 使用 cfg 构建连接地址并创建客户端实例。
//   - 使用 2 秒超时对 Redis 进行 Ping 检查连通性。
//   - 若 Ping 失败，关闭客户端并返回错误。
//
// 备注:
//   - 超时和错误处理有助于在启动时发现不可用的 Redis 实例，避免后续运行时异常。
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
		return nil, err
	}

	// 成功创建并验证连接后返回客户端
	return client, nil
}
