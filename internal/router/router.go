package router

import (
	"context"
	"net/http"
	"time"

	"HYH-Blog-Gin/internal/auth"
	"HYH-Blog-Gin/internal/config"
	"HYH-Blog-Gin/internal/database"
	"HYH-Blog-Gin/internal/handlers"
	"HYH-Blog-Gin/internal/middleware"
	"HYH-Blog-Gin/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// SetupRouter 构建并返回 Gin 引擎，集中注册中间件与路由。
// 统一 API 前缀为 /api/v1；静态资源与 CORS/日志中间件按配置注入。
func SetupRouter(cfg *config.Config, jwt *auth.JWTService, userHandler *handlers.UserHandler, noteHandler *handlers.NoteHandler, imageHandler *handlers.ImageHandler, tagHandler *handlers.TagHandler, db *database.DB, rdb *redis.Client) *gin.Engine {
	r := gin.New()

	// 中间件：请求ID、日志、恢复、CORS
	r.Use(middleware.RequestID())
	r.Use(middleware.ZapLogger(utils.Logger))
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware(cfg.Server.AllowedOrigins, cfg.Server.AllowCredentials))

	// 静态文件路由（来自配置）
	if cfg.Server.UploadURLPrefix != "" && cfg.Server.UploadDir != "" {
		r.Static(cfg.Server.UploadURLPrefix, cfg.Server.UploadDir)
	}

	// Add root-level health endpoints: /healthz and /readyz
	r.GET("/healthz", func(c *gin.Context) {
		utils.OK(c, gin.H{"status": "ok", "service": "HYH-Blog-Gin"})
	})

	r.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 1200*time.Millisecond)
		defer cancel()

		// 聚合检查 DB、Redis、gRPC
		type checkResult struct {
			name string
			ok   bool
			err  string
		}
		resCh := make(chan checkResult, 3)

		// DB ping
		go func() {
			ok := false
			errStr := ""
			if db != nil {
				if sqlDB, err := db.DB.DB(); err == nil {
					ctx2, cancel2 := context.WithTimeout(ctx, 600*time.Millisecond)
					defer cancel2()
					if err := sqlDB.PingContext(ctx2); err == nil {
						ok = true
					} else {
						errStr = err.Error()
					}
				} else {
					errStr = err.Error()
				}
			}
			resCh <- checkResult{"db", ok, errStr}
		}()

		// Redis ping
		go func() {
			ok := false
			errStr := ""
			if rdb != nil {
				ctx2, cancel2 := context.WithTimeout(ctx, 400*time.Millisecond)
				defer cancel2()
				if err := rdb.Ping(ctx2).Err(); err == nil {
					ok = true
				} else {
					errStr = err.Error()
				}
			}
			resCh <- checkResult{"redis", ok, errStr}
		}()

		// gRPC dial (image service)
		go func() {
			ok := false
			errStr := ""
			addr := cfg.Server.ImageGRPCAddr
			if addr == "" {
				addr = cfg.Server.GRPCAddr
			}
			if addr != "" {
				conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err == nil {
					defer func(conn *grpc.ClientConn) {
						err := conn.Close()
						if err != nil {
							utils.Logger.Error("failed to close grpc connection")
						}
					}(conn)
					ctx2, cancel2 := context.WithTimeout(ctx, 400*time.Millisecond)
					defer cancel2()
					state := conn.GetState()
					for state != connectivity.Ready {
						if !conn.WaitForStateChange(ctx2, state) {
							break
						}
						state = conn.GetState()
					}
					if state == connectivity.Ready {
						ok = true
					} else {
						errStr = "grpc not ready"
					}
				} else {
					errStr = err.Error()
				}
			}
			resCh <- checkResult{"grpc", ok, errStr}
		}()

		results := make(map[string]any)
		ready := true
		for i := 0; i < 3; i++ {
			select {
			case r := <-resCh:
				results[r.name] = gin.H{"ok": r.ok, "err": r.err}
				if !r.ok {
					ready = false
				}
			case <-ctx.Done():
				ready = false
				i = 2 // break outer
			}
		}
		code := http.StatusOK
		status := "ready"
		if !ready {
			code = http.StatusServiceUnavailable
			status = "not-ready"
		}
		utils.WithStatus(code, c, gin.H{"status": status, "checks": results})
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
	registerPublicRoutes(r, cfg, userHandler, rdb)
	registerProtectedRoutes(r, cfg, jwt, userHandler, noteHandler, imageHandler, tagHandler, rdb)

	return r
}
