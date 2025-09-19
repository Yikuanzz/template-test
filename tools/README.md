# 开发工具

本目录包含项目开发过程中使用的各种工具。

## 工具列表

### 1. GoZH - Go应用结构生成器

**位置**: `tools/gozh/`

**功能**: 快速生成标准的Go微服务应用结构。

**使用方法**:
```bash
# 使用 Task (推荐)
task gozh:generate -- app/user/service

# 直接运行
cd tools/gozh
go run main.go app/user/service

# 自定义模块名
go run main.go app/user/service -module=my-company.com/my-project
```

**生成结构**:
```
app/your-service/
├── cmd/main.go           # 应用入口
├── config/config.go      # 配置管理
├── internal/
│   ├── data/            # 数据层
│   ├── handler/         # 业务逻辑层
│   └── server/          # 服务器层
└── README.md            # 服务文档
```

### 2. Migrator - 数据库迁移工具

**位置**: `tools/migrator/`

**功能**: 提供完整的数据库版本管理功能。

**主要命令**:
```bash
# 创建迁移文件
task migrator:create -- create_users_table

# 执行迁移
task migrator:up

# 回滚迁移
task migrator:down

# 查看版本
task migrator:version

# 跳转版本
task migrator:goto -- 20240315120000
```

**支持功能**:
- ✅ 创建迁移文件
- ✅ 正向/反向迁移
- ✅ 版本跳转
- ✅ 迁移历史查看
- ✅ SQL/CSV 数据导入导出
- ✅ 指定步数迁移

## 配置说明

### GoZH 配置

GoZH 工具使用模板生成代码，支持以下参数：

- `AppName`: 应用名称 (从路径自动提取)
- `PackageName`: Go包名称 (自动转换)
- `ModulePath`: Go模块路径 (从go.mod读取)
- `AppPath`: 应用路径
- `ImportPrefix`: 导入路径前缀

### Migrator 配置

Migrator 工具的默认配置：

```go
const (
    DB_HOST     = "localhost"
    DB_PORT     = "3406"
    DB_USER     = "root"
    DB_PASSWORD = "root123"
    DB_NAME     = "go_template_db"
    MIGRATIONS_DIR = "../../db/migrations"
)
```

可以通过环境变量覆盖：
```bash
export DB_HOST="your-host"
export DB_PORT="3306"
export DB_USER="your-user"
export DB_PASSWORD="your-password"
export DB_NAME="your-database"
```

## 工具开发

### 添加新工具

1. 在 `tools/` 下创建新目录
2. 实现工具逻辑
3. 在 `tools/README.md` 中添加说明
4. 在根目录 `Taskfile.yml` 中添加任务

### 工具规范

- 使用 Go 语言实现
- 提供清晰的帮助信息
- 支持命令行参数
- 错误处理要完善
- 提供使用示例

## 贡献指南

欢迎贡献新的开发工具！请确保：

1. 工具有明确的用途和价值
2. 代码质量良好，有适当的错误处理
3. 提供完整的文档和使用示例
4. 在 Taskfile.yml 中添加对应的任务
5. 更新本 README 文件

## 故障排除

### GoZH 常见问题

**Q: 生成的代码导入路径错误**
A: 检查项目根目录的 go.mod 文件，确保模块名正确。

**Q: 生成的目录已存在**
A: 工具会提示是否覆盖，输入 `y` 确认覆盖或 `n` 取消。

### Migrator 常见问题

**Q: 连接数据库失败**
A: 检查数据库服务是否启动，配置信息是否正确。

**Q: 迁移文件不存在**
A: 确保在正确的目录下运行，迁移文件应在 `db/migrations/` 目录中。

**Q: 版本跳转失败**
A: 检查目标版本是否存在对应的迁移文件。
