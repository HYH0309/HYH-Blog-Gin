package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware 允许跨域访问：
// - 当 allowedOrigins 为空时，允许任意来源但不携带凭证（Allow-Origin:* 且不设置 Allow-Credentials）
// - 当 allowedOrigins 非空时，仅回显在白名单内的 Origin，并可按 allowCredentials 设置是否允许凭证
func CORSMiddleware(allowedOrigins []string, allowCredentials bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowedOrigin := ""
		if len(allowedOrigins) == 0 {
			// 宽松模式（开发）：任意来源，但不能携带凭证
			allowedOrigin = "*"
		} else if origin != "" {
			// 精确匹配白名单
			for _, o := range allowedOrigins {
				if strings.EqualFold(o, origin) {
					allowedOrigin = origin
					break
				}
			}
		}

		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
		}
		c.Header("Vary", "Origin")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With, X-Request-Id")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-Id")

		// 仅当回显具体 Origin 且允许时才允许凭证
		if allowedOrigin != "*" && allowedOrigin != "" && allowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
