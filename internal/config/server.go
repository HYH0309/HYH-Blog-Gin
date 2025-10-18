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
}

func loadServer() ServerConfig {
	return ServerConfig{
		Port:            getEnv("PORT", "8080"),
		UploadDir:       getEnv("UPLOAD_DIR", "./uploads"),
		UploadURLPrefix: getEnv("UPLOAD_URL_PREFIX", "/uploads"),
		GRPCAddr:        getEnv("GRPC_ADDR", "127.0.0.1:50051"),
		ImageGRPCAddr:   getEnv("IMAGE_GRPC_ADDR", ""),
	}
}
