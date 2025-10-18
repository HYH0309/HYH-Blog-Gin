# HYH-Blog-Gin

一个基于 Go + Gin 的轻量级博客/笔记后端脚手架，内置用户注册/登录（JWT 鉴权）、笔记 CRUD、标签多对多关联、分页等基础能力，开箱可用于个人博客、知识卡片、随笔记录等场景。

## 特色与作用
- 面向个人/团队的“笔记/文章”数据服务
- 标准分层：Handlers（HTTP）/ Repository（数据访问）/ Models（领域模型）
- 鉴权内置：注册、登录、JWT 颁发与校验
- 数据存储：PostgreSQL（GORM 自动迁移），Redis 预留（可做会话/缓存）
- 生产可扩展：中间件、错误处理、日志与配置结构化

## 统一响应格式（Unified API Response）
所有 HTTP 接口统一返回以下结构：

```
{
  "code": number,          // 业务码：0 表示成功，非 0 表示失败（与 HTTP 状态码一致或自定义）
  "message": string,       // 说明信息：如 "success"、"created"、具体错误描述
  "data": any,             // 业务数据（可选）
  "meta": {                // 分页信息（可选，仅列表接口返回）
    "page": number,
    "limit": number,
    "total": number
  }
}
```

- 成功：`code = 0`，`message` 通常为 `success/created`，携带 `data`。
- 失败：`code != 0`，与 HTTP 状态码对齐（如 400/401/403/404/500 等），`message` 为错误描述。

开发规范：
- 统一响应类型定义在 `internal/models/response.go`
- 统一返回工具在 `internal/utils/response.go`
- Handler 通过 `utils.OK/Created/BadRequest/Unauthorized/...` 返回响应

## 技术栈
- 语言/框架：Go 1.21，Gin
- ORM：GORM（PostgreSQL 驱动）
- 鉴权：JWT（HS256）
- 缓存/会话：Redis（可选）
- 配置：dotenv（.env）
- 密码：bcrypt（golang.org/x/crypto）

## 目录结构
```
.
├── cmd/
│   └── server/
│       └── main.go          # 程序入口，路由与依赖注入
├── internal/
│   ├── auth/                # JWT 服务
│   ├── config/              # 配置加载
│   ├── database/            # 数据库与 Redis 初始化
│   ├── handlers/            # HTTP 处理器（用户、笔记）
│   ├── middleware/          # 中间件（CORS、鉴权）
│   ├── models/              # 领域模型（Base/User/Note/Tag）
│   └── repository/          # 仓储实现（User/Note/Tag）
├── api/
│   └── docs/                # API 文档（Postman 集合）
├── migrations/              # 迁移（预留，当前由 GORM 自动迁移）
├── Dockerfile               # 占位
├── docker-compose.yaml      # 占位
├── .env.example             # 环境变量示例
├── .env                     # 你的本地环境变量（已生成，可按需修改）
└── go.mod
```

## 配置
通过 `.env` 或系统环境变量配置。关键项：
- 服务器
  - PORT（默认 8080）
- PostgreSQL
  - DB_HOST（默认 localhost）
  - DB_PORT（默认 5432）
  - DB_USER（默认 postgres）
  - DB_PASSWORD（默认 postgres）
  - DB_NAME（默认 postgres；示例 .env 中为 blog，可改回 postgres）
  - DB_SSL_MODE（默认 disable）
- Redis（可选）
  - REDIS_HOST、REDIS_PORT、REDIS_PASSWORD、REDIS_DB
- JWT
  - JWT_SECRET（请在生产使用强随机值）
  - JWT_EXPIRY（小时，默认 24）

仓库已提供 `.env.example`，你也可以直接使用根目录 `.env`（已生成）。

## 快速开始（Windows cmd）
确保本机已安装 Go 1.21+、PostgreSQL 与（可选）Redis，并保证数据库凭据与 `.env` 匹配。

```cmd
cd /d D:\dev\HYH-Blog\HYH-Blog-Gin

:: 安装依赖（可选，首次运行 go run 会自动拉取）
go mod tidy

:: 启动服务
go run .\cmd\server
```

打开浏览器或终端访问：
- 健康检查: http://localhost:8080/health

## API 概览
公共接口：
- POST /api/register
  - body: { "email": string, "username": string, "password": string }
  - 201: data = { id, email, username }
- POST /api/login
  - body: { "identifier": string, "password": string } // identifier 支持邮箱或用户名
  - 200: data = { token }

鉴权接口（需要请求头 Authorization: Bearer <token>）：
- GET /api/user/profile
- GET /api/notes?page=1&limit=10
- POST /api/notes
  - body: { "title": string, "content": string, "tags"?: string[], "public"?: bool }
- GET /api/notes/:id
- PUT /api/notes/:id
  - body: { "title"?: string, "content"?: string, "tags"?: string[], "public"?: bool }
- DELETE /api/notes/:id

说明：
- 标签为多对多（notes ⇄ tags），POST/PUT 支持覆盖式更新标签集合。
- 当前示例未暴露“搜索接口”，但仓储层已具备相关能力，可按需补充 Handler 与路由。

## 运行时行为
- 启动时自动连接 PostgreSQL 和（可选）Redis，并对 User/Note/Tag 执行 GORM 自动迁移。
- 注册时对密码进行 bcrypt 哈希存储；登录成功后返回 JWT Token。
- 鉴权中间件会解析 Bearer Token，并在上下文注入 userID。

## 常见问题
- FATAL: password authentication failed
  - 请检查 `.env` 中的 DB_USER/DB_PASSWORD/DB_NAME 是否与本地 Postgres 实际配置一致。
- 数据库不存在
  - 可创建目标数据库（如 blog），或把 DB_NAME 改为 postgres。
- 跨域
  - 默认 CORS 较宽松，生产可在 `internal/middleware/cors.go` 中收紧。
- JWT 安全
  - 生产必须修改 `JWT_SECRET`，并考虑 Token 过期策略、刷新机制等。

## 下一步可扩展
- 增加 swagger 文档（gin-swagger）
- 统一错误码与响应体
- 增加搜索接口、批量操作、回收站
- 引入配置校验、结构化日志、追踪与指标
- 单元测试与集成测试

---
如果你需要一键的 docker-compose（Postgres + Redis + 本服务），或补充 API 文档与测试，我可以继续为你加入配置与脚本。
