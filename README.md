# 校园技能交换

**项目类型标签：全栈Web应用**

校园技能交换是面向大学生的技能互助平台，覆盖技能发布、需求浏览、智能匹配、预约确认、评价信用、站内消息和个人技能墙。

## 快速启动

首次启动前复制环境变量，然后一键启动：

```bash
cp .env.example .env
docker compose up -d
```

访问地址：`http://localhost:18629`  
后端健康检查：`http://localhost:19629/api/health`

## 主要功能

- 技能发布与管理：技能描述、熟练度、可交换时间段、回报类型和作品凭证。
- 需求发布与浏览：按类别、校区、期望时间和响应数量查看求助需求。
- 智能匹配推荐：展示互补技能、匹配度、共同可用时间和推荐理由。
- 交换预约与确认：双方确认后锁定生效；待确认时仅发起人可撤回，撤回即释放双方时间档；锁定后只能提交取消原因等待对方处理；状态变更持久化，刷新后可回读。
- 评价与信用体系：评分、文字评价、信用分和信用等级用于推荐权重。
- 消息通知系统：会话未读红点、系统通知和预约提醒。
- 个人主页与技能墙：历史交换、收到评价和 ECharts 技能雷达图。

## 本地开发方式

前端：

```bash
cd frontend
npm install
npm run dev
```

后端：

```bash
cd backend
go run ./cmd/server
```

本地开发后端默认监听 `19629`，前端开发服务器已代理 `/api` 到该端口。

## 技术栈

| 分类 | 技术 |
| --- | --- |
| 前端 | Vue 3 + TypeScript + Vite + Element Plus + ECharts |
| 后端 | Go + Gin |
| 数据库 | MySQL 8.0 |
| 缓存 | Redis |
| 认证 | JWT 配置预留 |
| 部署 | Docker Compose + Nginx |

## 项目目录结构

```text
.
├── frontend
│   ├── src/components
│   ├── src/features
│   ├── src/services
│   ├── src/types
│   ├── Dockerfile
│   └── nginx.conf
├── backend
│   ├── cmd/server
│   ├── internal/controller
│   ├── internal/repository
│   ├── internal/service
│   └── Dockerfile
├── database
│   └── init.sql
├── docker-compose.yml
├── .env.example
└── README.md
```

## API 概览

- `GET /api/health`
- `GET /api/dashboard/overview`
- `GET /api/skills`
- `GET /api/needs`
- `GET /api/matches`
- `GET /api/appointments` / `GET /api/appointments/:id`
- `POST /api/appointments/:id/confirm`：任一方确认，Body `{"actor":"姓名"}`，双方确认后锁定
- `POST /api/appointments/:id/withdraw`：待确认时仅发起人可撤回，撤回即释放双方时间档
- `POST /api/appointments/:id/cancel-requests`：锁定后提交取消原因，Body `{"actor":"姓名","reason":"原因"}`
- `POST /api/appointments/:id/cancel-requests/respond`：对方处理取消申请，Body `{"actor":"姓名","action":"accept|reject"}`
- `GET /api/slots?user=姓名`：查询被进行中预约占用的时间档
- `GET /api/reviews`
- `GET /api/messages`
- `GET /api/profile`

预约状态机：`pending`（待确认）→ 双方确认后 `confirmed`（锁定）；`pending` 时发起人可撤回为 `withdrawn`（终态）；`confirmed` 后提交取消原因进入 `cancel_pending`，对方同意则 `cancelled`（终态）、拒绝则回到 `confirmed`。确认与撤回同时到达时只有一个生效，终态下的重复请求幂等返回、不改变状态。

## 环境变量说明

| 变量 | 说明 |
| --- | --- |
| COMPOSE_PROJECT_NAME | Compose 项目名，默认 `cyskillswap` |
| FRONTEND_PORT | 前端映射端口，默认 `18629` |
| BACKEND_PORT | 后端映射端口，默认 `19629` |
| DB_PORT | MySQL 宿主机映射端口，默认 `17629` |
| DB_NAME / DB_USER / DB_PASSWORD / DB_ROOT_PASSWORD | 数据库连接信息 |
| JWT_SECRET | JWT 签名密钥 |
| APPOINTMENT_STORE_PATH | 预约状态持久化文件路径，容器内默认 `/app/data/appointments.json` |

## Docker 部署说明

- Compose 顶层 `name: cyskillswap`，并在 `.env` 中提供 `COMPOSE_PROJECT_NAME=cyskillswap`，可在中文目录下启动。
- 前端容器通过 Nginx 托管静态资源，并将 `/api/` 反向代理到后端服务，前端代码只请求 `/api`。
- MySQL 和 Redis 使用命名卷持久化，避免绑定挂载中文路径。
- 预约状态通过 `backend_data` 命名卷持久化到 `/app/data/appointments.json`，容器重建后仍可回读。
- 端口冲突时修改 `.env` 中的 `FRONTEND_PORT`、`BACKEND_PORT` 或 `DB_PORT` 后重新执行 `docker compose up -d`。

## License

MIT
