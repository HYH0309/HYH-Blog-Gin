package utils

import "strconv"

// ParseUintParam  将字符串解析为 uint，解析失败返回 false。
// 放在单独的文件以便在多个 handler 中复用，避免重复定义。
func ParseUintParam(s string) (uint, bool) {
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(v), true
}
