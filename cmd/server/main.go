// @title HYH-Blog-Gin API
// @version 1.0
// @description HYH-Blog-Gin 是一个基于 Gin 的轻量级笔记/博客后端示例
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// cmd/server/main.go
package main

func main() {
	app := InitializeApplication()
	defer app.Cleanup()

	app.StartServer()
	app.WaitForShutdown()
}
