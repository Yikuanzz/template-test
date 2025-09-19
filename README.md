# Go 微服务模板 (Go Microservice Template)

一个现代化的 Go 微服务项目模板，集成了常用的中间件和最佳实践，可以快速构建高质量的 Go 应用程序。

## 📋 目录

- [特性](#特性)
- [项目结构](#项目结构)
- [环境要求](#环境要求)
- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [开发指南](#开发指南)
- [工具使用](#工具使用)
- [部署](#部署)

## 🎯 特性

- **模块化架构**: 清晰的分层架构，易于维护和扩展
- **依赖注入**: 使用 Provider 模式管理依赖关系
- **中间件支持**: 集成 MySQL、Redis、ETCD 等常用中间件
- **统一响应**: 标准化的 API 响应格式和错误处理
- **代码生成**: 内置应用结构生成工具
- **数据库迁移**: 完整的数据库版本管理工具
- **监控观测**: 集成 Prometheus 指标监控和 Grafana 可视化
- **链路追踪**: 集成 Jaeger 分布式链路追踪
- **代码质量**: 集成 golangci-lint 和 Git hooks
- **容器化**: Docker Compose 开箱即用

## 📁 项目结构

```
go-template/
├── app/                     # 应用程序目录（通过工具生成）
├── db/                      # 数据库相关
│   ├── migrations/          # 数据库迁移文件
│   └── model/               # 数据模型
├── pkg/                     # 公共包
│   ├── etcd/               # ETCD 连接
│   ├── helper/             # 日志辅助工具
│   ├── jaeger/             # Jaeger 链路追踪
│   ├── mysql/              # MySQL 连接
│   ├── prometheus/         # Prometheus 监控
│   └── redis/              # Redis 连接
├── utils/                   # 工具函数
│   └── common/             # 通用响应和错误代码
├── tools/                   # 开发工具
│   ├── gozh/               # Go 应用结构生成器
│   └── migrator/           # 数据库迁移工具
├── docs/                    # 文档和模板
├── monitoring/              # 监控配置
│   ├── prometheus/         # Prometheus 配置
│   └── grafana/            # Grafana 配置和仪表板
├── docker-compose.yaml     # Docker 服务编排
├── Taskfile.yml            # 任务运行器配置
└── README.md               # 项目说明
```

## 🔧 环境要求

### 必需软件

- **Go**: 1.21+ 
- **Docker**: 20.10 或更高版本
- **Docker Compose**: 2.0 或更高版本

### 可选工具

- **Task**: 用于运行项目任务
- **golangci-lint**: 代码质量检查
- **Node.js**: Git hooks 支持

## 🚀 快速开始

### 1. 克隆模板

```bash
git clone <repository-url>
cd go-template
```

### 2. 启动基础服务

```bash
# 启动所有服务：MySQL、Redis、ETCD、Prometheus、Grafana、Jaeger
docker compose up -d

# 或仅启动基础中间件
docker compose up -d mysql redis etcd
```

### 3. 生成应用程序

```bash
# 生成一个新的微服务应用
task gozh:generate -- app/user/service

# 或者直接使用工具
cd tools/gozh
go run main.go app/user/service
```

### 4. 运行数据库迁移

```bash
# 执行数据库迁移
task migrator:up
```

### 5. 启动应用

```bash
# 进入生成的应用目录
cd app/user/service

# 启动服务
go run cmd/main.go
```

### 6. 访问监控界面

启动后可以访问以下监控界面：

- **应用服务**: http://localhost:8080
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Jaeger**: http://localhost:16686
- **应用指标**: http://localhost:8080/metrics

## ⚙️ 配置说明

### 默认配置

项目使用 TOML 格式的配置文件：

```toml
[server]
port = "8080"
mode = "debug"

[database]
host = "localhost"
port = "3406"
user = "root"
password = "root123"
dbname = "your_database"

[redis]
host = "localhost"
port = "6379"
password = ""
db = 0

[etcd]
endpoints = ["localhost:2379"]
dial_timeout = 5
```

### 环境变量

支持通过环境变量覆盖配置：

```bash
export DB_HOST="localhost"
export DB_PORT="3306"
export REDIS_HOST="localhost"
export REDIS_PORT="6379"
```

## 🛠️ 开发指南

### 应用架构

每个生成的应用都遵循以下架构：

```
app/your-service/
├── cmd/                    # 应用入口
│   └── main.go
├── config/                 # 配置管理
│   └── config.go
├── internal/               # 内部包
│   ├── data/              # 数据层
│   ├── handler/           # 业务逻辑层
│   └── server/            # 服务器层
└── README.md              # 应用文档
```

### 添加新功能

1. **数据层**: 在 `internal/data/` 添加数据仓库
2. **业务层**: 在 `internal/handler/` 添加处理器
3. **路由层**: 在 `internal/server/http/` 添加路由

### 响应格式

使用统一的响应格式：

```go
// 成功响应
common.Success(c, data)

// 业务错误响应
common.BusinessResponse(c, common.CodeNotFound, nil)

// 自定义消息响应
common.BusinessResponseWithMessage(c, common.CodeSuccess, "操作成功", data)
```

## 🔨 工具使用

### GoZH 应用生成器

快速生成新的微服务应用：

```bash
# 生成标准微服务
task gozh:generate -- app/user/service

# 生成管理后台
task gozh:generate -- app/admin/backend

# 指定自定义模块名
cd tools/gozh
go run main.go app/custom/service -module=custom-module
```

### 数据库迁移工具

管理数据库版本：

```bash
# 创建迁移文件
task migrator:create -- create_users_table

# 执行迁移
task migrator:up

# 回滚迁移
task migrator:down

# 查看版本
task migrator:version

# 跳转到指定版本
task migrator:goto -- 20240101120000
```

### 代码质量检查

```bash
# 运行代码检查
task golangci:lint

# 自动格式化
task golangci:fmt

# 运行所有检查
golangci-lint run --fix
```

## 🚢 部署

### Docker 部署

```bash
# 构建镜像
docker build -t your-service .

# 运行容器
docker run -p 8080:8080 your-service
```

### Docker Compose 部署

```bash
# 启动所有服务
docker compose up -d

# 查看日志
docker compose logs -f

# 停止服务
docker compose down
```

## 📚 API 文档

生成的应用支持 Swagger 文档：

```bash
# 生成 Swagger 文档（需要在应用目录中）
swag init -g cmd/main.go

# 访问文档
# http://localhost:8080/swagger/index.html
```

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 打开 Pull Request

### 提交信息规范

使用 Conventional Commits 规范：

- `feat`: 新功能
- `fix`: 错误修复
- `docs`: 文档更改
- `style`: 代码格式（不影响代码运行的变动）
- `refactor`: 重构
- `test`: 添加测试
- `chore`: 构建过程或辅助工具的变动

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🆘 故障排除

### 常见问题

1. **端口冲突**: 检查以下端口是否被占用
   - 3406 (MySQL)
   - 6379 (Redis) 
   - 2379 (ETCD)
   - 9090 (Prometheus)
   - 3000 (Grafana)
   - 16686 (Jaeger)
2. **Docker 问题**: 确保 Docker 服务正在运行
3. **权限问题**: 确保有执行工具的权限
4. **监控数据**: 启动应用后等待几分钟再查看 Grafana 仪表板

### 获取帮助

- 查看 [文档](./docs/)
- 提交 [Issue](../../issues)
- 联系维护者

---

**Happy Coding! 🎉**