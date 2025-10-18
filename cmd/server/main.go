// cmd/server/main.go
package main

import (
	"HYH-Blog-Gin/internal/auth"
	"HYH-Blog-Gin/internal/config"
	"HYH-Blog-Gin/internal/database"
	"HYH-Blog-Gin/internal/handlers"
	"HYH-Blog-Gin/internal/repository"
	"HYH-Blog-Gin/internal/router"
	"HYH-Blog-Gin/internal/services"
	"log"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库 (PostgresSQL)
	db, err := database.NewDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("error closing database: %v", err)
		}
	}()

	// 初始化Redis
	redisClient, err := database.NewRedisClient(cfg)
	if err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
	} else {
		defer func() {
			if err := redisClient.Close(); err != nil {
				log.Printf("error closing redis: %v", err)
			}
		}()
		log.Println("Redis connection established")
	}

	// 初始化JWT服务
	jwtService := auth.NewJWTService(cfg)

	// 初始化 Repository
	userRepo := repository.NewUserRepository(db.DB)
	noteRepo := repository.NewNoteRepository(db.DB)

	// 初始化 Services
	userSvc := services.NewUserService(userRepo)
	noteSvc := services.NewNoteService(noteRepo)

	// 初始化处理器（使用 service 层）
	userHandler := handlers.NewUserHandler(userSvc, jwtService)
	noteHandler := handlers.NewNoteHandler(noteSvc)

	// 路由 与 中间件注册由独立模块负责
	r := router.SetupRouter(jwtService, userHandler, noteHandler)

	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
