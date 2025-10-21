package middleware

import (
	"time"

	"HYH-Blog-Gin/internal/utils"

	"github.com/gin-gonic/gin"
)

const requestIDHeader = "X-Request-Id"

// RequestID 从请求头读取 X-Request-Id；若不存在则生成一个随机 ID，注入到上下文与响应头。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(requestIDHeader)
		if rid == "" {
			if v, err := utils.RandHex(12); err == nil {
				rid = v
			} else {
				rid = time.Now().UTC().Format("20060102150405.000000000")
			}
		}
		c.Set("requestID", rid)
		c.Writer.Header().Set(requestIDHeader, rid)
		c.Next()
	}
}
