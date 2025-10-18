package router

import (
	"github.com/gin-gonic/gin"

	"HYH-Blog-Gin/internal/handlers"
)

// registerPublicRoutes 注册公开可访问的 API 路由（无鉴权），所有公开路由统一在 /api/v1 前缀下。
func registerPublicRoutes(r *gin.Engine, userHandler *handlers.UserHandler) {
	v1 := r.Group("/api/v1")
	{
		v1.POST("/register", userHandler.Register)
		v1.POST("/login", userHandler.Login)
	}
}
