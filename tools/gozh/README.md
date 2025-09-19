# GoZH - Go应用结构生成器

GoZH (Go Zhanhai) 是一个用于快速生成标准 Go 微服务应用结构的命令行工具。

## 功能特性

- 🚀 快速生成标准微服务结构
- 📁 自动创建完整的目录结构
- 🔧 支持自定义模块名
- 📝 自动生成基础代码模板
- 🎯 遵循最佳实践和架构模式

## 使用方法

### 基本用法

```bash
# 使用 Task 运行 (推荐)
task gozh:generate -- app/user/service

# 直接运行
cd tools/gozh
go run main.go app/user/service
```

### 高级用法

```bash
# 指定自定义模块名
go run main.go app/user/service -module=my-company.com/my-project

# 生成不同类型的服务
go run main.go app/user/api          # 用户API服务
go run main.go app/order/service     # 订单服务
go run main.go app/admin/dashboard   # 管理后台
go run main.go app/gateway/proxy     # API网关
```

## 生成的结构

每次运行 GoZH 都会生成如下标准结构：

```
app/your-service/
├── cmd/                    # 应用入口
│   └── main.go            # 主函数，包含依赖注入和启动逻辑
├── config/                # 配置管理
│   └── config.go          # 配置结构定义和加载
├── internal/              # 内部包，不对外暴露
│   ├── data/              # 数据层
│   │   └── provider.go    # 数据提供者，管理数据仓库
│   ├── handler/           # 业务逻辑层  
│   │   └── provider.go    # 处理器提供者，管理业务逻辑
│   └── server/            # 服务器层
│       ├── provider.go    # 服务器提供者
│       └── http/          # HTTP服务器实现
│           └── server.go  # HTTP路由和中间件
└── README.md              # 服务说明文档
```

## 代码模板

### 主函数模板 (cmd/main.go)

生成的主函数包含：
- 配置文件加载
- 依赖注入
- 服务启动
- 优雅关闭
- Swagger 文档注释

### 配置模板 (config/config.go)

支持的配置项：
- 服务器配置 (端口、模式)
- 数据库配置 (MySQL)
- 缓存配置 (Redis)
- 服务发现配置 (ETCD)

### 数据层模板 (internal/data/)

- 数据提供者模式
- 数据库连接管理
- 预留仓库接口

### 业务层模板 (internal/handler/)

- 处理器提供者模式
- 业务逻辑组织
- 预留处理器接口

### 服务器层模板 (internal/server/)

- HTTP服务器配置
- 路由管理
- 中间件支持
- CORS 配置
- 健康检查端点

## 命令行参数

| 参数 | 描述 | 示例 |
|------|------|------|
| `<应用路径>` | 要生成的应用路径 | `app/user/service` |
| `-module=<模块名>` | 自定义 Go 模块名 | `-module=github.com/my-org/my-project` |

## 环境要求

- Go 1.21+
- 项目根目录必须有 `go.mod` 文件

## 工作原理

1. **路径解析**: 从应用路径中提取应用名称和包信息
2. **模块检测**: 自动读取项目根目录的 `go.mod` 获取模块名
3. **模板渲染**: 使用 Go template 渲染所有代码文件
4. **文件生成**: 创建目录结构并写入生成的代码

## 模板变量

生成过程中使用的模板变量：

```go
type AppTemplate struct {
    AppName      string // 应用名称，如 "service", "api"
    PackageName  string // Go包名称，如 "service", "api"  
    ModulePath   string // Go模块路径，如 "github.com/my-org/my-project"
    AppPath      string // 应用路径，如 "app/user/service"
    ImportPrefix string // 导入路径前缀
}
```

## 自定义和扩展

### 修改模板

要自定义生成的代码模板，可以修改 `main.go` 中的模板常量：

- `cmdMainTemplate` - 主函数模板
- `configTemplate` - 配置文件模板
- `dataProviderTemplate` - 数据层模板
- `handlerProviderTemplate` - 业务层模板
- `serverProviderTemplate` - 服务器层模板
- `httpServerTemplate` - HTTP服务器模板
- `readmeTemplate` - README 模板

### 添加新模板

1. 在 `getAppTemplates()` 函数中添加新的文件模板
2. 定义对应的模板常量
3. 重新编译工具

## 最佳实践

### 命名约定

- 使用小写字母和连字符分隔
- 避免使用下划线
- 保持名称简洁明了

```bash
# 推荐
app/user/service
app/order/api
app/admin/dashboard

# 不推荐  
app/User/Service
app/order_management/api
```

### 目录结构

遵循标准的 Go 项目布局：
- `cmd/` - 应用程序入口
- `internal/` - 私有应用和库代码
- `config/` - 配置文件

### 依赖管理

生成的代码使用依赖注入模式：
- Provider 模式管理依赖关系
- 清晰的层次分离
- 易于测试和维护

## 故障排除

### 常见错误

**错误**: `未找到go.mod文件`
**解决**: 确保在 Go 项目根目录下运行，且存在 `go.mod` 文件

**错误**: `目录已存在`
**解决**: 工具会提示是否覆盖，输入 `y` 确认或 `n` 取消

**错误**: `模板渲染失败`
**解决**: 检查应用路径格式是否正确，避免特殊字符

### 调试技巧

1. 使用 `-v` 参数获取详细输出
2. 检查生成的 `go.mod` 文件模块名
3. 验证应用路径格式

## 更新日志

### v1.0.0
- 初始版本发布
- 支持基础微服务结构生成
- 支持自定义模块名
- 包含完整的代码模板

## 贡献

欢迎提交 Issue 和 Pull Request 来改进 GoZH 工具！