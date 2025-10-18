# API 参考 — HYH-Blog-Gin

本文档列出了 HYH-Blog-Gin 服务的主要 HTTP 接口、请求/响应示例以及鉴权说明。

基础 URL

- 本地开发默认：`http://localhost:8080`

鉴权

- 受保护的接口需要在 HTTP 头中包含：

```
Authorization: Bearer <JWT_TOKEN>
```

- 可通过 `POST /api/login` 获取 token。

响应格式

所有响应使用统一封装：

```json
{
  "code": 0,
  "message": "success",
  "data": "...", 
  "meta": { "page": 1, "limit": 10, "total": 100 }
}
```

- `code = 0` 表示成功，非零表示业务错误。

接口清单

1) 注册（Register）

- 方法：POST
- 路径：`/api/register`
- 鉴权：公开
- 请求 JSON：

```json
{
  "email": "you@example.com",
  "username": "yourname",
  "password": "secret"
}
```

- 成功：HTTP 201
- 成功示例数据：

```json
{
  "id": 1,
  "email": "you@example.com",
  "username": "yourname"
}
```

错误：400 Bad Request（校验或重复）

---

2) 登录（Login）

- 方法：POST
- 路径：`/api/login`
- 鉴权：公开
- 请求 JSON：

```json
{
  "identifier": "yourname or you@example.com",
  "password": "secret"
}
```

- 成功：HTTP 200
- 成功示例数据：

```json
{
  "token": "eyJhbGciOiJI..."
}
```

错误：401 Unauthorized（凭据无效）

---

3) 获取个人信息（Get profile）

- 方法：GET
- 路径：`/api/user/profile`
- 鉴权：需要
- 请求：头部 `Authorization: Bearer <token>`
- 成功：HTTP 200
- 成功示例（service 已去除 password）：

```json
{
  "id": 1,
  "email": "you@example.com",
  "username": "yourname",
  "notes": []
}
```

错误：401 Unauthorized、404 Not Found

---

4) 笔记列表（Notes — List）

- 方法：GET
- 路径：`/api/notes`
- 鉴权：需要
- 查询参数：
  - `page`（int，默认 1）
  - `limit`（int，默认 10）
- 成功：HTTP 200
- 示例响应数据：

```json
{
  "data": [ { "id": 10, "title": "...", "author_id": 1, "tags": [{"id":1,"name":"go"}] } ],
  "meta": { "page": 1, "limit": 10, "total": 42 }
}
```

---

5) 创建笔记（Notes — Create）

- 方法：POST
- 路径：`/api/notes`
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

- 成功：HTTP 201
- 成功返回：包含创建的笔记（含 tags 与 author）

错误：400 Bad Request

---

6) 查看单条笔记（Notes — Get one）

- 方法：GET
- 路径：`/api/notes/:id`
- 鉴权：需要
- 行为：当 `IsPublic == true` 或 请求者为作者时返回笔记，否则返回 403
- 成功：HTTP 200
- 错误：400（id 无效）、403、404

---

7) 更新笔记（Notes — Update）

- 方法：PUT
- 路径：`/api/notes/:id`
- 鉴权：需要（仅作者可更新）
- 请求 JSON（字段可选）：

```json
{
  "title": "New title",
  "content": "Updated",
  "tags": ["a","b"],   
  "public": false
}
```

- 成功：HTTP 200（返回更新后的笔记）
- 错误：400/403/404

说明：`tags: null` 表示不修改标签；`tags: []` 表示将标签替换为空集合。

---

8) 删除笔记（Notes — Delete）

- 方法：DELETE
- 路径：`/api/notes/:id`
- 鉴权：需要（仅作者）
- 成功：HTTP 204 No Content
- 错误：403/404

---

状态码与错误映射

- 200 OK — 成功（读取/更新）
- 201 Created — 创建成功
- 204 No Content — 删除成功
- 400 Bad Request — 校验/解析错误
- 401 Unauthorized — 缺少或无效 JWT
- 403 Forbidden — 已鉴权但无权限
- 404 Not Found — 资源不存在
- 500 Internal Server Error — 非预期服务错误


Postman

- 导入 `api/docs/HYH-Blog-Gin.postman_collection.json` 以便尝试这些接口。


开发者/实现者备注

- service 层负责业务逻辑并返回已脱敏的用户结构（password 已移除）。
- 仓储层的 `CreateWithTags` / `UpdateWithTags` 在单个数据库事务中处理标签的创建与关联，保证原子性。


后续（可选）

如需，我可以生成一个最小的 OpenAPI（YAML）文件或在 `api/swagger.yaml` 中提供，并添加一个 `/docs` 路由来展示 OpenAPI UI。你更倾向于哪种形式？
