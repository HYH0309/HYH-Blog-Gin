package router

import (
	"github.com/gin-gonic/gin"

	"HYH-Blog-Gin/internal/auth"
	"HYH-Blog-Gin/internal/handlers"
	"HYH-Blog-Gin/internal/middleware"
	"HYH-Blog-Gin/internal/utils"
)

// SetupRouter 构建并返回 Gin 引擎，集中注册中间件与路由。
func SetupRouter(jwt *auth.JWTService, userHandler *handlers.UserHandler, noteHandler *handlers.NoteHandler) *gin.Engine {
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORSMiddleware())

	// 公开路由
	public := r.Group("/api")
	{
		public.POST("/register", userHandler.Register)
		public.POST("/login", userHandler.Login)
	}

	// 需要认证的路由
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware(jwt))
	{
		// 用户相关
		protected.GET("/user/profile", userHandler.GetProfile)

		// 笔记相关
		protected.GET("/notes", noteHandler.GetNotes)
		protected.POST("/notes", noteHandler.CreateNote)
		protected.GET("/notes/:id", noteHandler.GetNote)
		protected.PUT("/notes/:id", noteHandler.UpdateNote)
		protected.DELETE("/notes/:id", noteHandler.DeleteNote)

	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		utils.OK(c, gin.H{
			"status":  "ok",
			"service": "HYH-Blog-Gin",
		})
	})

	return r
}
