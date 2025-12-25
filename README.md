# 🍛 CurryAPI

<div align="center">

[English](#english) | [中文](#中文)

[![GitHub stars](https://img.shields.io/github/stars/StenCurry/CurryAPI?style=social)](https://github.com/StenCurry/CurryAPI/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/StenCurry/CurryAPI?style=social)](https://github.com/StenCurry/CurryAPI/network/members)
[![GitHub issues](https://img.shields.io/github/issues/StenCurry/CurryAPI)](https://github.com/StenCurry/CurryAPI/issues)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Transform Cursor Web into a fully OpenAI-compatible API Gateway**

</div>

---

<a name="english"></a>
## 🌍 English

### 📖 Introduction

CurryAPI is a self-hosted multi-model AI gateway that converts Cursor Web into a fully OpenAI-compatible API. It provides a unified interface for accessing various AI models including GPT, Claude, Gemini, and DeepSeek, with built-in user management, quota control, and usage analytics.

### ✨ Features

- 🔄 **OpenAI Compatible API** - Seamlessly integrate with ChatGPT-Next-Web, LobeChat, and other OpenAI-compatible applications
- 🤖 **Multi-Model Support** - Access 30+ models: GPT-4o, GPT-5, Claude 4, Gemini 2.5, DeepSeek, and more
- 👥 **User Management** - Complete registration, login, OAuth (Google/GitHub), email verification
- 📊 **Usage Analytics** - Real-time token consumption tracking and API call statistics
- 💰 **Quota Management** - Flexible quota allocation for multiple users
- 🔐 **API Key Management** - Generate and manage multiple API keys per user
- 🎮 **Fun Features** - Built-in mini-games to earn extra quota
- 🌓 **Modern UI** - Beautiful dashboard with dark/light theme support

### 🛠️ Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go 1.24+, Gin, MySQL |
| Frontend | Vue 3, TypeScript, Naive UI, Vite |
| Auth | JWT, OAuth 2.0, Session |
| Database | MySQL 8.0+ |

### 📦 Quick Start

#### Prerequisites
- Go 1.24+
- Node.js 18+
- MySQL 8.0+

#### 1. Clone the repository
```bash
git clone https://github.com/StenCurry/CurryAPI.git
cd CurryAPI
```

#### 2. Setup Database
```bash
mysql -u root -p -e "CREATE DATABASE cursor2api CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -u root -p cursor2api < database/schema.sql
```
> Default admin account: `admin` / `admin123` (please change after first login)

#### 3. Configure Environment
```bash
cp .env.example .env
# Edit .env with your database credentials and other settings
```

#### 4. Start Backend
```bash
go mod download
go run main.go
```

#### 5. Start Frontend (Development)
```bash
cd frontend
npm install
npm run dev
```

#### 6. Access
- Frontend: `http://localhost:5173`
- API: `http://localhost:8002`

### 🐳 Docker Deployment

```bash
# Build and run
docker build -t curryapi:latest .
docker run -d --name curryapi -p 8002:8002 --env-file .env curryapi:latest
```

### 📡 API Examples

#### OpenAI Chat Completions
```bash
curl -X POST http://localhost:8002/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-api-key" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello!"}],
    "stream": false
  }'
```

#### Claude Messages (Anthropic Format)
```bash
curl -X POST http://localhost:8002/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-api-key" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-3.5-sonnet",
    "messages": [{"role": "user", "content": "Hello!"}],
    "max_tokens": 1024
  }'
```

### 🎯 Supported Models

| Tier | Models |
|------|--------|
| Free | gpt-4o-mini, claude-3.5-haiku, cursor-small |
| Pro | gpt-4o, gpt-4.1, claude-3.5-sonnet, claude-3.7-sonnet, gemini-2.5-pro |
| Pro+ | gpt-5, claude-4-sonnet, claude-4.5-sonnet, claude-4-opus, o3, o4-mini |

### 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

### 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

---

<a name="中文"></a>
## 🇨🇳 中文

### 📖 项目介绍

CurryAPI 是一个可自部署的多模型 AI 网关，能够将 Cursor Web 转换为完全兼容 OpenAI API 格式的接口。它提供统一的接口来访问各种 AI 模型，包括 GPT、Claude、Gemini 和 DeepSeek，并内置用户管理、配额控制和使用统计功能。

### ✨ 功能特性

- 🔄 **OpenAI 兼容 API** - 无缝对接 ChatGPT-Next-Web、LobeChat 等 OpenAI 兼容应用
- 🤖 **多模型支持** - 支持 30+ 模型：GPT-4o、GPT-5、Claude 4、Gemini 2.5、DeepSeek 等
- 👥 **用户管理** - 完整的注册登录、OAuth（Google/GitHub）、邮箱验证
- 📊 **使用统计** - 实时追踪 Token 消耗和 API 调用统计
- 💰 **配额管理** - 灵活的多用户配额分配
- 🔐 **API Key 管理** - 每个用户可生成和管理多个 API Key
- 🎮 **趣味功能** - 内置小游戏赚取额外配额
- 🌓 **现代化 UI** - 精美的仪表盘，支持明暗主题切换

### 🛠️ 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go 1.24+, Gin, MySQL |
| 前端 | Vue 3, TypeScript, Naive UI, Vite |
| 认证 | JWT, OAuth 2.0, Session |
| 数据库 | MySQL 8.0+ |

### 📦 快速开始

#### 环境要求
- Go 1.24+
- Node.js 18+
- MySQL 8.0+

#### 1. 克隆仓库
```bash
git clone https://github.com/StenCurry/CurryAPI.git
cd CurryAPI
```

#### 2. 准备数据库
```bash
mysql -u root -p -e "CREATE DATABASE cursor2api CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -u root -p cursor2api < database/schema.sql
```
> 默认管理员账户：`admin` / `admin123`（请首次登录后修改密码）

#### 3. 配置环境变量
```bash
cp .env.example .env
# 编辑 .env 文件，配置数据库连接等信息
```

#### 4. 启动后端
```bash
go mod download
go run main.go
```

#### 5. 启动前端（开发模式）
```bash
cd frontend
npm install
npm run dev
```

#### 6. 访问
- 前端：`http://localhost:5173`
- API：`http://localhost:8002`

### 🐳 Docker 部署

```bash
# 构建并运行
docker build -t curryapi:latest .
docker run -d --name curryapi -p 8002:8002 --env-file .env curryapi:latest
```

### 📡 API 示例

#### OpenAI Chat Completions
```bash
curl -X POST http://localhost:8002/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-api-key" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "你好！"}],
    "stream": false
  }'
```

#### Claude Messages（Anthropic 格式）
```bash
curl -X POST http://localhost:8002/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-api-key" \
  -H "anthropic-version: 2023-06-01" \
  -d '{
    "model": "claude-3.5-sonnet",
    "messages": [{"role": "user", "content": "你好！"}],
    "max_tokens": 1024
  }'
```

### 🎯 支持的模型

| 等级 | 模型 |
|------|------|
| Free | gpt-4o-mini, claude-3.5-haiku, cursor-small |
| Pro | gpt-4o, gpt-4.1, claude-3.5-sonnet, claude-3.7-sonnet, gemini-2.5-pro |
| Pro+ | gpt-5, claude-4-sonnet, claude-4.5-sonnet, claude-4-opus, o3, o4-mini |

### 🔧 Cursor Session 说明

#### 什么是 Cursor Session？
Cursor Session 是从 Cursor IDE 提取的认证令牌，本项目通过这些令牌调用 Cursor 的 AI 服务。

#### 如何获取？
1. 登录 [Cursor IDE](https://cursor.com/) 或 Cursor 网页版
2. 打开浏览器开发者工具（F12）
3. 在 Application → Cookies 中找到认证相关的 Cookie
4. 在管理后台的「Cursor 会话」页面添加 Session

### 📁 项目结构
```
CurryAPI/
├── main.go                 # 后端入口
├── config/                 # 配置
├── handlers/               # HTTP 处理器
├── middleware/             # 中间件
├── services/               # 业务逻辑
├── database/               # 数据访问层
│   └── schema.sql          # 数据库初始化脚本
├── models/                 # 数据模型
├── utils/                  # 工具函数
├── frontend/               # 前端项目 (Vue + Vite)
│   ├── src/
│   │   ├── views/          # 页面组件
│   │   ├── components/     # 通用组件
│   │   └── api/            # API 调用
│   └── package.json
├── .env.example            # 环境变量示例
└── README.md
```

### ❓ 常见问题

**Q: 刷新页面出现 404？**
A: 请通过前端入口访问，生产环境需要配置 Nginx 的 try_files。

**Q: 如何添加更多 Session？**
A: 在管理后台的「Cursor 会话」页面添加，建议添加多个以提高可用性。

**Q: 支持哪些 OAuth 登录？**
A: 目前支持 Google 和 GitHub OAuth 登录。

### 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

### 🤝 贡献

欢迎贡献代码！请随时提交 Pull Request。

### ⭐ Star History

如果这个项目对你有帮助，请给一个 Star ⭐ 支持一下！

---

<div align="center">
Made with ❤️ by <a href="https://github.com/StenCurry">StenCurry</a>
</div>
