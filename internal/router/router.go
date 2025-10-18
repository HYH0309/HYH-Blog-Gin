package router

import (
	"HYH-Blog-Gin/internal/auth"
	"HYH-Blog-Gin/internal/handlers"
	"HYH-Blog-Gin/internal/middleware"
	"HYH-Blog-Gin/internal/utils"

	"github.com/gin-gonic/gin"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter 构建并返回 Gin 引擎，集中注册中间件与路由。
// 将路由拆分到 registerPublicRoutes/registerProtectedRoutes 中，统一 API 前缀为 /api/v1。
func SetupRouter(jwt *auth.JWTService, userHandler *handlers.UserHandler, noteHandler *handlers.NoteHandler, imageHandler *handlers.ImageHandler, tagHandler *handlers.TagHandler) *gin.Engine {
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORSMiddleware())

	// 静态文件路由（图片保存目录）
	r.Static("/static/images", "./static/images")

	// Add root-level health endpoints: /healthz and /readyz
	r.GET("/healthz", func(c *gin.Context) {
		utils.OK(c, gin.H{"status": "ok", "service": "HYH-Blog-Gin"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		// Basic readiness: the app is running. For stronger readiness checks,
		// consider injecting DB/Redis clients and performing Ping checks.
		utils.OK(c, gin.H{"status": "ready", "service": "HYH-Blog-Gin"})
	})

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 健康检查（放到 /api/v1/health）
	v1 := r.Group("/api/v1")
	v1.GET("/health", func(c *gin.Context) {
		utils.OK(c, gin.H{
			"status":  "ok",
			"service": "HYH-Blog-Gin",
		})
	})

	// swagger 文档路由（统一到 /api/v1/swagger.json）
	r.StaticFile("/api/v1/swagger.json", "./api/swagger.json")
	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/api/v1/swagger.json")))

	// register routes
	registerPublicRoutes(r, userHandler)
	registerProtectedRoutes(r, jwt, userHandler, noteHandler, imageHandler, tagHandler)

	return r
}
