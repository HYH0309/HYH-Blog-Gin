HYH-Blog-Gin
=====================

简介
----
HYH-Blog-Gin 是一个基于 Gin + GORM 的轻量级笔记/博客后端示例，包含用户、笔记、标签与图片管理。项目采用分层架构（handlers/services/repository），并使用 Redis 做缓存、PostgreSQL 做持久化、gRPC 调用远端图片转换服务。

快速开始（Windows cmd）
-----------------------
1. 克隆仓库并进入目录：

```cmd
cd /d D:\dev\HYH-Blog\HYH-Blog-Gin
```

2. 配置环境变量（参考 `.env.example`）并在项目根放置 `.env`（可选）：

```cmd
copy .env.example .env
:: 编辑 .env 文件，填入真实连接信息
```

3. 下载依赖并构建：

```cmd
go mod tidy
go build ./...
```

4. 运行（默认监听端口 8080）：

```cmd
go run ./cmd/server
```

重要环境变量（见 `.env.example`）
----------------------------------
- PORT: HTTP 服务端口，默认 8080
- UPLOAD_DIR: 本地静态文件存储目录，默认 ./uploads
- UPLOAD_URL_PREFIX: 静态文件的 URL 前缀，默认 /uploads
- GRPC_ADDR / IMAGE_GRPC_ADDR: gRPC 服务地址（图片转换）
- DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME/DB_SSLMODE: PostgreSQL 连接配置
- REDIS_HOST/REDIS_PORT/REDIS_PASSWORD/REDIS_DB: Redis 连接配置
- JWT_SECRET / JWT_EXPIRY: JWT 秘钥与过期时长（小时）

数据库迁移
-----------
本仓库包含 `migrations/` 文件夹（SQL）。使用你喜欢的迁移工具（例如 `migrate` CLI）运行迁移：

示例（假设使用 golang-migrate）：

```cmd
migrate -path migrations -database "postgres://USER:PASSWORD@HOST:PORT/DBNAME?sslmode=disable" up
```

Swagger
-------
仓库内包含 `api/swagger.yaml` 与生成结果。项目使用 swag 注释生成接口文档（如果需要重新生成）：

```cmd
swag init -g ./cmd/server/app.go -o ./api
```

图片上传与响应格式
------------------
API 层使用 `internal/utils/response.go` 提供统一响应结构（{code, message, data, meta}）。图片上传接口会返回 JSON 格式的 URL 字段（例如 `/static/images/xxx.webp`），静态文件由本地 `UPLOAD_DIR` 提供访问。

测试与 CI
--------
建议在 CI 中运行以下命令：

```cmd
go mod tidy
go vet ./...
go test ./...
go build ./...
```

我已在仓库添加了一个基础 GitHub Actions workflow（.github/workflows/ci.yml）用于构建与测试。

优化建议速览
-------------
- 添加完整的单元/集成测试（cover 关键路径：image, note, auth）。
- 将图片存储切换到对象存储（S3 等）并加 CDN。图片转换建议异步化并加入重试策略。
- 引入请求 ID 中间件并将其打印到日志中以便追踪。
- 增加 golangci-lint 与 go vet 的 CI 门禁，保持代码质量。
- 改善配置管理：在生产使用 secret 管理（不要直接在 `.env` 中存放敏感信息）。

我可以继续：生成示例 `.env`（已添加）、为关键 handler 添加单元测试、或把 CI workflow 扩展为包含 lint/coverage。你想先做哪项？

