package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ZapLogger 记录每个请求的基本信息：方法、路径、状态码、耗时、客户端 IP、请求 ID。
func ZapLogger(logger *zap.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = zap.NewNop()
	}
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		rid, _ := c.Get("requestID")
		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", latency),
		}
		if ridStr, ok := rid.(string); ok && ridStr != "" {
			fields = append(fields, zap.String("request_id", ridStr))
		}

		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Error("request error", append(fields, zap.String("error", e.Error()))...)
			}
		} else {
			logger.Info("request", fields...)
		}
	}
}
