package utils

import (
	"os"

	"go.uber.org/zap"
)

var Logger *zap.Logger

// InitLogger 初始化全局 zap logger，并返回实例；调用方应在退出时调用 logger.Sync()
func InitLogger() *zap.Logger {
	if Logger != nil {
		return Logger
	}

	var logger *zap.Logger
	var err error
	if os.Getenv("ENV") == "production" {
		cfg := zap.NewProductionConfig()
		cfg.Encoding = "json"
		logger, err = cfg.Build()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		panic(err)
	}
	Logger = logger
	return Logger
}
