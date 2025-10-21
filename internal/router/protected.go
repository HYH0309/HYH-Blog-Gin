package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"HYH-Blog-Gin/internal/auth"
	"HYH-Blog-Gin/internal/config"
	"HYH-Blog-Gin/internal/handlers"
	"HYH-Blog-Gin/internal/middleware"
)

// registerProtectedRoutes 注册需要鉴权的路由，统一在 /api/v1 前缀下。
func registerProtectedRoutes(r *gin.Engine, cfg *config.Config, jwt *auth.JWTService, userHandler *handlers.UserHandler, noteHandler *handlers.NoteHandler, imageHandler *handlers.ImageHandler, tagHandler *handlers.TagHandler, rdb *redis.Client) {
	v1 := r.Group("/api/v1")
	// 使用鉴权中间件
	v1.Use(middleware.AuthMiddleware(jwt))
	{
		// 用户相关
		v1.GET("/user/profile", userHandler.GetProfile)

		// 笔记相关
		v1.GET("/notes", noteHandler.GetNotes)
		v1.POST("/notes", noteHandler.CreateNote)
		v1.GET("/notes/:id", noteHandler.GetNote)
		v1.PUT("/notes/:id", noteHandler.UpdateNote)
		v1.DELETE("/notes/:id", noteHandler.DeleteNote)
		// 点赞接口 - 使用配置限流
		likeRule := cfg.RateLimit.Like
		likeWindow := time.Duration(likeRule.WindowSeconds) * time.Second
		likeLimiter := middleware.RateLimitUser(rdb, "like", likeRule.Limit, likeWindow)
		v1.POST("/notes/:id/like", likeLimiter, noteHandler.LikeNote)

		// 图片管理：上传、列表、info、删除（统一为 /images）
		if imageHandler != nil {
			// 上传限流 - 使用配置
			uploadRule := cfg.RateLimit.UploadImage
			uploadWindow := time.Duration(uploadRule.WindowSeconds) * time.Second
			uploadLimiter := middleware.RateLimitUser(rdb, "upload_image", uploadRule.Limit, uploadWindow)
			v1.POST("/images", uploadLimiter, imageHandler.Upload)
			v1.GET("/images", imageHandler.List)
			v1.GET("/images/info", imageHandler.Info)
			v1.DELETE("/images", imageHandler.Delete) // keep query param ?url=...
		}

		// 标签管理：CRUD
		if tagHandler != nil {
			v1.GET("/tags", tagHandler.List)
			v1.POST("/tags", tagHandler.Create)
			v1.GET("/tags/:id", tagHandler.Get)
			v1.PUT("/tags/:id", tagHandler.Update)
			v1.DELETE("/tags/:id", tagHandler.Delete)
		}
	}
}
