package config

// ServerConfig 定义 HTTP 服务相关配置。
type ServerConfig struct {
	Port string // 监听端口，如 "8080"

	// UploadDir 本地文件存储目录（相对或绝对路径）。
	UploadDir string

	// UploadURLPrefix 对外访问静态文件的 URL 前缀，例如 "/static" 或 "/uploads"。
	UploadURLPrefix string

	// GRPCAddr 全局 gRPC 服务地址（可作为默认）
	GRPCAddr string

	// ImageGRPCAddr 专用图片服务的 gRPC 地址（优先于 GRPCAddr）
	ImageGRPCAddr string

	// 允许的跨域来源（逗号分隔）。为空则表示允许任意来源但不携带凭证。
	AllowedOrigins []string

	// 是否允许跨域携带凭证（仅当 AllowedOrigins 非空时生效）。
	AllowCredentials bool
}

func loadServer() ServerConfig {
	// 解析允许的来源列表（逗号分隔，去空格，忽略空条目）
	originsCSV := getEnv("ALLOWED_ORIGINS", "")
	var origins []string
	if originsCSV != "" {
		// 简单分割并修剪
		raw := originsCSV
		start := 0
		for i := 0; i <= len(raw); i++ {
			if i == len(raw) || raw[i] == ',' {
				item := raw[start:i]
				// 修剪空白
				for len(item) > 0 && (item[0] == ' ' || item[0] == '\t') {
					item = item[1:]
				}
				for len(item) > 0 && (item[len(item)-1] == ' ' || item[len(item)-1] == '\t') {
					item = item[:len(item)-1]
				}
				if item != "" {
					origins = append(origins, item)
				}
				start = i + 1
			}
		}
	}

	return ServerConfig{
		Port:             getEnv("PORT", "8080"),
		UploadDir:        getEnv("UPLOAD_DIR", "./uploads"),
		UploadURLPrefix:  getEnv("UPLOAD_URL_PREFIX", "/uploads"),
		GRPCAddr:         getEnv("GRPC_ADDR", "127.0.0.1:50051"),
		ImageGRPCAddr:    getEnv("IMAGE_GRPC_ADDR", ""),
		AllowCredentials: getEnvBool("ALLOW_CREDENTIALS", true),
		AllowedOrigins:   origins,
	}
}
