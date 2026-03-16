# InternHub

实习/校招信息聚合与投递管理平台（微服务架构）。

## 项目概述

InternHub 采用 Go 语言微服务架构，包含 API 网关、认证、用户、职位、投递、推荐等服务。

## 认证与用户架构

- **auth-service** 为**唯一用户与认证源**：负责注册、登录、JWT 签发；用户表存于 PostgreSQL（默认 8081）。
- **user-service** 仅负责**用户扩展资料**：提供 `GET/PATCH /api/v1/users/me`（昵称、头像等），数据存于 PostgreSQL（默认 8082）；依赖网关注入的 `X-User-Id`。
- 注册、登录经 **api-gateway** 转发到 auth-service；需登录的 `/users/me` 经网关校验 JWT 后转发到 user-service。

## 代码完整度

| 模块           | 状态     | 说明 |
|----------------|----------|------|
| api-gateway    | ✅ 已实现 | 注册/登录代理、JWT 中间件、/users/me 代理、/health、/metrics |
| auth-service   | ✅ 已实现 | 注册、登录、JWT，PostgreSQL |
| user-service   | ✅ 已实现 | 用户资料 GET/PATCH /users/me，PostgreSQL |
| job-service    | ⏳ 占位  | 仅有 go.mod |
| apply-service  | ⏳ 占位  | 仅有 go.mod |
| recommend-service | ⏳ 占位 | 仅有 go.mod |
| pkg/logger     | ✅ 已实现 | zap 日志 |

## 技术栈

- Go 1.22+
- Gin、GORM、PostgreSQL
- Prometheus（api-gateway /metrics）

## 项目结构

```
internhub/
├── api-gateway/     # 网关 :8080
├── auth-service/    # 认证 :8081
├── user-service/    # 用户资料 :8082
├── job-service/    # 职位（占位）
├── apply-service/  # 投递（占位）
├── recommend-service/
├── pkg/logger/
├── docker-compose.yml
├── go.work
└── README.md
```

## 环境要求

- Go 1.22+
- PostgreSQL（可用 Docker 启动）

## 快速开始

### 1. 依赖

```bash
cd internhub
go work sync
```

### 2. 启动 PostgreSQL

```bash
docker compose up -d
```

默认库 `internhub`，用户 `postgres`，密码 `postgres`，端口 5432。

### 3. 启动服务

开四个终端，在项目根目录执行：

```bash
go run ./api-gateway/cmd
go run ./auth-service
go run ./user-service/cmd
```

### 4. 验证

- 健康：`curl http://localhost:8080/health`
- 注册：`curl -X POST http://localhost:8080/api/v1/users/register -H "Content-Type: application/json" -d '{"email":"a@b.com","password":"password123","name":"Test"}'`
- 登录：`curl -s -X POST http://localhost:8080/api/v1/users/login -H "Content-Type: application/json" -d '{"email":"a@b.com","password":"password123"}'`（返回中的 `access_token` 用于下一步）
- 受保护：`curl -H "Authorization: Bearer <access_token>" http://localhost:8080/api/v1/protected`
- 当前用户资料：`curl -H "Authorization: Bearer <access_token>" http://localhost:8080/api/v1/users/me`

## 配置说明

### api-gateway

| 环境变量         | 默认值               |
|------------------|----------------------|
| AUTH_SERVICE_URL | http://127.0.0.1:8081 |
| USER_SERVICE_URL | http://127.0.0.1:8082 |
| JWT_SECRET       | internhub-secret     |

### auth-service / user-service

| 环境变量   | 默认值    |
|------------|-----------|
| PORT       | 8081 / 8082 |
| PG_HOST    | localhost |
| PG_USER    | postgres  |
| PG_PASSWORD| postgres  |
| PG_PORT    | 5432      |
| PG_DATABASE| internhub |
| JWT_SECRET | internhub-secret（仅 auth） |

生产环境请通过环境变量设置 JWT_SECRET、PG_PASSWORD 等，勿用默认值。

## 开发说明

- 根目录 `go work` 可一次加载所有模块。
- 构建：`go build ./api-gateway/cmd`、`go build ./auth-service`、`go build ./user-service/cmd`。
