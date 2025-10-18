package router

import (
	"github.com/gin-gonic/gin"

	"HYH-Blog-Gin/internal/auth"
	"HYH-Blog-Gin/internal/handlers"
	"HYH-Blog-Gin/internal/middleware"
)

// registerProtectedRoutes 注册需要鉴权的路由，统一在 /api/v1 前缀下。
func registerProtectedRoutes(r *gin.Engine, jwt *auth.JWTService, userHandler *handlers.UserHandler, noteHandler *handlers.NoteHandler, imageHandler *handlers.ImageHandler, tagHandler *handlers.TagHandler) {
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
		// 点赞接口
		v1.POST("/notes/:id/like", noteHandler.LikeNote)

		// 图片管理：上传、列表、info、删除（统一为 /images）
		if imageHandler != nil {
			v1.POST("/images", imageHandler.Upload)
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
