package utils

import (
	"path/filepath"
	"regexp"
	"strings"
)

// SanitizeFileName 清理文件名，移除路径分隔、空白与不安全字符，保留扩展名。
// 结果只包含字母、数字、下划线、连字符和点（用于扩展名），并截断到合理长度。
func SanitizeFileName(name string) string {
	if name == "" {
		return "file"
	}
	// 仅保留文件名部分
	name = filepath.Base(name)
	// 将空格替换为下划线
	name = strings.ReplaceAll(name, " ", "_")
	// 允许的字符集：字母数字 - _ .
	re := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	name = re.ReplaceAllString(name, "")
	// 防止过长
	if len(name) > 255 {
		ext := filepath.Ext(name)
		base := name[:255-len(ext)]
		name = base + ext
	}
	if name == "" {
		return "file"
	}
	return name
}

// IsAllowedImageContentType 根据 DetectContentType 的返回判断是否为允许的图片类型
func IsAllowedImageContentType(contentType string) bool {
	ct := strings.ToLower(contentType)
	switch ct {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
		return true
	default:
		// 有些浏览器/工具会返回带有 ;charset 的 content-type
		if strings.HasPrefix(ct, "image/") {
			return true
		}
		return false
	}
}
