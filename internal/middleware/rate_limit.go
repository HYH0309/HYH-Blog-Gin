package middleware

import (
	"context"
	"fmt"
	"time"

	"HYH-Blog-Gin/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// rateLimitKey 生成限流键，形如 rl:<action>:<scope>:<id>
func rateLimitKey(action, scope, id string) string {
	return fmt.Sprintf("rl:%s:%s:%s", action, scope, id)
}

// incrWithTTL 在首次命中时设置过期时间，不会在后续请求刷新 TTL
func incrWithTTL(ctx context.Context, rdb *redis.Client, key string, window time.Duration) (int64, error) {
	val, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if val == 1 {
		// 首次创建计数器，设置过期时间
		_ = rdb.Expire(ctx, key, window).Err()
	}
	return val, nil
}

// RateLimitIP 按 IP 地址对某 action 进行限流
func RateLimitIP(rdb *redis.Client, action string, limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rdb == nil || limit <= 0 || window <= 0 {
			c.Next()
			return
		}
		ip := c.ClientIP()
		key := rateLimitKey(action, "ip", ip)
		ctx, cancel := context.WithTimeout(c.Request.Context(), 200*time.Millisecond)
		defer cancel()
		count, err := incrWithTTL(ctx, rdb, key, window)
		if err != nil {
			// 失败时不阻断请求，但记录到 gin 错误
			_ = c.Error(err)
			c.Next()
			return
		}
		if count > limit {
			utils.TooManyRequests(c, "rate limit exceeded")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RateLimitUser 优先使用用户ID限流；若无用户上下文则回退到 IP 限流
func RateLimitUser(rdb *redis.Client, action string, limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rdb == nil || limit <= 0 || window <= 0 {
			c.Next()
			return
		}
		// 尝试从上下文获取用户ID
		if uid, ok := utils.GetUserIDFromContext(c); ok && uid > 0 {
			key := rateLimitKey(action, "uid", fmt.Sprint(uid))
			ctx, cancel := context.WithTimeout(c.Request.Context(), 200*time.Millisecond)
			defer cancel()
			count, err := incrWithTTL(ctx, rdb, key, window)
			if err != nil {
				_ = c.Error(err)
			} else if count > limit {
				utils.TooManyRequests(c, "rate limit exceeded")
				c.Abort()
				return
			}
			c.Next()
			return
		}
		// 回退到 IP 限流
		RateLimitIP(rdb, action, limit, window)(c)
	}
}
