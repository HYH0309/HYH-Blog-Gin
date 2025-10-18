package config

// RedisConfig 定义 Redis 连接配置。
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func loadRedis() RedisConfig {
	return RedisConfig{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnv("REDIS_PORT", "6379"),
		Password: getEnv("REDIS_PASSWORD", "123456"),
		DB:       getEnvInt("REDIS_DB", 0),
	}
}
