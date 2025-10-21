package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"HYH-Blog-Gin/internal/config"
	"HYH-Blog-Gin/internal/handlers"
	"HYH-Blog-Gin/internal/middleware"
)

// registerPublicRoutes 注册公开可访问的 API 路由（无鉴权），所有公开路由统一在 /api/v1 前缀下。
func registerPublicRoutes(r *gin.Engine, cfg *config.Config, userHandler *handlers.UserHandler, rdb *redis.Client) {
	v1 := r.Group("/api/v1")
	{
		// 登录限流（按 IP）
		loginRule := cfg.RateLimit.Login
		loginWindow := time.Duration(loginRule.WindowSeconds) * time.Second
		loginLimiter := middleware.RateLimitIP(rdb, "login", loginRule.Limit, loginWindow)

		v1.POST("/register", userHandler.Register)
		v1.POST("/login", loginLimiter, userHandler.Login)
	}
}
