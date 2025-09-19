package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// 应用模板数据
type AppTemplate struct {
	AppName      string // 应用名称，如 "admin", "client"
	PackageName  string // Go包名称
	ModulePath   string // Go模块路径
	AppPath      string // 应用路径，如 "app/ticketing/admin"
	ImportPrefix string // 导入路径前缀
}

// 文件模板定义
type FileTemplate struct {
	Path    string // 文件相对路径
	Content string // 文件内容模板
	IsDir   bool   // 是否为目录
}

// 获取应用模板
func getAppTemplates() []FileTemplate {
	return []FileTemplate{
		// 目录结构
		{Path: "cmd", IsDir: true},
		{Path: "config", IsDir: true},
		{Path: "internal", IsDir: true},
		{Path: "internal/data", IsDir: true},
		{Path: "internal/handler", IsDir: true},
		{Path: "internal/server", IsDir: true},
		{Path: "internal/server/http", IsDir: true},

		// cmd/main.go
		{Path: "cmd/main.go", Content: cmdMainTemplate},

		// config/config.go
		{Path: "config/config.go", Content: configTemplate},

		// internal/data/provider.go
		{Path: "internal/data/provider.go", Content: dataProviderTemplate},

		// internal/handler/provider.go
		{Path: "internal/handler/provider.go", Content: handlerProviderTemplate},

		// internal/server/provider.go
		{Path: "internal/server/provider.go", Content: serverProviderTemplate},

		// internal/server/http/server.go
		{Path: "internal/server/http/server.go", Content: httpServerTemplate},

		// README.md
		{Path: "README.md", Content: readmeTemplate},
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: gozh <应用路径> [选项]")
		fmt.Println("示例: gozh app/ticketing/admin")
		fmt.Println("      gozh app/ticketing/order")
		fmt.Println("      gozh app/user/service -module=custom-module")
		fmt.Println("")
		fmt.Println("选项:")
		fmt.Println("  -module=<模块名>  手动指定Go模块名（覆盖自动检测）")
		os.Exit(1)
	}

	appPath := strings.TrimSuffix(os.Args[1], "/")
	var customModule string

	// 解析命令行参数
	for i := 2; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "-module=") {
			customModule = strings.TrimPrefix(arg, "-module=")
		}
	}

	// 获取项目信息
	projectInfo, err := getProjectInfo()
	if err != nil {
		log.Fatalf("获取项目信息失败: %v", err)
	}

	// 如果指定了自定义模块名，则使用它
	if customModule != "" {
		projectInfo.ModulePath = customModule
	}

	// 解析应用信息
	appTemplate, err := parseAppPath(appPath, projectInfo)
	if err != nil {
		log.Fatalf("解析应用路径失败: %v", err)
	}

	// 构建完整的目标路径（相对于项目根目录）
	targetPath := filepath.Join(projectInfo.RootDir, appPath)

	// 检查目标目录是否存在
	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		fmt.Printf("目录 %s 已存在，是否覆盖? (y/N): ", targetPath)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		response := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if response != "y" && response != "yes" {
			fmt.Println("操作已取消")
			return
		}
	}

	// 生成应用结构
	err = generateApp(targetPath, appTemplate)
	if err != nil {
		log.Fatalf("生成应用失败: %v", err)
	}

	fmt.Printf("✅ 成功生成应用: %s\n", appPath)
	fmt.Printf("📁 项目根目录: %s\n", projectInfo.RootDir)
	fmt.Printf("📦 模块名称: %s\n", projectInfo.ModulePath)
	fmt.Printf("📁 包含以下结构:\n")
	fmt.Printf("   ├── cmd/main.go          (应用入口)\n")
	fmt.Printf("   ├── config/config.go     (配置管理)\n")
	fmt.Printf("   ├── internal/data/       (数据层)\n")
	fmt.Printf("   ├── internal/handler/    (处理器层)\n")
	fmt.Printf("   ├── internal/server/     (服务器层)\n")
	fmt.Printf("   └── README.md            (说明文档)\n")
}

// 解析应用路径，提取应用信息
func parseAppPath(appPath string, projectInfo *ProjectInfo) (*AppTemplate, error) {
	// 从路径中提取应用名称
	parts := strings.Split(appPath, "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("无效的应用路径")
	}

	appName := parts[len(parts)-1]
	packageName := strings.ReplaceAll(appName, "-", "_")
	importPrefix := projectInfo.ModulePath + "/" + appPath

	return &AppTemplate{
		AppName:      appName,
		PackageName:  packageName,
		ModulePath:   projectInfo.ModulePath,
		AppPath:      appPath,
		ImportPrefix: importPrefix,
	}, nil
}

// ProjectInfo 项目信息
type ProjectInfo struct {
	RootDir    string // 项目根目录
	ModulePath string // Go模块路径
}

// 获取项目信息（根目录和模块路径）
func getProjectInfo() (*ProjectInfo, error) {
	// 查找go.mod文件
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// 读取go.mod文件第一行获取模块名
			file, err := os.Open(goModPath)
			if err != nil {
				return nil, err
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			if scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "module ") {
					modulePath := strings.TrimSpace(strings.TrimPrefix(line, "module"))
					return &ProjectInfo{
						RootDir:    currentDir,
						ModulePath: modulePath,
					}, nil
				}
			}
			return nil, fmt.Errorf("无法解析go.mod文件")
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			break
		}
		currentDir = parent
	}

	return nil, fmt.Errorf("未找到go.mod文件")
}

// 生成应用结构
func generateApp(appPath string, appTemplate *AppTemplate) error {
	templates := getAppTemplates()

	for _, tmpl := range templates {
		targetPath := filepath.Join(appPath, tmpl.Path)

		if tmpl.IsDir {
			// 创建目录
			err := os.MkdirAll(targetPath, 0755)
			if err != nil {
				return fmt.Errorf("创建目录 %s 失败: %v", targetPath, err)
			}
			continue
		}

		// 创建父目录
		dir := filepath.Dir(targetPath)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("创建目录 %s 失败: %v", dir, err)
		}

		// 渲染模板内容
		content, err := renderTemplate(tmpl.Content, appTemplate)
		if err != nil {
			return fmt.Errorf("渲染模板 %s 失败: %v", tmpl.Path, err)
		}

		// 写入文件
		err = os.WriteFile(targetPath, []byte(content), 0644)
		if err != nil {
			return fmt.Errorf("写入文件 %s 失败: %v", targetPath, err)
		}
	}

	return nil
}

// 渲染模板
func renderTemplate(templateContent string, data *AppTemplate) (string, error) {
	tmpl, err := template.New("template").Parse(templateContent)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ======================== 模板定义 ========================

// cmd/main.go 模板
const cmdMainTemplate = `package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"{{.ImportPrefix}}/config"
	"{{.ImportPrefix}}/internal/data"
	"{{.ImportPrefix}}/internal/handler"
	"{{.ImportPrefix}}/internal/server"
	"{{.ModulePath}}/pkg/helper"
	"{{.ModulePath}}/pkg/jaeger"
	"{{.ModulePath}}/pkg/mysql"
	"{{.ModulePath}}/pkg/prometheus"
	"{{.ModulePath}}/pkg/redis"
)

// @title {{.AppName}}服务API
// @version 1.0
// @description {{.AppName}}服务API文档
//
// @termsOfService http://swagger.io/terms/
//
// @contact.name {{.AppName}}开发团队
// @contact.url https://github.com/your-org/ticketing-system
// @contact.email support@yourcompany.com
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8080
// @BasePath /api/v1

func main() {
	// 解析命令行参数
	configFile := flag.String("config", "config.toml", "配置文件路径")
	flag.Parse()

	// 加载配置
	var cfg *config.Config

	if _, err := os.Stat(*configFile); os.IsNotExist(err) {
		cfg = config.DefaultConfig()
	} else {
		cfg, err = config.LoadConfig(*configFile)
		if err != nil {
			cfg = config.DefaultConfig()
		}
	}

	// 创建应用程序实例
	app, cleanup, err := initApp(cfg)
	if err != nil {
		panic(fmt.Sprintf("初始化应用程序失败: %v", err))
	}

	// 设置优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// goroutine 中启动服务器
	go func() {
		if err := app.Run(); err != nil {
			panic(fmt.Sprintf("服务器启动失败: %v", err))
		}
	}()

	// 等待关闭信号
	<-quit
	fmt.Println("正在关闭服务...")

	// 清理资源
	cleanup()
	fmt.Println("服务已关闭")
}

// App 应用程序结构
type App struct {
	Config          *config.Config
	DataProvider    *data.DataProvider
	HandlerProvider *handler.HandlerProvider
	ServerProvider  *server.ServerProvider
	Metrics         *prometheus.Metrics
	Tracer          *jaeger.TracingProvider
}

// initApp 初始化应用程序
func initApp(cfg *config.Config) (*App, func(), error) {
	// 创建日志记录器
	logger := helper.NewSimpleLogger()

	// 创建数据库连接
	db := mysql.NewMySQL(&mysql.MySQLConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
	}, logger)

	// 创建Redis连接
	rdb := redis.NewRedis(&redis.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}, logger)

	// 创建Prometheus监控
	metrics := prometheus.NewMetrics(prometheus.DefaultConfig("{{.AppName}}"), logger)

	// 创建Jaeger链路追踪
	tracer, err := jaeger.NewTracingProvider(jaeger.DefaultConfig("{{.AppName}}-service"), logger)
	if err != nil {
		logger.Error("Failed to initialize tracing", "error", err)
		// 继续运行，但不使用追踪
	}

	// 创建数据提供者
	dataProvider := data.NewDataProvider(db, rdb, logger)

	// 创建处理器提供者
	handlerProvider := handler.NewHandlerProvider(dataProvider, logger)

	// 创建服务器提供者
	serverProvider := server.NewServerProvider(handlerProvider, logger, metrics, tracer)

	app := &App{
		Config:          cfg,
		DataProvider:    dataProvider,
		HandlerProvider: handlerProvider,
		ServerProvider:  serverProvider,
		Metrics:         metrics,
		Tracer:          tracer,
	}

	// 返回清理函数
	cleanup := func() {
		if db != nil {
			if sqlDB, err := db.DB(); err == nil {
				if err := sqlDB.Close(); err != nil {
					logger.Error("关闭数据库连接失败", "error", err)
				}
			}
		}

		if rdb != nil {
			if err := rdb.Close(); err != nil {
				logger.Error("关闭Redis连接失败", "error", err)
			}
		}

		if tracer != nil {
			if err := tracer.Shutdown(context.Background()); err != nil {
				logger.Error("关闭链路追踪失败", "error", err)
			}
		}

		logger.Info("应用程序资源已清理")
	}

	return app, cleanup, nil
}

// Run 启动应用程序
func (app *App) Run() error {
	// 启动HTTP服务器
	return app.ServerProvider.HTTPServer.Start(app.Config.Server.Port)
}
`

// config/config.go 模板
const configTemplate = `package config

import (
	"log"

	"github.com/BurntSushi/toml"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   ` + "`toml:\"server\"`" + `
	Database DatabaseConfig ` + "`toml:\"database\"`" + `
	Redis    RedisConfig    ` + "`toml:\"redis\"`" + `
	Etcd     EtcdConfig     ` + "`toml:\"etcd\"`" + `
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string ` + "`toml:\"port\"`" + `
	Mode string ` + "`toml:\"mode\"`" + `
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string ` + "`toml:\"host\"`" + `
	Port     string ` + "`toml:\"port\"`" + `
	User     string ` + "`toml:\"user\"`" + `
	Password string ` + "`toml:\"password\"`" + `
	DBName   string ` + "`toml:\"dbname\"`" + `
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string ` + "`toml:\"host\"`" + `
	Port     string ` + "`toml:\"port\"`" + `
	Password string ` + "`toml:\"password\"`" + `
	DB       int    ` + "`toml:\"db\"`" + `
}

// EtcdConfig Etcd配置
type EtcdConfig struct {
	Endpoints   []string ` + "`toml:\"endpoints\"`" + `
	DialTimeout int      ` + "`toml:\"dial_timeout\"`" + `
}

// LoadConfig 加载配置文件
func LoadConfig(filename string) (*Config, error) {
	var config Config

	if _, err := toml.DecodeFile(filename, &config); err != nil {
		log.Printf("加载配置文件失败: %v", err)
		return nil, err
	}

	return &config, nil
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: "8080",
			Mode: "debug",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "3306",
			User:     "root",
			Password: "root123",
			DBName:   "{{.PackageName}}_db",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
			DB:       0,
		},
		Etcd: EtcdConfig{
			Endpoints:   []string{"localhost:2379"},
			DialTimeout: 5,
		},
	}
}
`

// internal/data/provider.go 模板
const dataProviderTemplate = `package data

import (
	zhlog "codeup.aliyun.com/chevalierteam/zhanhai-kit/plugins/logger/zap"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// DataProvider 数据提供者
type DataProvider struct {
	mySQL *gorm.DB
	redis *redis.Client
	log   *zhlog.Helper
}

// NewDataProvider 创建数据提供者
func NewDataProvider(mysql *gorm.DB, redis *redis.Client, log *zhlog.Helper) *DataProvider {
	return &DataProvider{mySQL: mysql, redis: redis, log: log}
}

// TODO: 在这里添加您的数据仓库提供方法
// 示例:
// func (d *DataProvider) ProvideUserRepo() *user.UserRepo {
//     return user.NewUserRepo(d.mySQL, d.redis, d.log)
// }
`

// internal/handler/provider.go 模板
const handlerProviderTemplate = `package handler

import (
	"{{.ImportPrefix}}/internal/data"

	zhlog "codeup.aliyun.com/chevalierteam/zhanhai-kit/plugins/logger/zap"
)

// HandlerProvider 处理器提供者
type HandlerProvider struct {
	data *data.DataProvider
	log  *zhlog.Helper
}

// NewHandlerProvider 创建处理器提供者
func NewHandlerProvider(dataProvider *data.DataProvider, log *zhlog.Helper) *HandlerProvider {
	return &HandlerProvider{
		data: dataProvider,
		log:  log,
	}
}

// TODO: 在这里添加您的处理器提供方法
// 示例:
// func (h *HandlerProvider) ProvideUserHandler() *user.UserHandler {
//     return user.NewUserHandler(h.data.ProvideUserRepo(), h.log)
// }
`

// internal/server/provider.go 模板
const serverProviderTemplate = `package server

import (
	"{{.ImportPrefix}}/internal/handler"
	"{{.ImportPrefix}}/internal/server/http"

	zhlog "codeup.aliyun.com/chevalierteam/zhanhai-kit/plugins/logger/zap"
)

// ServerProvider 服务器提供者
type ServerProvider struct {
	HTTPServer *http.HTTPServer
	log        *zhlog.Helper
}

// NewServerProvider 创建服务器提供者
func NewServerProvider(handlerProvider *handler.HandlerProvider, log *zhlog.Helper, metrics *prometheus.Metrics, tracer *jaeger.TracingProvider) *ServerProvider {
	// 创建HTTP服务器
	httpServer := http.NewHTTPServer(handlerProvider, log, metrics, tracer)

	// 设置路由
	httpServer.SetupRoutes()

	return &ServerProvider{
		HTTPServer: httpServer,
		log:        log,
	}
}
`

// internal/server/http/server.go 模板
const httpServerTemplate = `package http

import (
	"fmt"
	"net/http"

	"{{.ImportPrefix}}/internal/handler"
	"{{.ModulePath}}/pkg/jaeger"
	"{{.ModulePath}}/pkg/prometheus"
	"{{.ModulePath}}/utils/common"

	"github.com/gin-gonic/gin"
	zhlog "codeup.aliyun.com/chevalierteam/zhanhai-kit/plugins/logger/zap"
)

// HTTPServer HTTP服务器
type HTTPServer struct {
	router  *gin.Engine
	handler *handler.HandlerProvider
	metrics *prometheus.Metrics
	tracer  *jaeger.TracingProvider
	log     *zhlog.Helper
}

// NewHTTPServer 创建HTTP服务器
func NewHTTPServer(handlerProvider *handler.HandlerProvider, log *zhlog.Helper, metrics *prometheus.Metrics, tracer *jaeger.TracingProvider) *HTTPServer {
	router := gin.Default()

	// 添加链路追踪中间件
	if tracer != nil {
		router.Use(tracer.GinMiddleware())
	}

	// 添加Prometheus监控中间件
	if metrics != nil {
		router.Use(metrics.GinMiddleware())
	}

	// 添加CORS中间件
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	return &HTTPServer{
		router:  router,
		handler: handlerProvider,
		metrics: metrics,
		tracer:  tracer,
		log:     log,
	}
}

// SetupRoutes 设置路由
func (s *HTTPServer) SetupRoutes() {
	// Prometheus metrics 端点
	if s.metrics != nil {
		s.router.GET("/metrics", gin.WrapH(s.metrics.Handler()))
	}

	api := s.router.Group("/api/v1")

	// 健康检查
	api.GET("/health", func(c *gin.Context) {
		common.SuccessResponseFunc(c, "服务正常", gin.H{"status": "ok", "service": "{{.AppName}}"})
	})

	// TODO: 在这里添加您的路由
	// 示例:
	// userGroup := api.Group("/users")
	// {
	//     userHandler := s.handler.ProvideUserHandler()
	//     userGroup.GET("", userHandler.ListUsers)
	//     userGroup.POST("", userHandler.CreateUser)
	//     userGroup.GET("/:id", userHandler.GetUser)
	//     userGroup.PUT("/:id", userHandler.UpdateUser)
	//     userGroup.DELETE("/:id", userHandler.DeleteUser)
	// }
}

// Start 启动服务器
func (s *HTTPServer) Start(port string) error {
	s.log.Info("启动HTTP服务器", "port", port)
	return s.router.Run(fmt.Sprintf(":%s", port))
}
`

// README.md 模板
const readmeTemplate = `# {{.AppName}}服务

{{.AppName}}服务的API实现，基于Go语言和Gin框架开发。

## 项目结构

` + "```" + `
{{.AppPath}}/
├── cmd/                    # 应用入口
│   └── main.go            # 主函数，应用启动入口
├── config/                # 配置管理
│   └── config.go          # 配置结构定义和加载逻辑
├── internal/              # 内部包，不对外暴露
│   ├── data/              # 数据层
│   │   └── provider.go    # 数据提供者，管理所有数据仓库
│   ├── handler/           # 处理器层，业务逻辑
│   │   └── provider.go    # 处理器提供者，管理所有处理器
│   └── server/            # 服务器层
│       ├── provider.go    # 服务器提供者，管理所有服务器
│       └── http/          # HTTP服务器实现
│           └── server.go  # HTTP路由和中间件配置
└── README.md              # 项目说明文档
` + "```" + `

## 开发指南

### 1. 添加新的数据仓库

在 ` + "`internal/data/`" + ` 目录下创建新的包，然后在 ` + "`provider.go`" + ` 中添加提供方法。

### 2. 添加新的处理器

在 ` + "`internal/handler/`" + ` 目录下创建新的包，然后在 ` + "`provider.go`" + ` 中添加提供方法。

### 3. 添加新的路由

在 ` + "`internal/server/http/server.go`" + ` 的 ` + "`SetupRoutes`" + ` 方法中添加新的路由。

## 配置说明

应用支持通过TOML配置文件进行配置，默认配置文件为 ` + "`config.toml`" + `。

### 配置示例

` + "```toml" + `
[server]
port = "8080"
mode = "debug"

[database]
host = "localhost"
port = "3306"
user = "root"
password = "root123"
dbname = "{{.PackageName}}_db"

[redis]
host = "localhost"
port = "6379"
password = ""
db = 0

[etcd]
endpoints = ["localhost:2379"]
dial_timeout = 5
` + "```" + `

## 运行方式

` + "```bash" + `
# 使用默认配置
go run cmd/main.go

# 指定配置文件
go run cmd/main.go -config=custom-config.toml
` + "```" + `

## API文档

启动服务后，可以通过以下地址查看API文档：

- Swagger UI: http://localhost:8080/swagger/index.html (如已集成Swagger)
- 健康检查: http://localhost:8080/api/v1/health

## 开发注意事项

1. 所有的业务逻辑应该放在 ` + "`handler`" + ` 层
2. 数据库操作应该放在 ` + "`data`" + ` 层
3. 网络相关的逻辑应该放在 ` + "`server`" + ` 层
4. 配置相关的逻辑应该放在 ` + "`config`" + ` 包
5. 遵循依赖注入的原则，通过Provider模式管理依赖关系
`
