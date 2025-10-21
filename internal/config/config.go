// Package config go
// Package config 提供应用配置的加载与访问。
// 配置值主要来自环境变量；若项目根目录存在 .env 文件，将在进程启动时尝试加载以填充环境变量。
package config

// Config 汇总服务、数据库、Redis 与 JWT 的配置项。
type Config struct {
	// Server 包含 HTTP 服务相关配置。
	Server ServerConfig

	// Database 包含数据库连接相关配置（PostgreSQL）。
	Database PostgresConfig

	// Redis 包含 Redis 连接相关配置。
	Redis RedisConfig

	// JWT 包含 JWT 令牌相关配置。
	JWT JWTConfig

	// RateLimit 包含各接口的限流配置。
	RateLimit RateLimitConfig
}

// Load 尝试从项目根目录的 .env 文件加载环境变量（可选），
// 然后基于环境变量构建并返回 *Config。
// 说明：忽略 .env 加载错误；若环境变量缺失将采用安全的默认值；函数始终返回非 nil 的 *Config。
func Load() *Config {
	// 加载 .env 文件（可选）
	LoadDotEnv()

	cfg := &Config{}

	// Server 配置
	cfg.Server = loadServer()

	// Database 配置（PostgreSQL）
	cfg.Database = loadPostgres()

	// Redis 配置
	cfg.Redis = loadRedis()

	// JWT 配置
	cfg.JWT = loadJWT()

	// RateLimit 配置
	cfg.RateLimit = loadRateLimit()

	return cfg
}
