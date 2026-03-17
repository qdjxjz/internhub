# InternHub

实习/校招信息聚合与投递管理平台（微服务架构）。



## Demo

https://github.com/user-attachments/assets/6c2bb26a-a465-4a47-97dc-5804a569553d



## 架构图

![InternHub 架构图](images/internhub316.png)

## 项目概述

InternHub 采用 Go 语言微服务架构，包含 API 网关、认证、用户、职位、投递、推荐等服务。

## 认证与用户架构

- **auth-service** 为**唯一用户与认证源**：负责注册、登录、JWT 签发；用户表存于 PostgreSQL（默认 8081）。
- **user-service** 仅负责**用户扩展资料**：提供 `GET/PATCH /api/v1/users/me`（昵称、头像等），数据存于 PostgreSQL（默认 8082）；依赖网关注入的 `X-User-Id`。
- 注册、登录经 **api-gateway** 转发到 auth-service；需登录的 `/users/me` 经网关校验 JWT 后转发到 user-service。

## 代码完整度


| 模块                | 状态    | 说明                                                             |
| ----------------- | ----- | -------------------------------------------------------------- |
| api-gateway       | ✅ 已实现 | 注册/登录/jobs/applications 代理、JWT、/users/me、/health、/metrics      |
| auth-service      | ✅ 已实现 | 注册、登录、JWT，PostgreSQL                                           |
| user-service      | ✅ 已实现 | 用户资料 GET/PATCH /users/me，PostgreSQL                            |
| job-service       | ✅ 已实现 | 职位 GET/POST /jobs、GET /jobs/:id，PostgreSQL（:8083）              |
| apply-service     | ✅ 已实现 | 投递 POST /applications、GET /applications/me，PostgreSQL（:8084）   |
| recommend-service | ✅ 已实现 | 岗位推荐 GET /recommendations（需登录），可接 ChatGPT API 做 AI 推荐，默认 :8085 |
| pkg/logger        | ✅ 已实现 | zap 日志                                                         |


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
├── job-service/     # 职位 :8083
├── apply-service/   # 投递 :8084
├── recommend-service/ # 推荐 :8085
├── web/             # 前端 Vite + React，开发时 :5173
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

**方式一：一键启动（推荐）**

在项目根目录执行一条命令即可启动全部六个服务，同一终端内按 Ctrl+C 会一起停止：

```bash
./scripts/start-all.sh
```

**方式二：分终端启动**

开六个终端，在项目根目录分别执行：

```bash
go run ./api-gateway/cmd
go run ./auth-service
go run ./user-service/cmd
go run ./job-service/cmd
go run ./apply-service/cmd
go run ./recommend-service/cmd
```

### 4. 启动前端

前端为 Vite + React + TypeScript，通过代理访问网关（`http://localhost:8080`）。在**后端服务已启动**的前提下，新开终端执行：

```bash
cd web
npm install
npm run dev
```

浏览器访问 Vite 给出的本地地址（通常为 `http://localhost:5173`）即可使用职位列表、我的投递、岗位推荐、个人资料、登录/注册等功能。

### 5. 验证

- 健康：`curl http://localhost:8080/health`
- 注册：`curl -X POST http://localhost:8080/api/v1/users/register -H "Content-Type: application/json" -d '{"email":"a@b.com","password":"password123","name":"Test"}'`
- 登录：`curl -s -X POST http://localhost:8080/api/v1/users/login -H "Content-Type: application/json" -d '{"email":"a@b.com","password":"password123"}'`（返回中的 `access_token` 整段复制到下面请求）
- 受保护：`curl -H "Authorization: Bearer <access_token>" http://localhost:8080/api/v1/protected`
- 当前用户资料：`curl -H "Authorization: Bearer <access_token>" http://localhost:8080/api/v1/users/me`
- 职位列表：`curl http://localhost:8080/api/v1/jobs`
- 创建职位：`curl -X POST http://localhost:8080/api/v1/jobs -H "Content-Type: application/json" -d '{"title":"Go 实习","company":"某公司","link":"https://example.com"}'`
- 投递（需登录）：`curl -X POST http://localhost:8080/api/v1/applications -H "Authorization: Bearer <access_token>" -H "Content-Type: application/json" -d '{"job_id":1}'`
- 我的投递列表（需登录）：`curl -H "Authorization: Bearer <access_token>" http://localhost:8080/api/v1/applications/me`
- 岗位推荐（需登录）：`curl -H "Authorization: Bearer <access_token>" http://localhost:8080/api/v1/recommendations`

### 6. 一键功能测试（完整示例）

在**所有服务已启动**的前提下，在项目根目录执行：

```bash
./scripts/test-example.sh
```

脚本会依次执行：健康检查 → 注册 → 登录 → 创建 4 个职位 → 查职位列表 → **获取推荐（未投递时应为全部）** → 投递职位 1 和 2 → 查我的投递 → **再次获取推荐（应排除已投递的 1、2）** → 更新昵称 → 查当前用户资料。

- 若已安装 `jq`，输出会格式化并自动从登录响应中提取 token。
- 未安装 `jq` 也能运行，token 会通过 sed 提取。
- 多次执行时，同一邮箱会提示已注册，脚本会继续用该账号登录完成后续步骤。

## 配置说明

### api-gateway


| 环境变量                  | 默认值                                            |
| --------------------- | ---------------------------------------------- |
| AUTH_SERVICE_URL      | [http://127.0.0.1:8081](http://127.0.0.1:8081) |
| USER_SERVICE_URL      | [http://127.0.0.1:8082](http://127.0.0.1:8082) |
| JOB_SERVICE_URL       | [http://127.0.0.1:8083](http://127.0.0.1:8083) |
| APPLY_SERVICE_URL     | [http://127.0.0.1:8084](http://127.0.0.1:8084) |
| RECOMMEND_SERVICE_URL | [http://127.0.0.1:8085](http://127.0.0.1:8085) |
| JWT_SECRET            | internhub-secret                               |


### auth-service / user-service


| 环境变量        | 默认值                              |
| ----------- | -------------------------------- |
| PORT        | 8081 / 8082 / 8083 / 8084（各服务不同） |
| PG_HOST     | localhost                        |
| PG_USER     | postgres                         |
| PG_PASSWORD | postgres                         |
| PG_PORT     | 5432                             |
| PG_DATABASE | internhub                        |
| JWT_SECRET  | internhub-secret（仅 auth）         |


job-service、apply-service 使用相同 PG_* 环境变量，可与其它服务共用同一库。

### recommend-service


| 环境变量              | 默认值                                            |
| ----------------- | ---------------------------------------------- |
| PORT              | 8085                                           |
| JOB_SERVICE_URL   | [http://127.0.0.1:8083](http://127.0.0.1:8083) |
| APPLY_SERVICE_URL | [http://127.0.0.1:8084](http://127.0.0.1:8084) |
| USER_SERVICE_URL  | [http://127.0.0.1:8082](http://127.0.0.1:8082) |
| OPENAI_API_KEY    | （空则不做 AI 推荐，仅返回未投递列表）                          |
| OPENAI_BASE_URL   | 空（使用 OpenAI 默认）；可设为兼容接口地址（如 Azure、国内代理）        |


## 开发说明

- 根目录 `go work` 可一次加载所有模块。
- 构建：`go build ./api-gateway/cmd`、`go build ./auth-service`、`go build ./user-service/cmd`、`go build ./job-service/cmd`、`go build ./apply-service/cmd`、`go build ./recommend-service/cmd`。
- 测试：`go test ./recommend-service/...`（覆盖 recommend-service 的 ai、handler、service 单元/集成测试）。

