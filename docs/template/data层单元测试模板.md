# Data层单元测试模板

## 概述

这是一个基于 `github.com/ory/dockertest/v3` 的 Go 语言数据层单元测试模板，用于在测试环境中启动临时 MySQL 容器进行集成测试。

## 核心特性

- ✅ 使用随机端口避免端口冲突
- ✅ 支持中文字符（UTF8MB4字符集）
- ✅ 优化的连接池配置避免资源泄漏
- ✅ 自动容器清理和资源管理
- ✅ 稳定的连接重试机制
- ✅ 支持 Redis 集成测试

## 完整代码模板

### 1. 基础结构和导入

```go
package your_package_name

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"your-project/db/model" // 替换为你的模型包路径

	zhlog "codeup.aliyun.com/chevalierteam/zhanhai-kit/plugins/logger/zap" // 替换为你的日志包
	_ "github.com/go-sql-driver/mysql"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 全局变量，供所有测试使用
var (
	db          *gorm.DB
	redisClient *redis.Client
	logger      *zhlog.Helper
)
```

### 2. 随机端口获取函数

```go
// getRandomPort 获取一个可用的随机端口，避免端口冲突
func getRandomPort() (string, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", err
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return fmt.Sprintf("%d", addr.Port), nil
}
```

### 3. TestMain 函数（核心配置）

```go
func TestMain(m *testing.M) {
	// 创建 dockertest pool
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("无法连接到 Docker: %s", err)
	}

	// 获取随机端口，避免3306端口被占用
	randomPort, err := getRandomPort()
	if err != nil {
		log.Fatalf("无法获取随机端口: %s", err)
	}

	// 启动 MySQL 容器，使用随机端口映射
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mysql",
		Tag:        "5.7",
		Env: []string{
			"MYSQL_ROOT_PASSWORD=testpassword",
			"MYSQL_DATABASE=testdb",
			"MYSQL_USER=testuser",
			"MYSQL_PASSWORD=testpass",
			"MYSQL_CHARSET=utf8mb4",                    // 支持中文字符
			"MYSQL_COLLATION=utf8mb4_unicode_ci",       // Unicode排序规则
		},
		Cmd: []string{
			"mysqld",
			"--character-set-server=utf8mb4",           // 服务器字符集
			"--collation-server=utf8mb4_unicode_ci",    // 服务器排序规则
			"--default-authentication-plugin=mysql_native_password", // 兼容性设置
		},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"3306/tcp": {{HostPort: randomPort}},
		},
	})
	if err != nil {
		log.Fatalf("无法启动 MySQL 容器: %s", err)
	}

	// 设置清理函数，确保测试结束后清理容器
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("清理容器失败: %s", err)
		}
	}()

	// 等待容器完全启动（重要：避免连接过早）
	log.Printf("等待MySQL容器启动，端口: %s", randomPort)
	time.Sleep(10 * time.Second)

	// 构建数据库连接字符串，包含超时和字符集配置
	dsn := fmt.Sprintf("testuser:testpass@tcp(localhost:%s)/testdb?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=True&loc=Local&timeout=30s&readTimeout=30s&writeTimeout=30s", randomPort)

	// 等待数据库启动并重试连接
	var gormDB *gorm.DB
	if err := pool.Retry(func() error {
		var err error
		gormDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: nil, // 禁用GORM日志以减少测试输出干扰
		})
		if err != nil {
			return err
		}

		// 测试连接
		sqlDB, err := gormDB.DB()
		if err != nil {
			return err
		}

		// 在重试阶段使用保守的连接设置
		sqlDB.SetConnMaxLifetime(time.Minute * 5)
		sqlDB.SetMaxOpenConns(1) // 测试时使用较少的连接数
		sqlDB.SetMaxIdleConns(1)

		return sqlDB.Ping()
	}); err != nil {
		log.Fatalf("无法连接到数据库: %s", err)
	}

	// 配置数据库连接池 - 使用优化的设置避免资源泄漏
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("获取数据库实例失败: %s", err)
	}
	sqlDB.SetMaxOpenConns(2)                   // 减少最大连接数
	sqlDB.SetMaxIdleConns(1)                   // 减少空闲连接数
	sqlDB.SetConnMaxLifetime(time.Minute * 10) // 减少连接生命周期
	sqlDB.SetConnMaxIdleTime(time.Minute * 5)  // 设置空闲连接超时

	// 自动迁移数据库表
	if err := gormDB.AutoMigrate(&model.YourModel{}); err != nil { // 替换为你的模型
		log.Fatalf("数据库迁移失败: %s", err)
	}

	// 初始化 Redis 客户端（如果需要）
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 初始化日志
	logger = zhlog.NewHelper(zhlog.NewLogger())

	// 设置全局变量
	db = gormDB

	// 运行测试
	code := m.Run()

	// 清理资源
	if err := sqlDB.Close(); err != nil {
		log.Printf("关闭数据库连接失败: %s", err)
	}
	redisClient.Close()

	os.Exit(code)
}
```

### 4. 测试辅助函数

```go
// createTestRepo 创建测试用的仓储实例
func createTestRepo() *YourRepo {
	return NewYourRepo(db, redisClient, logger) // 替换为你的仓储构造函数
}

// cleanupTestData 清理测试数据，确保测试独立性
func cleanupTestData(t *testing.T) {
	t.Helper()
	if err := db.Exec("DELETE FROM your_table_name").Error; err != nil {
		t.Fatalf("清理测试数据失败: %v", err)
	}
}

// createTestData 创建测试数据
func createTestData(t *testing.T, data *model.YourModel) *model.YourModel {
	t.Helper()
	if err := db.Create(data).Error; err != nil {
		t.Fatalf("创建测试数据失败: %v", err)
	}
	return data
}
```

### 5. 完整测试示例

```go
func TestYourRepo_GetAll(t *testing.T) {
	repo := createTestRepo()
	ctx := context.Background()

	t.Run("获取所有数据", func(t *testing.T) {
		// 清理数据
		cleanupTestData(t)

		// 创建测试数据
		createTestData(t, &model.YourModel{
			ID:   "test-1",
			Name: "测试数据1",
		})
		createTestData(t, &model.YourModel{
			ID:   "test-2", 
			Name: "测试数据2",
		})

		// 执行测试
		results, err := repo.GetAll(ctx)

		// 验证结果
		if err != nil {
			t.Fatalf("GetAll 失败: %v", err)
		}
		if len(results) != 2 {
			t.Fatalf("期望获取 2 条记录，实际获取 %d 条", len(results))
		}

		// 验证数据内容
		found1, found2 := false, false
		for _, item := range results {
			if item.ID == "test-1" && item.Name == "测试数据1" {
				found1 = true
			}
			if item.ID == "test-2" && item.Name == "测试数据2" {
				found2 = true
			}
		}
		if !found1 || !found2 {
			t.Fatal("获取的数据内容不正确")
		}
	})

	t.Run("空数据时返回空切片", func(t *testing.T) {
		// 清理数据
		cleanupTestData(t)

		// 执行测试
		results, err := repo.GetAll(ctx)

		// 验证结果
		if err != nil {
			t.Fatalf("GetAll 失败: %v", err)
		}
		if len(results) != 0 {
			t.Fatalf("期望获取 0 条记录，实际获取 %d 条", len(results))
		}
	})
}

func TestYourRepo_Create(t *testing.T) {
	repo := createTestRepo()
	ctx := context.Background()

	t.Run("创建数据", func(t *testing.T) {
		// 清理数据
		cleanupTestData(t)

		// 准备测试数据
		data := &model.YourModel{
			ID:   "create-test",
			Name: "创建测试数据",
		}

		// 执行测试
		err := repo.Create(ctx, data)

		// 验证结果
		if err != nil {
			t.Fatalf("Create 失败: %v", err)
		}

		// 验证数据是否已创建
		var count int64
		if err := db.Model(&model.YourModel{}).Where("id = ?", "create-test").Count(&count).Error; err != nil {
			t.Fatalf("查询创建的数据失败: %v", err)
		}
		if count != 1 {
			t.Fatalf("期望创建 1 条记录，实际创建 %d 条", count)
		}
	})
}
```

## 关键配置说明

### 1. 随机端口机制
- 使用 `getRandomPort()` 函数获取可用端口
- 避免3306端口被占用导致的测试失败
- 每次测试运行都使用不同的端口

### 2. 字符集配置
- 使用 `utf8mb4` 字符集完全支持中文字符
- 设置 `utf8mb4_unicode_ci` 排序规则
- 在连接字符串中明确指定字符集参数

### 3. 连接池优化
- 设置合理的最大连接数（2个）
- 配置连接生命周期和空闲超时
- 避免连接泄漏和资源浪费

### 4. 容器启动等待
- 增加10秒等待时间确保MySQL完全启动
- 使用 `pool.Retry` 机制进行连接重试
- 在重试阶段使用保守的连接设置

### 5. 测试数据管理
- 每个测试前后清理数据确保独立性
- 使用 `t.Helper()` 标记辅助函数
- 提供数据创建和清理的辅助函数

## 使用说明

1. **替换包名和导入路径**：将模板中的包名和导入路径替换为你的实际路径
2. **替换模型类型**：将 `model.YourModel` 替换为你的实际模型
3. **替换仓储类型**：将 `YourRepo` 替换为你的实际仓储类型
4. **调整表名**：在清理函数中替换为你的实际表名
5. **运行测试**：使用 `go test -v` 运行测试

## 常见问题解决

### 1. unexpected EOF 错误
- 增加容器启动等待时间
- 优化连接池配置
- 添加连接超时设置

### 2. 字符编码错误
- 确保使用 `utf8mb4` 字符集
- 在容器启动参数中设置字符集
- 在连接字符串中指定字符集

### 3. 端口冲突
- 使用随机端口获取函数
- 避免硬编码端口号

### 4. 连接超时
- 增加连接超时时间
- 使用 `pool.Retry` 机制
- 设置合理的重试间隔

这个模板提供了完整的、生产就绪的测试环境配置，可以直接用于你的项目中。
