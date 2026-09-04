# Gopher Social Network

![Go Version](https://img.shields.io/badge/Go-1.27-00ADD8?style=flat-square&logo=go)
![Vue Version](https://img.shields.io/badge/Vue-3.5-4FC08D?style=flat-square&logo=vuedotjs)
![Docker](https://img.shields.io/badge/Docker-24.0-2496ED?style=flat-square&logo=docker)
![Redis](https://img.shields.io/badge/Redis-7.2-DC382D?style=flat-square&logo=redis)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16.3-336791?style=flat-square&logo=postgresql)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)

基于 Go 后端和 Vue 3 前端构建的社交网络应用，支持用户认证、帖子发布、评论互动、关注系统等功能。

## 功能特性

- **用户系统**：注册、登录、邮箱激活、密码重置
- **帖子管理**：创建、编辑、删除帖子，支持标签和搜索
- **评论系统**：多级评论和回复，支持点赞
- **关注系统**：关注/取消关注用户，查看动态流
- **头像上传**：支持图片上传和裁剪
- **实时通知**：邮件通知和密码重置
- **性能优化**：Redis 缓存和速率限制

## 技术栈

### 后端
- **语言**：Go 1.27
- **路由**：Chi Router
- **数据库**：PostgreSQL 16
- **缓存**：Redis 7
- **认证**：JWT (JSON Web Token)
- **邮件**：MailTrap
- **日志**：Zap
- **热重载**：Air

### 前端
- **框架**：Vue 3.5
- **语言**：TypeScript
- **构建工具**：Vite 8
- **UI 库**：Element Plus
- **状态管理**：Pinia
- **路由**：Vue Router 4

## 快速开始

### 先决条件

- Go 1.27+
- Node.js 22+ 或 24+
- PostgreSQL 16
- Redis 7
- Docker（可选，用于容器化部署）

### 安装步骤

1. **克隆仓库**
   ```bash
   git clone https://github.com/loeia/gopher-social-network.git
   cd gopher-social-network
   ```

2. **配置环境变量**
   ```bash
   cp .envrc.example .envrc  # 编辑 .envrc 文件，填入你的配置
   cp web/.env.example web/.env
   ```

3. **启动数据库服务**
   ```bash
   # 使用 Docker Compose 启动 PostgreSQL 和 Redis
   docker compose up -d
   ```

4. **数据库迁移**
   ```bash
   make seed
   ```

5. **启动后端服务**
   ```bash
   make run
   ```

6. **启动前端服务**
   ```bash
   cd web
   npm install
   npm run dev
   ```

### Docker 部署

```bash
# 一键启动所有服务
make up

# 停止所有服务
make down
```

## 环境变量配置

### 后端环境变量 (.envrc)

#### 服务器配置
- `ADDR`：服务器监听地址，默认 `:8080`
- `ENV`：运行环境，默认 `development`，可选 `production`
- `FRONTEND_URL`：前端地址，默认 `http://localhost:5173`

#### 数据库配置
- `DB_DSN`：PostgreSQL 连接字符串，默认 `postgres://admin:admin123@localhost/gopher-social-network?sslmode=disable`
- `DB_MAX_OPEN_CONNS`：最大打开连接数，默认 `30`
- `DB_MAX_IDLE_CONNS`：最大空闲连接数，默认 `30`
- `DB_MAX_IDLE_TIME`：空闲连接超时时间，默认 `15m`
- `DB_MAX_LIFE_TIME`：连接最大生命周期，默认 `5m`

#### 认证配置
- `AUTH_SECRET`：JWT 密钥（必须设置）
- `AUTH_EXP`：JWT 过期时间（小时），默认 `72`
- `AUTH_ISSUER`：JWT 签发者，默认 `Bearer`

#### 邮件配置
- `FROM_EMAIL`：发件人邮箱（必须设置）
- `MAILTRAP_API_KEY`：MailTrap API 密钥
- `MAILTRAP_SANDBOX_USER`：MailTrap 沙箱用户名
- `MAILTRAP_SANDBOX_PASS`：MailTrap 沙箱密码

#### Redis 配置
- `REDIS_ENABLED`：是否启用 Redis，默认 `false`
- `REDIS_ADDR`：Redis 地址，默认 `localhost:6379`
- `REDIS_PASSWORD`：Redis 密码
- `REDIS_DB`：Redis 数据库编号，默认 `0`

#### 速率限制配置
- `RATE_LIMITER_ENABLED`：是否启用速率限制，默认 `true`
- `RATELIMITER_REQUESTS_COUNT`：时间窗口内请求数限制，默认 `20`
- `RATELIMITER_TIME_FRAME`：时间窗口（秒），默认 `5`

#### CORS 配置
- `CORS_ALLOWED_ORIGIN`：允许的跨域来源，默认 `http://localhost:5173`

#### Goose 迁移配置（可选）
- `GOOSE_DRIVER`：数据库驱动，默认 `pgx`
- `GOOSE_DBSTRING`：迁移数据库连接字符串
- `GOOSE_MIGRATION_DIR`：迁移文件目录，默认 `./cmd/migrate/migrations/`

### 前端环境变量 (web/.env)

- `VITE_API_URL`：后端 API 地址，默认 `http://localhost:8080`

## API 路由文档

### 健康检查
- `GET /health`：检查服务是否正常运行

### 帖子相关 (/posts)
- `GET /posts/search`：搜索帖子（支持分页和过滤）
- `GET /posts/free`：获取随机帖子列表
- `POST /posts`：创建新帖子
- `GET /posts/{postId}`：获取帖子详情（公开）
- `PATCH /posts/{postId}`：更新帖子
- `DELETE /posts/{postId}`：删除帖子
- `PUT /posts/{postId}/like`：点赞帖子
- `DELETE /posts/{postId}/like`：取消点赞

### 帖子评论 (/posts/{postId}/comments)
- `GET /posts/{postId}/comments`：获取帖子评论列表（支持分页）
- `POST /posts/{postId}/comments`：创建评论
- `POST /posts/{postId}/comments/{commentId}/replies`：回复评论

### 用户相关 (/users)
- `GET /users/feed`：获取用户动态流
- `PATCH /users/me/password`：重置密码
- `PUT /users/me/avatar`：上传头像
- `DELETE /users/me/avatar`：删除头像
- `PATCH /users/me/profile`：更新个人资料
- `GET /users/me/comment-likes`：获取用户点赞的评论
- `PATCH /users/me/username`：修改用户名

### 用户详情 (/users/{userId})
- `GET /users/{userId}`：获取用户信息（公开）
- `GET /users/{userId}/posts`：获取用户帖子列表
- `GET /users/{userId}/post-likes`：获取用户点赞的帖子
- `GET /users/{userId}/comments`：获取用户评论列表
- `GET /users/{userId}/followers`：获取用户粉丝列表
- `GET /users/{userId}/following`：获取用户关注列表
- `PUT /users/{userId}/follow`：关注用户
- `DELETE /users/{userId}/follow`：取消关注
- `GET /users/{userId}/avatar`：获取用户头像

### 评论相关 (/comments)
- `GET /comments/{commentId}`：获取评论详情（公开）
- `GET /comments/{commentId}/replies`：获取评论回复列表
- `DELETE /comments/{commentId}`：删除评论
- `PUT /comments/{commentId}/like`：点赞评论
- `DELETE /comments/{commentId}/like`：取消点赞

### 认证相关 (/auth)
- `POST /auth/register`：用户注册
- `POST /auth/token`：用户登录，获取 JWT 令牌
- `POST /auth/activate/{token}`：激活用户账号
- `POST /auth/forgot-password`：发送密码重置邮件
- `POST /auth/reset-password`：通过令牌重置密码

### 管理员功能 (/admin)
- `PATCH /admin/users/{username}/ban`：封禁用户
- `PATCH /admin/users/{username}/unban`：解封用户

## 项目结构

```
gopher-social-network/
├── cmd/                    # 应用程序入口
│   ├── api/               # API 服务器
│   └── migrate/           # 数据库迁移工具
├── internal/              # 内部包
│   ├── auth/              # 认证模块
│   ├── avatar/            # 头像生成
│   ├── db/                # 数据库连接
│   ├── env/               # 环境变量处理
│   ├── mailer/            # 邮件服务
│   ├── ratelimiter/       # 速率限制
│   └── store/             # 数据存储层
├── web/                   # 前端应用
│   ├── src/               # 源代码
│   ├── public/            # 静态资源
│   └── package.json       # 前端依赖
├── docker-compose.yaml    # Docker 编排文件
├── Makefile              # 构建脚本
├── go.mod                # Go 模块文件
└── .envrc                # 环境变量配置
```

## 截图

### 登录页面
![Login](screenshots/login.png)

### 首页
![Home](screenshots/home.png)

### 帖子详情
![Post Detail](screenshots/post-detail.png)

### 创建帖子
![Create Post](screenshots/create-post.png)

### 搜索
![Search](screenshots/search.png)

### 个人主页
![Profile](screenshots/profile.png)

### 订阅列表
![Subscription](screenshots/subscription.png)

## 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

### 代码规范

- 后端遵循 Go 官方代码规范
- 前端使用 ESLint 和 Prettier 格式化代码
- 提交信息使用 Conventional Commits 规范

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件
