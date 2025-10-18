package middleware

import (
	"strings"

	"HYH-Blog-Gin/internal/auth"
	"HYH-Blog-Gin/internal/config"
	"HYH-Blog-Gin/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 验证 Bearer Token，有效则将 userID 写入上下文
func AuthMiddleware(jwt *auth.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authz := c.GetHeader("Authorization")
		parts := strings.SplitN(authz, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			utils.Unauthorized(c, "missing or invalid Authorization header")
			c.Abort()
			return
		}
		userID, err := jwt.ParseToken(parts[1])
		if err != nil || userID == 0 {
			utils.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}
		c.Set(config.ContextUserIDKey, userID)
		c.Next()
	}
}
