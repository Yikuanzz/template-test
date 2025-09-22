# 配置系统使用示例

## 概述

重构后的配置系统支持通过 `.env` 文件存储敏感信息，通过 `.toml` 文件存储非敏感配置信息。

## 文件结构

```
项目根目录/
├── .env                 # 敏感信息（不提交到版本控制）
├── config.toml          # 非敏感配置
├── env.example          # 环境变量示例文件
└── config.toml.example  # 配置文件示例
```

## 使用方法

### 1. 基本用法

```go
package main

import (
    "log"
    "your-project/config"
)

func main() {
    // 加载配置，同时指定 .toml 和 .env 文件路径
    cfg, err := config.LoadConfig("config.toml", ".env")
    if err != nil {
        log.Fatal("加载配置失败:", err)
    }
    
    // 使用配置
    log.Printf("服务器端口: %s", cfg.Server.Port)
    log.Printf("数据库主机: %s", cfg.Database.Host)
    log.Printf("Redis 密码: %s", cfg.Redis.Password)
}
```

### 2. 配置分离说明

#### .env 文件（敏感信息）
```bash
# 数据库敏感信息
DB_HOST=localhost
DB_PORT=3406
DB_USER=root
DB_PASSWORD=your-secret-password
DB_NAME=myapp_db

# Redis 敏感信息
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your-redis-password
REDIS_DB=0

# 安全配置
JWT_SECRET=your-jwt-secret-key

# 监控配置
JAEGER_URL=http://localhost:14268/api/traces
JAEGER_SAMPLE_RATIO=1.0
JAEGER_DISABLED=false

# 日志配置
LOG_LEVEL=info
LOG_FORMAT=json

# 服务器模式
SERVER_MODE=release
```

#### config.toml 文件（非敏感配置）
```toml
[server]
port = "8080"
mode = "debug"

[database]
max_open_conns = 25
max_idle_conns = 10
conn_max_lifetime = "1h"
conn_max_idle_time = "30m"
connect_timeout = "10s"
query_timeout = "30s"

[redis]
pool_size = 10
min_idle_conns = 5
max_retries = 3
dial_timeout = "5s"
read_timeout = "3s"
write_timeout = "3s"

[monitoring]
prometheus_namespace = "myapp"
prometheus_subsystem = "service"
metrics_path = "/metrics"
collect_interval = "15s"
jaeger_service_name = "myapp-service"
jaeger_environment = "production"

[logging]
level = "info"
format = "json"
output = "stdout"
filename = "logs/app.log"
max_size = 100
max_age = 30
max_backups = 3
compress = true

[security]
jwt_expire_hours = 24
bcrypt_cost = 12
rate_limit_requests = 100
rate_limit_window = "1m"

[app]
name = "myapp"
version = "1.0.0"
request_timeout = "30s"
shutdown_timeout = "10s"

[cache]
default_ttl = "1h"
cleanup_interval = "10m"
max_memory = "100MB"
```

### 3. 环境变量优先级

配置加载的优先级顺序：
1. 环境变量（最高优先级）
2. .env 文件
3. config.toml 文件
4. 默认值（最低优先级）

### 4. 配置结构说明

#### 敏感信息字段（从环境变量加载）
- `DatabaseConfig`: Host, Port, User, Password, DBName
- `RedisConfig`: Host, Port, Password, DB
- `EtcdConfig`: Endpoints
- `SecurityConfig`: JWTSecret
- `MonitoringConfig`: JaegerURL, JaegerSampleRatio, JaegerDisabled
- `LoggingConfig`: Level, Format
- `ServerConfig`: Mode

#### 非敏感信息字段（从 TOML 加载）
- 所有连接池配置
- 超时配置
- 监控指标配置
- 日志轮转配置
- 应用元信息
- 缓存配置

### 5. 错误处理

```go
cfg, err := config.LoadConfig("config.toml", ".env")
if err != nil {
    // 处理配置加载错误
    log.Printf("配置加载失败: %v", err)
    // 可以使用默认配置
    cfg = config.DefaultConfig()
}
```

### 6. 默认配置

如果配置文件不存在或加载失败，可以使用默认配置：

```go
cfg := config.DefaultConfig()
```

## 安全建议

1. **永远不要将 `.env` 文件提交到版本控制系统**
2. **使用强密码和密钥**
3. **定期轮换敏感信息**
4. **在生产环境中使用环境变量而不是文件**
5. **限制配置文件的访问权限**

## 部署建议

### 开发环境
- 使用 `.env` 文件存储本地开发配置
- 使用 `config.toml` 存储应用配置

### 生产环境
- 通过环境变量设置所有敏感信息
- 使用 `config.toml` 存储非敏感配置
- 考虑使用配置管理工具（如 Consul、Etcd）
