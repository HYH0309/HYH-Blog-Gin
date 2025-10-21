# API 参考 — HYH-Blog-Gin

本文档面向前端开发者，列出后端主要 HTTP 接口、请求/响应示例与鉴权说明。文档以本地开发（http://localhost:8080）为基础。

基础信息
--------
- 服务根地址（本地开发）：`http://localhost:8080`
- API 前缀：`/api/v1`

鉴权（JWT）
-----------
- 受保护接口必须在 HTTP 头中携带：

```
Authorization: Bearer <JWT_TOKEN>
```

- 获取 Token：调用 `POST /api/v1/login`，成功响应的 `data.token` 中返回 JWT。

统一响应格式
------------
所有接口均使用统一响应包装：

```json
{
  "code": 0,
  "message": "success",
  "data": "...",  
  "meta": {"page":1, "limit":10, "total":100} 
}
```
- `code = 0` 表示成功；非 0 表示业务错误（message 给出可读错误信息）。

常用 HTTP 状态及含义
-------------------
- 200 OK — 成功（读取/更新）
- 201 Created — 创建成功
- 204 No Content — 删除成功（或无返回体）
- 400 Bad Request — 参数校验或解析错误
- 401 Unauthorized — 未鉴权或 token 无效
- 403 Forbidden — 无权限（已鉴权但无权限）
- 404 Not Found — 资源不存在
- 500 Internal Server Error — 服务内部错误

接口清单（示例请求与响应）
------------------------
说明：下面示例均以 curl 展示，所有路径以 `http://localhost:8080/api/v1` 为前缀。

1) 注册 — POST /api/v1/register
- 说明：使用 email、username、password 注册新用户（公开）。
- 请求示例：

```bash
curl -X POST "http://localhost:8080/api/v1/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","username":"yourname","password":"secret"}'
```
- 成功响应（HTTP 201，data 为注册用户）：

```json
{
  "code": 0,
  "message": "created",
  "data": {"id": 1, "email": "you@example.com", "username": "yourname"}
}
```

2) 登录 — POST /api/v1/login
- 说明：使用用户名或邮箱 + 密码登录，返回 JWT。
- 请求示例：

```bash
curl -X POST "http://localhost:8080/api/v1/login" \
  -H "Content-Type: application/json" \
  -d '{"identifier":"yourname","password":"secret"}'
```
- 成功响应（HTTP 200）：

```json
{
  "code": 0,
  "message": "success",
  "data": {"token": "eyJ..."}
}
```
- 使用方法：在后续受保护请求中把该 token 放到 `Authorization: Bearer <token>` 头中。

3) 获取当前用户信息 — GET /api/v1/user/profile
- 鉴权：需要
- 请求示例：

```bash
curl -X GET "http://localhost:8080/api/v1/user/profile" \
  -H "Authorization: Bearer <token>"
```
- 成功响应（HTTP 200，示例）：

```json
{
  "code": 0,
  "message": "success",
  "data": {"id":1,"email":"you@example.com","username":"yourname","notes":[]}
}
```

4) 笔记 — 列表 GET /api/v1/notes
- 鉴权：需要
- 查询参数：`page`（默认 1），`limit`（默认 10）
- 请求示例：

```bash
curl -X GET "http://localhost:8080/api/v1/notes?page=1&limit=10" \
  -H "Authorization: Bearer <token>"
```
- 成功响应（data 为数组，meta 包含分页信息）：

```json
{
  "code": 0,
  "message": "success",
  "data": [ {"id":10,"title":"...","author_id":1,"tags":[{"id":1,"name":"go"}] } ],
  "meta": {"page":1,"limit":10,"total":42}
}
```

5) 创建笔记 — POST /api/v1/notes
- 鉴权：需要
- 请求 JSON：

```json
{
  "title": "My Note",
  "content": "Hello world",
  "tags": ["go","gin"],
  "public": true
}
```
- 请求示例：

```bash
curl -X POST "http://localhost:8080/api/v1/notes" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title":"My Note","content":"Hello world","tags":["go","gin"],"public":true}'
```
- 成功响应（HTTP 201，data 为创建后的笔记对象）：

```json
{
  "code": 0,
  "message": "created",
  "data": {"id": 123, "title":"My Note", "tags": [{"id":1,"name":"go"}], "author_id": 1}
}
```

6) 获取单条笔记 — GET /api/v1/notes/{id}
- 鉴权：需要（如果笔记非公开且请求者不是作者则返回 403）
- 请求示例：

```bash
curl -X GET "http://localhost:8080/api/v1/notes/123" \
  -H "Authorization: Bearer <token>"
```
- 错误示例：
  - 400：id 无效
  - 403：无权限
  - 404：未找到

7) 更新笔记 — PUT /api/v1/notes/{id}
- 鉴权：需要（仅作者）
- 请求 JSON（字段可选）：

```json
{ "title": "New title", "content": "Updated", "tags": ["a","b"], "public": false }
```
- 说明：
  - `tags: null` 表示不修改标签集合；
  - `tags: []` 表示把标签替换为空集合。

8) 删除笔记 — DELETE /api/v1/notes/{id}
- 鉴权：需要（仅作者）
- 成功：HTTP 204 或统一包装的成功响应（具体实现可能返回 message）

9) 给笔记点赞 — POST /api/v1/notes/{id}/like
- 鉴权：需要
- 说明：高频写操作，先写入 Redis，再由后台同步到数据库。

10) 标签 — /api/v1/tags
- 列表：GET `/api/v1/tags?page=1&per_page=20`（鉴权）
- 创建：POST `/api/v1/tags`，body: `{ "name": "tech" }`（鉴权）
- 单个：GET/PUT/DELETE `/api/v1/tags/{id}`（鉴权）

11) 图片管理
- 上传：POST `/api/v1/images`（multipart/form-data，字段 `file`，可选 `filename`），返回图片 URL（鉴权）

```bash
curl -X POST "http://localhost:8080/api/v1/images" \
  -H "Authorization: Bearer <token>" \
  -F "file=@/path/to/img.jpg" -F "filename=cover.jpg"
```

- 列表：GET `/api/v1/images?page=1&per_page=50`（鉴权）
- 元信息：GET `/api/v1/images/info?url=/uploads/xxx.webp`（鉴权）
- 删除：DELETE `/api/v1/images?url=/uploads/xxx.webp`（鉴权）

错误响应示例
-------------
统一错误示例（HTTP 4xx/5xx）：

```json
{
  "code": 400,
  "message": "invalid request: title is required",
  "data": null
}
```

开发者备注与工具
---------------
- 后端使用统一响应（见顶部），service 层返回的用户对象已去除 password 字段。
- 项目内包含 Swagger/OpenAPI 定义：`api/swagger.yaml`（或由 `swag init` 生成）。
- 本地查看方式：可使用 `swagger-ui` 或将生成的 `api/docs.go` 编译进服务并通过 `/swagger` 路由访问。
- Postman：可导入仓库内的 Postman collection（如有）。

如果你希望，我可以：
- 将文档进一步转换为英文版；
- 自动生成前端 TypeScript client（基于 OpenAPI）；
- 在服务中挂载 swagger UI 并提交对应的路由实现。

---

文档更新：保持该文件与后端路由同步；如后端有新接口或变动，请告知我来更新该文档。
