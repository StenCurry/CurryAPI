# Curry2API

> 一个可自部署的多模型 AI 网关，兼容 OpenAI / Claude 接口，内置用户与配额管理、使用统计、管理后台与前端仪表盘。
>
> 🚀 **完全开源，支持自行修改和部署**：可部署到自己的服务器，数据完全自主可控，适合个人与团队自托管。
> （旧版本）体验:www.kesug.icu


## 特性概览
- **多模型网关**：支持 GPT、Claude、Gemini、DeepSeek 等，统一入口与鉴权。
- **双协议兼容**：OpenAI Chat Completions & Anthropic Claude Messages。
- **流式输出**：支持 SSE 流式响应。
- **用户/密钥/配额**：注册登录、API Key 管理、配额限制、使用统计。
- **管理后台**：用户、会话、密钥、使用监控、数据导出。
- **使用分析**：按模型/时间的 Token 统计与趋势，支持仪表盘。
- **前端仪表盘**：现代极简风格，明暗主题，全局设计系统。

## 技术栈
- **后端**：Go 1.24+，Fiber / 标准库，SQLite（默认）
- **前端**：Vue 3 + Vite + TypeScript，Naive UI
- **测试**：Go `go test`、前端 `vitest`
- **构建**：Vite / vue-tsc

## 目录结构
```
Curry2API-go/
├── main.go                 # 后端入口
├── config/                 # 配置
├── handlers/               # HTTP 处理（OpenAI/Claude/Usage 等）
├── middleware/             # 认证、可选认证等
├── services/               # 业务逻辑（usage tracker 等）
├── database/               # 数据访问
│   └── schema.sql          # 数据库初始化脚本
├── models/                 # 数据模型
├── utils/                  # 工具集
├── static/                 # 静态文件
├── docs/                   # 文档
├── frontend/               # 前端项目 (Vue + Vite)
│   ├── src/
│   │   ├── views/          # 页面（Dashboard、Login、Chat 等）
│   │   ├── components/     # 组件（图表、卡片等）
│   │   └── style.css       # 全局设计系统
│   └── package.json
├── .env.example            # 环境变量示例
└── README.md               # 本文件
```

## 环境要求
- Go 1.24+（建议使用最新版）
- Node.js 18+ 与 npm
- MySQL 8.0+（必需）
- 操作系统：Windows / macOS / Linux

## 快速开始（本地开发）
### 1) 准备数据库
```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE cursor2api CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 导入表结构
mysql -u root -p cursor2api < database/schema.sql
```
> 默认会创建管理员账户：用户名 `admin`，密码 `admin123`，请首次登录后修改密码！

### 2) 准备后端
```bash
# 复制环境变量示例
cp .env.example .env
# 修改 .env 中的数据库连接配置
# 安装依赖
go mod download
# 启动后端
go run main.go
```
后端默认监听 `http://localhost:8002`。

### 2) 准备前端
```bash
cd frontend
npm install
npm run dev
```
前端开发服默认 `http://localhost:5173`，自动代理后端。

### 3) 访问
- 前端：`http://localhost:5173`
- 后端 API：`http://localhost:8002`（被前端代理，不要直接刷新后端路由）

## 配置说明(.env)
`.env.example` 提供示例，核心配置：
```env
PORT=8002                 # 后端端口
DEBUG=true                # 调试模式

# 数据库配置
DB_HOST=localhost         # 数据库主机
DB_PORT=3306              # 数据库端口
DB_USER=root              # 数据库用户名
DB_PASSWORD=your_password # 数据库密码
DB_NAME=cursor2api        # 数据库名称

# API 配置
API_KEY=0000              # 默认 API Key
MODELS=gpt-5.2,gpt-5,...  # 支持的模型列表
SYSTEM_PROMPT_INJECT=     # 全局系统提示词注入

# 请求配置
TIMEOUT=30                # 秒
USER_AGENT=Mozilla/5.0...

# Cursor 相关
SCRIPT_URL=https://cursor.com/...

# 使用跟踪
USAGE_TRACKING_ENABLED=true
USAGE_CHANNEL_SIZE=1000
USAGE_BATCH_SIZE=100
USAGE_FLUSH_INTERVAL=5
USAGE_MAX_RETRIES=3
USAGE_RETRY_BACKOFF_MS=100
USAGE_RETENTION_DAYS=90
USAGE_CLEANUP_HOUR=3
USAGE_CLEANUP_MINUTE=0
```

## 构建与测试
### 后端测试
```bash
go test ./...
```

### 前端检查 & 构建
```bash
cd frontend
npm install              # 首次需要
npm run build:check      # 等价于 vue-tsc -b && vite build（类型+构建）
npm run build            # 仅构建
npm run test             # vitest
```

## 生产部署
### 简单部署（单机）
```bash
# 构建前端
cd frontend
npm install
npm run build

# 启动后端（会服务前端静态文件）
cd ..
go run main.go
# 访问 http://localhost:8002
```

### 进程守护（示例）
- 使用 `systemd`、`supervisor` 或 `pm2`（通过 `pm2 start go run main.go`）。

### 反向代理（示例 Nginx）
```nginx
server {
  listen 80;
  server_name your.domain;
  location / {
    proxy_pass http://127.0.0.1:8002;
  }
}
```

### Docker 部署
#### 方案 A：单容器（后端内置静态资源）
1. 复制环境变量：
   ```bash
   cp .env.example .env
   # 根据需要修改 .env
   ```
2. 构建镜像：
   ```bash
   docker build -t curry2api-go:latest .
   ```
   > Dockerfile 应包含前端构建（`npm install && npm run build`）并将产物拷入后端静态目录，然后编译/运行 Go。
3. 启动容器：
   ```bash
   docker run -d --name curry2api-go \
     -p 8002:8002 \
     --env-file .env \
     curry2api-go:latest
   ```
4. 访问 `http://localhost:8002`。

#### 方案 B：前后端分容器
1. 先构建前端产物：
   ```bash
   cd frontend
   npm install
   npm run build
   ```
2. 前端 Nginx 镜像（示例 Dockerfile）：
   ```Dockerfile
   FROM nginx:alpine
   COPY dist/ /usr/share/nginx/html
   COPY docker/nginx-frontend.conf /etc/nginx/conf.d/default.conf
   ```
   构建：
   ```bash
   docker build -t curry2api-frontend:latest -f docker/frontend.Dockerfile .
   ```
3. 后端镜像（仅 Go）：
   ```bash
   docker build -t curry2api-backend:latest -f docker/backend.Dockerfile .
   docker run -d --name curry2api-backend \
     -p 8002:8002 \
     --env-file .env \
     curry2api-backend:latest
   ```
4. 前端容器：
   ```bash
   docker run -d --name curry2api-frontend \
     -p 80:80 \
     curry2api-frontend:latest
   ```
5. 前端通过 Nginx 将 API 请求代理到后端（示例 `nginx-frontend.conf`）：
   ```nginx
   server {
     listen 80;
     server_name _;
     root /usr/share/nginx/html;
     location /api/ {
       proxy_pass http://curry2api-backend:8002;
     }
     location /v1/ {
       proxy_pass http://curry2api-backend:8002;
     }
     try_files $uri $uri/ /index.html;
   }
   ```

#### 方案 C：docker-compose（推荐）
`docker-compose.yml` 示例：
```yaml
version: "3.8"
services:
  backend:
    build:
      context: .
      dockerfile: docker/backend.Dockerfile
    env_file:
      - .env
    ports:
      - "8002:8002"
    restart: unless-stopped

  frontend:
    build:
      context: .
      dockerfile: docker/frontend.Dockerfile
    depends_on:
      - backend
    ports:
      - "80:80"
    restart: unless-stopped
```
启动：
```bash
docker compose up -d --build
```

#### 镜像优化与注意事项
- 使用多阶段构建：前端 build、Go build、最终 runtime。
- 非 root 运行：在最终镜像中使用非 root 用户。
- 健康检查：在 compose 或 K8s 中加入 `/health` 等。
- 卷挂载：如需持久化日志或数据库，可挂载到宿主机路径。


## 运行与验证
1. 后端启动：日志无报错，监听 8002。
2. 前端启动：访问 5173（或生产 8002）能看到登录页。
3. 构建检查：`npm run build:check` 通过。
4. 后端自测：`go test ./...` 通过。

## API 速览与示例
### OpenAI Chat Completions
```bash
curl -X POST http://localhost:8002/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-api-key" \
  -d '{
    "model": "gpt-5.1-codex-max",
    "messages": [{"role": "user","content": "Hello!"}],
    "stream": false,
    "max_tokens": 1024
  }'
```

### 流式响应
```bash
curl -X POST http://localhost:8002/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-api-key" \
  -d '{
    "model": "gpt-5",
    "messages": [{"role": "user","content": "Tell me a story"}],
    "stream": true
  }'
```

### Claude Messages
```bash
curl -X POST http://localhost:8002/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-api-key" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-3.5-sonnet",
    "messages": [{"role":"user","content":"Hello!"}],
    "max_tokens": 1024
  }'
```

## 前端使用指南
- 主题：支持明暗主题，自动应用全局 CSS 设计系统（style.css）。
- 页面：Dashboard（使用统计）、Login/Register、Chat、TokenManagement、AdminPanel 等。
- 构建：`npm run build:check`；生产直接 `npm run build`。
- 静态服务：后端 main.go 自动服务前端构建产物。

## 常见问题 FAQ
1) **刷新 404**：请通过前端入口访问，不要直接访问后端路由；生产环境使用前端构建产物。
2) **环境变量无效**：确认 `.env` 与启动目录一致，重启服务。
3) **Codex CLI 缺少 env**：`env_key` 需为合法变量名（无连字符），并在系统中设置同名环境变量。
4) **前端构建失败**：运行 `npm run build:check` 查看类型与打包错误；确保 Node 版本 ≥18。
5) **端口冲突**：修改 `.env` 的 `PORT` 或前端 Vite 端口 `vite.config.ts`。

## Cursor Session 说明

### 什么是 Cursor Session？
Cursor Session 是从 Cursor IDE 提取的认证令牌，本项目通过这些令牌调用 Cursor 的 AI 服务。每个 Session 对应一个 Cursor 账户，系统会自动轮询可用的 Session 来处理 API 请求。

### 如何获取 Cursor Session？
1. 登录 [Cursor IDE](https://cursor.com/) 或 Cursor 网页版
2. 打开浏览器开发者工具（F12）
3. 在 Application → Cookies 中找到认证相关的 Cookie
4. 在管理后台的「Cursor 会话」页面添加 Session

### 账户类型与可用模型

| 账户类型 | 月费 | 可用模型 |
|---------|------|---------|
| **Free** | $0 | `gpt-4o-mini`, `claude-3.5-haiku`, `cursor-small` 等等|
| **Pro** | $20 | Free 全部 + `gpt-4o`, `gpt-4.1`, `claude-3.5-sonnet`, `claude-3.7-sonnet`, `gemini-2.5-pro`, `gemini-2.5-flash` |
| **Pro+** | $40 | Pro 全部 + `gpt-5`, `gpt-5-mini`, `claude-4-sonnet`, `claude-4.5-sonnet`, `claude-4-opus`, `o3`, `o4-mini` |
| **Business** | $40/人 | Pro+ 全部 + 团队管理功能 |

> ⚠️ **注意**：
> - 模型列表可能随 Cursor 官方更新而变化
> - Free 账户有每日请求限制
> - Pro/Pro+ 账户有更高的配额但仍有速率限制
> - 建议添加多个 Session 以提高可用性和负载均衡

### Session 管理
- **自动轮询**：系统自动选择可用的 Session 处理请求
- **配额追踪**：记录每个 Session 的使用量
- **失效检测**：自动标记失效的 Session
- **手动刷新**：支持在管理后台手动检查 Session 状态

## 贡献指南
1. Fork & PR，遵循现有代码风格与 ESLint/TypeScript 约定。
2. 新增功能请附上相应测试（Go `go test`；前端 `vitest`）。
3. 提交前请执行：
   ```bash
   go test ./...
   cd frontend && npm run build:check
   ```

## 许可证
本项目采用 MIT 许可证，详见 [LICENSE](LICENSE)。
