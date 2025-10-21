// cmd/server/app.go
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"HYH-Blog-Gin/internal/auth"
	"HYH-Blog-Gin/internal/cache"
	"HYH-Blog-Gin/internal/config"
	"HYH-Blog-Gin/internal/database"
	"HYH-Blog-Gin/internal/grpcclient"
	"HYH-Blog-Gin/internal/handlers"
	"HYH-Blog-Gin/internal/repository"
	"HYH-Blog-Gin/internal/router"
	"HYH-Blog-Gin/internal/services"
	"HYH-Blog-Gin/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Application 封装应用的核心组件
type Application struct {
	Config       *config.Config
	Database     *database.DB
	Redis        *redis.Client
	Cache        cache.Cache
	JWTService   *auth.JWTService
	Services     *ServiceContainer
	Handlers     *HandlerContainer
	Router       *gin.Engine
	Server       *http.Server
	CleanupFuncs []func()
}

// ServiceContainer 服务容器（使用接口类型）
type ServiceContainer struct {
	UserService  services.UserService
	NoteService  services.NoteService
	TagService   services.TagService
	ImageService services.ImageService
}

// HandlerContainer 处理器容器
type HandlerContainer struct {
	UserHandler  *handlers.UserHandler
	NoteHandler  *handlers.NoteHandler
	TagHandler   *handlers.TagHandler
	ImageHandler *handlers.ImageHandler
}

// InitializeApplication 初始化应用的所有组件
func InitializeApplication() *Application {
	app := &Application{CleanupFuncs: make([]func(), 0)}

	// Init structured logger
	logger := utils.InitLogger()
	app.registerCleanup(func() {
		_ = logger.Sync()
	})

	app.loadConfiguration()
	app.initializeDataStores()
	app.initializeServices()
	app.initializeHandlers()
	app.initializeRouterAndServer()
	app.startBackgroundTasks()

	return app
}

// loadConfiguration 加载应用配置
func (app *Application) loadConfiguration() {
	app.Config = config.Load()
	// 强制要求 JWT_SECRET 在任何环境下必须显式设置
	if app.Config.JWT.Secret == "" {
		log.Fatal("JWT_SECRET is not set; set it in environment or .env before starting the server")
	}
	log.Println("配置加载完成")
}

// initializeDataStores 初始化数据库与缓存
func (app *Application) initializeDataStores() {
	// 初始化主数据库
	db, err := database.NewDB(app.Config)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	app.Database = db

	// 注册数据库关闭清理函数
	app.registerCleanup(func() {
		if err := app.Database.Close(); err != nil {
			log.Printf("关闭数据库失败: %v", err)
		}
	})

	// 初始化 Redis/Cache
	redisClient, err := database.NewRedisClient(app.Config)
	if err != nil {
		log.Printf("连接 Redis 失败，使用 Noop Cache: %v", err)
		app.Cache = cache.NewNoOpCache()
		app.Redis = nil
		return
	}
	app.Redis = redisClient
	app.Cache = cache.NewRedisCache(redisClient)
	log.Println("Redis 已连接")

	// 注册 Redis 关闭
	app.registerCleanup(func() {
		if app.Redis != nil {
			if err := app.Redis.Close(); err != nil {
				log.Printf("关闭 Redis 失败: %v", err)
			}
		}
	})
}

// initializeServices 初始化服务层
func (app *Application) initializeServices() {
	app.JWTService = auth.NewJWTService(app.Config)

	// 仓储
	userRepo := repository.NewUserRepository(app.Database.DB)
	noteRepoBase := repository.NewNoteRepository(app.Database.DB)
	// 包装缓存仓储
	noteRepo := repository.NewCachedNoteRepository(noteRepoBase, app.Cache, 5*time.Minute)
	tagRepo := repository.NewTagRepository(app.Database.DB)
	imageRepo := repository.NewImageRepository(app.Database.DB)

	app.Services = &ServiceContainer{
		UserService:  services.NewUserService(userRepo),
		NoteService:  services.NewNoteService(noteRepo),
		TagService:   services.NewTagService(tagRepo),
		ImageService: services.NewImageService(nil, nil, 80, imageRepo), // will be replaced below
	}

	// 初始化 image service (may use grpc client)
	ctx := context.Background()
	addr := app.Config.Server.ImageGRPCAddr
	if addr == "" {
		addr = app.Config.Server.GRPCAddr
	}
	grpcClient, grpcCleanup, err := grpcclient.NewImageClient(ctx, addr)
	if err != nil {
		log.Printf("图片 gRPC 客户端创建失败，将使用本地转换: %v", err)
		grpcClient = nil
		grpcCleanup = func() {}
	}
	// register grpc cleanup if any
	app.registerCleanup(grpcCleanup)

	// 本地存储目录/URL 统一使用配置
	uploadDir := app.Config.Server.UploadDir
	uploadURL := app.Config.Server.UploadURLPrefix
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		log.Fatalf("创建上传目录失败: %v", err)
	}
	storage := services.NewLocalStorage(uploadDir, uploadURL)
	app.Services.ImageService = services.NewImageService(grpcClient, storage, 80, imageRepo)
}

// initializeHandlers 初始化HTTP处理器
func (app *Application) initializeHandlers() {
	app.Handlers = &HandlerContainer{
		UserHandler:  handlers.NewUserHandler(app.Services.UserService, app.JWTService),
		NoteHandler:  handlers.NewNoteHandler(app.Services.NoteService, app.Cache),
		TagHandler:   handlers.NewTagHandler(app.Services.TagService),
		ImageHandler: handlers.NewImageHandler(app.Services.ImageService),
	}
}

// initializeRouterAndServer 初始化路由和HTTP服务器
func (app *Application) initializeRouterAndServer() {
	app.Router = router.SetupRouter(app.Config, app.JWTService, app.Handlers.UserHandler, app.Handlers.NoteHandler, app.Handlers.ImageHandler, app.Handlers.TagHandler, app.Database, app.Redis)
	app.Server = &http.Server{Addr: ":" + app.Config.Server.Port, Handler: app.Router}
}

// startBackgroundTasks 启动后台任务（如计数器同步）
func (app *Application) startBackgroundTasks() {
	if app.Redis != nil {
		ctx, cancel := context.WithCancel(context.Background())
		app.registerCleanup(cancel)
		go StartCounterSync(ctx, app.Database.DB, app.Cache, 10*time.Second)
		log.Println("计数器同步任务已启动")
	}
}

// StartServer 启动HTTP服务器并在后台运行
func (app *Application) StartServer() {
	go func() {
		log.Printf("服务器启动在端口 %s", app.Config.Server.Port)
		if err := app.Server.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
			log.Fatalf("启动服务器失败: %v", err)
		}
	}()
}

// WaitForShutdown 阻塞直到收到终止信号，并优雅关闭
func (app *Application) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("收到关闭信号，正在优雅停止...")

	// 执行优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.Server.Shutdown(ctx); err != nil {
		log.Fatalf("服务器强制关闭: %v", err)
	}
}

// Cleanup 执行注册的清理函数（逆序）
func (app *Application) Cleanup() {
	log.Println("执行清理函数...")
	for i := len(app.CleanupFuncs) - 1; i >= 0; i-- {
		app.CleanupFuncs[i]()
	}
	log.Println("清理完成")
}

// registerCleanup 注册清理函数
func (app *Application) registerCleanup(f func()) {
	app.CleanupFuncs = append(app.CleanupFuncs, f)
}
