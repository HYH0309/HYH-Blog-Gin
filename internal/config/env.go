package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// LoadDotEnv 尝试从项目根目录加载 .env 文件（若存在）。
func LoadDotEnv() {
	_ = godotenv.Load()
}

// 获取环境变量，若不存在则返回默认值
func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// 获取环境变量并转换为整数，若不存在或转换失败则返回默认值
func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// 获取环境变量并转换为布尔值，支持 true/false、1/0、t/f、yes/no 等；若不存在或转换失败则返回默认值
func getEnvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}
