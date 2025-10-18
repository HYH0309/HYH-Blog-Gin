package utils

import (
	"encoding/json"
	"strconv"

	"HYH-Blog-Gin/internal/config"

	"github.com/gin-gonic/gin"
)

// GetUserIDFromContext 从 gin.Context 安全地提取用户 ID。
// 支持中间件可能存入的多种类型：uint/uint64/int/int64/string/json.Number/float64/float32。
// 返回 (id, true) 表示成功，返回 (0, false) 表示未登录或类型不匹配/无效。
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	v, ok := c.Get(config.ContextUserIDKey)
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
	case json.Number:
		if id.String() == "" {
			return 0, false
		}
		parsed, err := strconv.ParseUint(id.String(), 10, 64)
		if err != nil || parsed == 0 {
			return 0, false
		}
		return uint(parsed), true
	case float64:
		if id <= 0 {
			return 0, false
		}
		return uint(id), true
	case float32:
		if id <= 0 {
			return 0, false
		}
		return uint(id), true
	default:
		return 0, false
	}
}
