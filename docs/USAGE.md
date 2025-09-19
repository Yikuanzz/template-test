# 使用指南

本文档详细介绍如何使用这个Go微服务模板。

## 快速开始

### 1. 环境准备

确保您的开发环境已安装：

- Go 1.21+
- Docker & Docker Compose
- Task (可选，推荐)

### 2. 克隆模板

```bash
git clone <your-template-repo>
cd go-template

# 重命名为您的项目名
```

### 3. 初始化项目

```bash
# 修改 go.mod 中的模块名
# 将 github.com/your-org/go-template 改为您的项目名

# 启动基础服务
docker compose up -d

# 检查服务状态
docker compose ps
```

### 4. 生成第一个微服务

```bash
# 生成用户服务
task gozh:generate -- app/user/service

# 生成订单服务  
task gozh:generate -- app/order/service

# 生成管理后台
task gozh:generate -- app/admin/backend
```

### 5. 配置数据库

```bash
# 复制配置文件
cp config.toml.example config.toml

# 执行数据库迁移
task migrator:up

# 查看迁移状态
task migrator:version
```

### 6. 启动服务

```bash
# 进入生成的服务目录
cd app/user/service

# 启动服务
go run cmd/main.go

# 或使用配置文件
go run cmd/main.go -config=config.toml
```

## 深入使用

### 应用结构生成器 (GoZH)

GoZH 是内置的应用结构生成器，可以快速创建标准的微服务结构。

#### 基本用法

```bash
# 生成标准微服务
task gozh:generate -- app/user/service

# 生成 API 网关
task gozh:generate -- app/gateway/api

# 生成后台管理
task gozh:generate -- app/admin/dashboard
```

#### 自定义模块名

```bash
cd tools/gozh
go run main.go app/custom/service -module=my-company.com/my-project
```

#### 生成的结构

每个生成的应用都包含：

```
app/your-service/
├── cmd/main.go           # 应用入口，包含依赖注入
├── config/config.go      # 配置管理
├── internal/
│   ├── data/            # 数据层，数据库操作
│   ├── handler/         # 业务逻辑层
│   └── server/          # 服务器层，HTTP路由
└── README.md            # 服务说明文档
```

### 数据库迁移工具

内置的迁移工具提供完整的数据库版本管理。

#### 创建迁移

```bash
# 创建新的迁移文件
task migrator:create -- create_posts_table

# 这会生成两个文件：
# - 20240315120000_create_posts_table.up.sql   (正向迁移)
# - 20240315120000_create_posts_table.down.sql (回滚迁移)
```

#### 执行迁移

```bash
# 执行所有待执行的迁移
task migrator:up

# 执行指定步数的迁移
task migrator:up:steps -- 2

# 回滚最后一次迁移
task migrator:down

# 回滚指定步数
task migrator:down:steps -- 2
```

#### 版本管理

```bash
# 查看当前版本和迁移历史
task migrator:version

# 跳转到指定版本
task migrator:goto -- 20240315120000
```

#### 数据导入导出

```bash
# 导入 SQL 文件
task migrator:import -- data.sql

# 导入 CSV 文件
task migrator:import -- users.csv

# 导出为 SQL
task migrator:export:sql -- backup.sql

# 导出为 CSV
task migrator:export:csv -- ./export/
```

### 开发最佳实践

#### 1. 添加新的API端点

以用户服务为例，添加用户管理功能：

1. **定义数据模型** (在 `db/model/model.go`)

```go
type User struct {
    BaseModel
    Username string `gorm:"uniqueIndex;size:50" json:"username"`
    Email    string `gorm:"uniqueIndex;size:100" json:"email"`
    Status   int    `gorm:"default:1" json:"status"`
}
```

2. **创建数据仓库** (在 `internal/data/user/`)

```go
// internal/data/user/repo.go
type UserRepo struct {
    db  *gorm.DB
    log *zhlog.Helper
}

func (r *UserRepo) Create(user *model.User) error {
    return r.db.Create(user).Error
}

func (r *UserRepo) GetByID(id uint) (*model.User, error) {
    var user model.User
    err := r.db.First(&user, id).Error
    return &user, err
}
```

3. **添加业务逻辑** (在 `internal/handler/user/`)

```go
// internal/handler/user/handler.go
type UserHandler struct {
    userRepo *user.UserRepo
    log      *zhlog.Helper
}

func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        common.BadRequest(c, "参数错误", err.Error())
        return
    }
    
    // 业务逻辑处理
    user := &model.User{
        Username: req.Username,
        Email:    req.Email,
    }
    
    if err := h.userRepo.Create(user); err != nil {
        common.InternalError(c, "创建用户失败", err.Error())
        return
    }
    
    common.Success(c, user)
}
```

4. **添加路由** (在 `internal/server/http/server.go`)

```go
func (s *HTTPServer) SetupRoutes() {
    api := s.router.Group("/api/v1")
    
    // 用户相关路由
    userGroup := api.Group("/users")
    {
        userHandler := s.handler.ProvideUserHandler()
        userGroup.POST("", userHandler.CreateUser)
        userGroup.GET("/:id", userHandler.GetUser)
        userGroup.PUT("/:id", userHandler.UpdateUser)
        userGroup.DELETE("/:id", userHandler.DeleteUser)
    }
}
```

#### 2. 错误处理

使用统一的错误响应格式：

```go
// 成功响应
common.Success(c, data)

// 业务错误
common.BusinessResponse(c, common.CodeUserNotFound, nil)

// 自定义消息
common.BusinessResponseWithMessage(c, common.CodeSuccess, "用户创建成功", user)

// 传统方式 (向后兼容)
common.BadRequest(c, "参数错误", nil)
common.NotFound(c, "用户不存在", nil)
```

#### 3. 配置管理

支持多种配置方式：

```go
// 1. 使用配置文件
cfg, err := config.LoadConfig("config.toml")

// 2. 使用默认配置
cfg := config.DefaultConfig()

// 3. 环境变量覆盖
os.Setenv("DB_HOST", "localhost")
os.Setenv("DB_PORT", "3306")
```

#### 4. 日志记录

使用统一的日志接口：

```go
// 创建日志实例
logger := helper.NewSimpleLogger()

// 记录日志
logger.Info("用户创建成功", "user_id", user.ID)
logger.Error("数据库连接失败", "error", err)
logger.Debug("调试信息", "data", debugData)
```

### 部署

#### Docker 部署

```bash
# 构建镜像
docker build -t your-service:latest .

# 运行容器
docker run -p 8080:8080 your-service:latest
```

#### 使用 Docker Compose

```bash
# 启动所有服务
docker compose up -d

# 查看日志
docker compose logs -f your-service

# 停止服务
docker compose down
```

### 测试

#### 单元测试

使用提供的测试模板：

- [数据层单元测试模板](template/data层单元测试模板.md)
- [处理器层单元测试模板](template/handler层单元测试模板.md)

#### API 测试

```bash
# 健康检查
curl http://localhost:8080/api/v1/health

# 创建用户
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"username":"john","email":"john@example.com"}'
```

## 常见问题

### Q: 如何修改默认端口？

A: 在配置文件中修改 `server.port` 或设置环境变量 `PORT`。

### Q: 如何添加新的中间件？

A: 在 `pkg/` 目录下添加新的中间件包，参考现有的 `mysql`、`redis` 包的实现。

### Q: 如何自定义响应格式？

A: 修改 `utils/common/response.go` 和 `utils/common/code.go` 文件。

### Q: 数据库迁移失败怎么办？

A: 检查数据库连接配置，确保 MySQL 服务正在运行，使用 `task migrator:version` 查看当前状态。

## 更多资源

- [API 设计规范](./api-design.md) (TODO)
- [测试策略](./testing-strategy.md) (TODO)
- [性能优化](./performance.md) (TODO)
- [故障排除](./troubleshooting.md) (TODO)
