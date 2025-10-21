package config

// JWTConfig 定义 JWT 相关配置。
type JWTConfig struct {
	Secret string
	Expiry int // 过期时长（小时）
}

func loadJWT() JWTConfig {
	return JWTConfig{
		Secret: getEnv("JWT_SECRET", ""),
		Expiry: getEnvInt("JWT_EXPIRY", 24),
	}
}
