# 校园技能交换

**项目类型标签：全栈Web应用**

校园技能交换 是 大学生 的技能互助与交换预约平台。

## 快速启动

开发模式：

```bash
npm install
npm run dev
```

访问地址：`http://localhost:18629`

生产构建：

```bash
npm run build
npm run preview
```

## 主要功能

- 技能发布管理：作品、时间段和回报类型
- 需求发布浏览：筛选和响应
- 智能匹配推荐：互补技能和共同时间
- 交换预约确认：线上线下地点协商
- 评价信用体系：评分、等级和权重
- 消息通知：会话和系统提醒
- 个人主页技能墙：雷达图和历史记录

## 本地开发方式

在项目根目录执行：

```bash
npm install
npm run dev
```

## 技术栈

| 分类 | 技术 |
| --- | --- |
| 前端 | Vue 3 + TypeScript + Vite + Element Plus |
| 后端 | Go + Gin |
| 数据库 | MySQL 8.0 |
| 缓存 | Redis |
| 认证 | JWT |

## 项目目录结构

```text
. 
├── src
│   ├── components
│   ├── constants
│   ├── data
│   ├── errors
│   ├── features
│   ├── logger
│   ├── services
│   ├── types
│   └── utils
├── Dockerfile
├── nginx.conf
├── package.json
└── README.md
```

## 环境变量说明

纯前端项目默认不需要后端环境变量。地图或第三方 API Key 可放入本地 .env 文件。

## 使用说明

应用数据存储在浏览器本地。清空浏览器站点数据会重置演示数据。

## License

MIT
