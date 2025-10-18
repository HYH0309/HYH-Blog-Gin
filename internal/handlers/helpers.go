package handlers

import (
	"strconv"

	"HYH-Blog-Gin/internal/middleware"

	"github.com/gin-gonic/gin"
)

// getUserIDFromContext 从 gin.Context 安全地提取用户 ID。
// 支持中间件可能存入的多种类型：uint/uint64/int/int64/string。
// 返回 (id, true) 表示成功，返回 (0, false) 表示未登录或类型不匹配/无效。
func getUserIDFromContext(c *gin.Context) (uint, bool) {
	v, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		return 0, false
	}
	switch id := v.(type) {
	case uint:
		if id == 0 {
			return 0, false
		}
		return id, true
	case uint64:
		if id == 0 {
			return 0, false
		}
		return uint(id), true
	case int:
		if id <= 0 {
			return 0, false
		}
		return uint(id), true
	case int64:
		if id <= 0 {
			return 0, false
		}
		return uint(id), true
	case string:
		if id == "" {
			return 0, false
		}
		parsed, err := strconv.ParseUint(id, 10, 64)
		if err != nil || parsed == 0 {
			return 0, false
		}
		return uint(parsed), true
	default:
		return 0, false
	}
}
