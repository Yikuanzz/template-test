# 监控指南

本文档介绍如何使用项目中集成的监控和追踪功能。

## 🔍 监控架构

```
应用程序 ──→ Prometheus ──→ Grafana
    │
    └─────→ Jaeger
```

### 组件说明

- **Prometheus**: 指标收集和存储
- **Grafana**: 指标可视化和告警
- **Jaeger**: 分布式链路追踪

## 📊 Prometheus 监控

### 内置指标

项目自动收集以下指标：

#### HTTP 请求指标
- `go_template_http_requests_total`: HTTP请求总数
- `go_template_http_request_duration_seconds`: HTTP请求响应时间
- `go_template_http_response_size_bytes`: HTTP响应大小
- `go_template_http_request_size_bytes`: HTTP请求大小

#### 应用指标  
- `go_template_uptime_seconds_total`: 应用运行时间

### 自定义指标

在应用中添加自定义业务指标：

```go
package main

import (
    "github.com/prometheus/client_golang/prometheus"
    "your-module/pkg/prometheus"
)

// 创建自定义指标
var (
    userRegistrations = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Namespace: "go_template",
            Subsystem: "user",
            Name:      "registrations_total",
            Help:      "Total number of user registrations.",
        },
        []string{"source"},
    )

    activeConnections = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Namespace: "go_template", 
            Subsystem: "websocket",
            Name:      "active_connections",
            Help:      "Number of active WebSocket connections.",
        },
        []string{"type"},
    )
)

func init() {
    // 注册自定义指标到Prometheus
    metrics := prometheus.NewMetrics(/* config */, /* logger */)
    metrics.RegisterMetrics(userRegistrations, activeConnections)
}

// 在业务代码中使用
func registerUser(source string) {
    // 业务逻辑...
    
    // 增加用户注册计数
    userRegistrations.WithLabelValues(source).Inc()
}

func handleWebSocketConnection(connType string) {
    // 增加活跃连接数
    activeConnections.WithLabelValues(connType).Inc()
    
    defer func() {
        // 连接关闭时减少计数
        activeConnections.WithLabelValues(connType).Dec()
    }()
    
    // 连接处理逻辑...
}
```

### 查询示例

在Prometheus UI中使用的查询示例：

```promql
# HTTP请求速率
rate(go_template_http_requests_total[5m])

# 95分位数响应时间
histogram_quantile(0.95, rate(go_template_http_request_duration_seconds_bucket[5m]))

# 错误率
rate(go_template_http_requests_total{status_code!~"2.."}[5m]) / rate(go_template_http_requests_total[5m])

# 应用实例数
up{job="go-template-apps"}
```

## 📈 Grafana 仪表板

### 默认仪表板

项目提供了一个预配置的Grafana仪表板，包含：

1. **HTTP请求速率图表**
2. **响应时间分布图表** 
3. **状态码分布饼图**
4. **应用运行时间图表**

### 访问仪表板

1. 打开 http://localhost:3000
2. 使用 `admin/admin` 登录
3. 导航到 "Go Microservice Dashboard"

### 自定义仪表板

创建新的仪表板：

1. 点击 "+" → "Dashboard"
2. 添加面板选择指标
3. 配置查询和可视化选项
4. 保存仪表板

#### 常用可视化

**响应时间热图**
```promql
sum(rate(go_template_http_request_duration_seconds_bucket[5m])) by (le)
```

**请求量趋势**
```promql
sum(rate(go_template_http_requests_total[5m])) by (method, path)
```

**错误率监控**
```promql
(
  sum(rate(go_template_http_requests_total{status_code!~"2.."}[5m])) 
  / 
  sum(rate(go_template_http_requests_total[5m]))
) * 100
```

## 🔍 Jaeger 链路追踪

### 自动追踪

项目自动为所有HTTP请求创建traces，包含：

- 请求方法和路径
- 响应状态码
- 请求和响应大小
- 客户端IP和User-Agent

### 手动添加Span

在关键业务逻辑中添加自定义span：

```go
package handler

import (
    "context"
    "your-module/pkg/jaeger"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"
)

type UserHandler struct {
    tracer *jaeger.TracingProvider
}

func (h *UserHandler) CreateUser(c *gin.Context) {
    ctx := c.Request.Context()
    
    // 创建数据库操作span
    ctx, dbSpan := h.tracer.StartSpan(ctx, "database.create_user",
        attribute.String("operation", "INSERT"),
        attribute.String("table", "users"),
    )
    defer dbSpan.End()
    
    // 数据库操作
    if err := h.createUserInDB(ctx, user); err != nil {
        h.tracer.RecordError(ctx, err,
            attribute.String("error.type", "database_error"),
        )
        h.tracer.SetSpanStatus(ctx, trace.StatusCodeError, "Failed to create user")
        return
    }
    
    // 创建缓存操作span
    ctx, cacheSpan := h.tracer.StartSpan(ctx, "cache.set_user")
    defer cacheSpan.End()
    
    // 缓存操作
    h.cacheUser(ctx, user)
    
    h.tracer.SetSpanStatus(ctx, trace.StatusCodeOk, "User created successfully")
}

func (h *UserHandler) createUserInDB(ctx context.Context, user *User) error {
    // 在数据库操作中也可以添加更细粒度的span
    ctx, span := h.tracer.StartSpan(ctx, "database.insert",
        attribute.String("sql.table", "users"),
        attribute.Int("user.id", user.ID),
    )
    defer span.End()
    
    // 执行数据库操作...
    return nil
}
```

### 跨服务追踪

在调用其他服务时传播trace context：

```go
func callExternalService(ctx context.Context, tracer *jaeger.TracingProvider) {
    // 创建HTTP客户端
    client := &http.Client{}
    
    // 创建请求
    req, _ := http.NewRequestWithContext(ctx, "GET", "http://other-service/api", nil)
    
    // 注入trace context到请求头
    tracer.InjectHeaders(ctx, req)
    
    // 发送请求
    resp, err := client.Do(req)
    // 处理响应...
}
```

### 查看Traces

1. 打开 http://localhost:16686
2. 选择服务名称
3. 设置时间范围
4. 点击 "Find Traces"
5. 点击具体trace查看详细信息

## 🚨 告警配置

### Prometheus 告警规则

创建 `monitoring/prometheus/alert_rules.yml`：

```yaml
groups:
  - name: go-template-alerts
    rules:
      - alert: HighErrorRate
        expr: (sum(rate(go_template_http_requests_total{status_code!~"2.."}[5m])) / sum(rate(go_template_http_requests_total[5m]))) * 100 > 5
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value }}% for the last 5 minutes"

      - alert: HighResponseTime
        expr: histogram_quantile(0.95, rate(go_template_http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High response time detected"
          description: "95th percentile response time is {{ $value }}s"

      - alert: ServiceDown
        expr: up{job="go-template-apps"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Service is down"
          description: "{{ $labels.instance }} has been down for more than 1 minute"
```

### Grafana 告警

1. 在仪表板面板中点击 "Alert"
2. 设置查询条件和阈值
3. 配置通知渠道（邮件、Slack等）
4. 测试告警规则

## 📊 性能优化

### 采样策略

根据环境调整Jaeger采样率：

```go
// 开发环境 - 100%采样
config := jaeger.DefaultConfig("my-service")
config.SampleRatio = 1.0

// 生产环境 - 10%采样
config := jaeger.ProductionConfig("my-service") 
config.SampleRatio = 0.1

// 高流量环境 - 1%采样
config.SampleRatio = 0.01
```

### Prometheus 优化

```yaml
# prometheus.yml
global:
  scrape_interval: 15s     # 全局采集间隔
  evaluation_interval: 15s # 规则评估间隔

scrape_configs:
  - job_name: 'go-template-apps'
    scrape_interval: 5s    # 针对特定job的采集间隔
    scrape_timeout: 3s     # 采集超时时间
    metrics_path: '/metrics'
    honor_labels: true
```

## 🔧 故障排除

### 常见问题

1. **指标不显示**
   - 检查应用是否暴露 `/metrics` 端点
   - 确认Prometheus配置中的目标地址正确
   - 检查防火墙设置

2. **Traces不显示**
   - 确认Jaeger Collector地址正确
   - 检查采样率设置
   - 查看应用日志中的错误信息

3. **Grafana仪表板为空**
   - 确认Prometheus作为数据源配置正确
   - 检查时间范围设置
   - 验证查询语句语法

### 调试命令

```bash
# 检查Prometheus targets状态
curl http://localhost:9090/api/v1/targets

# 检查应用metrics端点
curl http://localhost:8080/metrics

# 查看Jaeger健康状态
curl http://localhost:16686/api/services

# 测试Grafana API
curl http://admin:admin@localhost:3000/api/health
```

## 📚 最佳实践

### 指标命名

遵循Prometheus命名约定：
- 使用下划线分隔单词
- 包含单位后缀（`_seconds`, `_bytes`, `_total`）
- 使用有意义的标签

### Trace 设计

- 为每个重要操作创建span
- 使用有意义的span名称
- 添加相关属性和标签
- 正确处理错误情况

### 告警策略

- 设置合理的阈值和时间窗口
- 区分警告和严重告警
- 避免告警风暴
- 定期审查和调整告警规则
