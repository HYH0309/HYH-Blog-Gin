# HYH-Blog-Gin

这是一个基于 Go + Gin 的轻量级博客后端示例工程，包含 HTTP REST API、PostgreSQL 数据存储、Redis 缓存、以及可选的图片转换 gRPC 后端。

## 核心功能
- 用户注册 / 登录（JWT 鉴权）
- 笔记（notes）的增删改查，支持标签（tags）管理
- 图片上传/存储（本地静态存储 + 可选 gRPC 转换服务）
- Redis 用于计数缓存（views/likes）与加速，后台任务负责定期同步到 Postgres
- Docker 和 docker-compose 支持本地一键启动
- OpenAPI / Swagger 文档（项目中的 `api/API.md` 提供接口概要）

## 仓库结构（关键信息）
- `cmd/server` - 应用启动与生命周期管理
- `internal/config` - 配置加载（支持 `.env`）
- `internal/database` - DB/Redis 客户端构建
- `internal/handlers` - HTTP 处理器
- `internal/services` - 业务逻辑层
- `internal/repository` - 数据访问层（GORM）
- `internal/cache` - 缓存抽象
- `proto` - `imageconv.proto`：图片转换的 gRPC 定义
- `Dockerfile`, `docker-compose.yaml` - 容器化支持

## 快速开始（开发）
前提：已安装 Go 1.25+、Docker（可选）

1) 克隆并进入项目目录

Windows (cmd.exe)：

    cd D:\dev\HYH-Blog\HYH-Blog-Gin

2) 使用本地 PostgreSQL + Redis（推荐用 docker-compose）

    docker compose up -d

3) 配置环境变量
- 将必要配置写入 `.env`（或使用环境变量覆盖）。常用项：
  - `PORT`（默认 8080）
  - `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_PORT`
  - `REDIS_HOST`, `REDIS_PORT`
  - `JWT_SECRET` 等（见 `internal/config`）

4) 运行服务（开发模式）

Windows (cmd.exe)：

    go run ./cmd/server

5) 访问
- API 文档参考：`api/API.md`
- 默认监听：`http://localhost:8080`

## 构建与 Docker

构建本地二进制：

    go build -o bin/server ./cmd/server

使用 Dockerfile 构建镜像并运行：

    docker build -t hyh-blog:latest .
    docker run --rm -p 8080:8080 --env-file .env hyh-blog:latest

或使用 docker-compose（会启动 Postgres / Redis）：

    docker compose up --build

## gRPC / Protobuf

项目包含 `proto/imageconv.proto`，可用于生成 Go 的 protobuf/gRPC 代码：

    protoc --proto_path=proto --go_out=. --go-grpc_out=. proto/imageconv.proto

（需要安装 `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`）

## 测试与静态检查

- 运行单元测试：

    go test ./...

- 常用静态检查建议：
  - go vet
  - golangci-lint 或 staticcheck
  - go fmt/gofmt 或 gofumpt

## 常见问题与说明

- 初始 README 中提到的网关示例（HTTP->gRPC）为仓库的一部分，但本项目同时包含完整的博客后端。`internal/pb/imageconvpb` 在早期实现中可能为占位，建议使用 protoc 生成的代码替换占位实现以确保兼容。
- `api/API.md` 描述了主要 HTTP 接口与返回格式（统一包装响应）。

## 下一步（改进与优化建议）
下面列出按优先级分组的改进建议，并在每一项给出具体动作或示例命令，方便你逐项落地。

### 要点（高优先级）
1) 安全与密钥/秘密管理
   - 不要把敏感信息提交到仓库（`.env` 应加入 `.gitignore`）。
   - 使用强随机 `JWT_SECRET`，并考虑使用短期访问 + 刷新 token 流程。
   - 在生产环境把密钥放到安全的 secret 管理系统（例如 Vault、云平台 Secret Manager）。

2) 数据库迁移与版本管理
   - 引入迁移工具（推荐 `golang-migrate/migrate` 或 `pressly/goose`），将 schema 建表与变更以可重复脚本管理。
   - 在 `docker-compose` 或 CI 中自动运行迁移。

3) 输入验证与上传安全
   - 限制上传文件大小（在 Gin 中使用 `c.Request.Body = http.MaxBytesReader(...)` 或 `c.Request.ParseMultipartForm(maxMemory)`）。
   - 验证文件类型（magic bytes / MIME sniffing），并对文件名进行去除/重写以避免路径遍历。

4) 日志与可观测性
   - 使用结构化日志（例如 `zap` 或 `logrus`），便于采集与查询。
   - 添加请求/响应链路日志（trace id）、慢查询日志与错误上下文。
   - 暴露 /healthz、/readyz 以及 Prometheus metrics（`promhttp`）以便平台监控。

5) 超时、上下文与资源管理
   - 对外部调用（数据库、Redis、gRPC）传递 `context` 并使用合理超时/重试策略，防止请求堆积。
   - 为长时间运行任务使用 worker/队列并限制并发。

### 中等优先级
6) 测试覆盖与 CI
   - 编写单元测试（services、repository 的核心逻辑），并为 HTTP handlers 添加集成测试（httptest 或基于 docker-compose 的端到端测试）。
   - 增加 GitHub Actions（或其他 CI）来运行 `go test`、`go vet`、`golangci-lint` 和构建镜像。

7) 代码质量与静态分析
   - 配置 `golangci-lint`（或 `staticcheck`）并在 CI 中运行。
   - 确保 `go mod tidy` 保持依赖整洁、定期审查依赖升级（`dependabot`）。

8) gRPC 与 proto 工作流
   - 把生成的 pb 文件纳入构建流程：在 CI 中运行 `protoc` 以确保 proto 与代码同步。
   - 为 gRPC 客户端添加超时/断路器（例如使用 `resilience` 模式或 `go-retryablehttp` 的思路）。

### 低优先级 / 可选
9) 性能与缓存策略
   - Redis 缓存时考虑 key TTL、缓存一致性与并发竞争；`StartCounterSync` 已有良好设计，但注意错误恢复与幂等性。
   - 对热点查询应用适当索引与分页策略，避免全表扫描。

10) 镜像与部署优化
   - Docker 镜像可以使用 `distroless` 或更小的基础镜像来缩小体积。
   - 考虑多阶段构建并启用 build cache；已使用 multi-stage。

### 具体小修建议（可直接实施）
- 在 `router` 层添加一个简单的 `GET /healthz` 与 `GET /readyz`。
- 在文件上传 handler 中设置最大上传大小（例如 10 MiB）并校验 MIME。
- 在 `cmd/server/app.go` 的 `WaitForShutdown` 中，将 shutdown timeout 参数化为配置项（而不是硬编码 5s）。
- 为 `StartCounterSync` 日志添加计数成功/失败的统计信息，方便定位异常。
- 在 `internal/utils/user_context.go` 的 `GetUserIDFromContext` 中增加对 `json.Number` 或 `float64` 的兼容（如果某些中间件/框架把 id 存为这些类型）。

### 建议的 GitHub Action（最小）：
- 在 `.github/workflows/ci.yml` 中运行：
  - checkout
  - setup-go
  - go mod download
  - go test ./...
  - golangci-lint run
  - build

## 迁移与数据（建议）
- 使用 `migrate` 添加基础表迁移并在 `docker-compose` 的 app 服务启动前执行迁移。
- 提供一个 `seed` 脚本来创建测试账户与示例数据，方便本地开发。

## 安全审计/进一步工作
- 对用户密码存储确认使用 `bcrypt` 或 `argon2`（检查 `internal/services/user_service.go` 中的实现）。
- 添加速率限制（例如 `github.com/ulule/limiter` 或基于 Nginx/Ingress 层的速率策略），防止暴力破解登录或滥用上传。

## 结束语

我已经把项目的 README 替换为上面的更全面文档（包含快速启动与改进建议）。

接下来我将运行代码质量/错误检查（`go vet`/编译错误检查）以确保编辑不会影响到代码状态，并给出一份需求覆盖的简短汇总。
