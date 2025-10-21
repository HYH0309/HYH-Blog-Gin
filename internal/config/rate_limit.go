package config

// RateLimitRule 表示某个动作的限流规则
// Limit: 窗口内允许的最大请求数；WindowSeconds: 时间窗长度（秒）
type RateLimitRule struct {
	Limit         int64
	WindowSeconds int
}

// RateLimitConfig 包含登录、图片上传、点赞三类动作的限流配置
// 对应环境变量（秒为单位）：
// - RL_LOGIN_LIMIT（默认 10） RL_LOGIN_WINDOW（默认 60）
// - RL_UPLOAD_LIMIT（默认 10） RL_UPLOAD_WINDOW（默认 60）
// - RL_LIKE_LIMIT（默认 30） RL_LIKE_WINDOW（默认 60）
type RateLimitConfig struct {
	Login       RateLimitRule
	UploadImage RateLimitRule
	Like        RateLimitRule
}

func loadRateLimit() RateLimitConfig {
	return RateLimitConfig{
		Login: RateLimitRule{
			Limit:         int64(getEnvInt("RL_LOGIN_LIMIT", 10)),
			WindowSeconds: getEnvInt("RL_LOGIN_WINDOW", 60),
		},
		UploadImage: RateLimitRule{
			Limit:         int64(getEnvInt("RL_UPLOAD_LIMIT", 10)),
			WindowSeconds: getEnvInt("RL_UPLOAD_WINDOW", 60),
		},
		Like: RateLimitRule{
			Limit:         int64(getEnvInt("RL_LIKE_LIMIT", 30)),
			WindowSeconds: getEnvInt("RL_LIKE_WINDOW", 60),
		},
	}
}
