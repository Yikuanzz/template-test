package prometheus

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	zhlog "codeup.aliyun.com/chevalierteam/zhanhai-kit/plugins/logger/zap"
)

// PrometheusConfig Prometheus配置
type PrometheusConfig struct {
	Namespace   string // 命名空间
	Subsystem   string // 子系统
	MetricsPath string // 指标路径，默认 /metrics
}

// Metrics Prometheus指标集合
type Metrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	responseSize    *prometheus.HistogramVec
	requestSize     *prometheus.HistogramVec
	uptime          prometheus.Counter
	registry        *prometheus.Registry
	config          *PrometheusConfig
	logger          *zhlog.Helper
}

// NewMetrics 创建Prometheus指标收集器
func NewMetrics(config *PrometheusConfig, logger *zhlog.Helper) *Metrics {
	if config.MetricsPath == "" {
		config.MetricsPath = "/metrics"
	}

	registry := prometheus.NewRegistry()

	metrics := &Metrics{
		registry: registry,
		config:   config,
		logger:   logger,
	}

	// HTTP请求总数
	metrics.requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests made.",
		},
		[]string{"method", "path", "status_code"},
	)

	// HTTP请求响应时间
	metrics.requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request latencies in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path", "status_code"},
	)

	// HTTP响应大小
	metrics.responseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "http_response_size_bytes",
			Help:      "HTTP response sizes in bytes.",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path", "status_code"},
	)

	// HTTP请求大小
	metrics.requestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "http_request_size_bytes",
			Help:      "HTTP request sizes in bytes.",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	// 应用启动时间
	metrics.uptime = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "uptime_seconds_total",
			Help:      "Total uptime of the application in seconds.",
		},
	)

	// 注册指标
	registry.MustRegister(
		metrics.requestsTotal,
		metrics.requestDuration,
		metrics.responseSize,
		metrics.requestSize,
		metrics.uptime,
	)

	// 启动uptime计数器
	go metrics.startUptimeCounter()

	logger.Info("Prometheus metrics initialized", "namespace", config.Namespace, "subsystem", config.Subsystem)
	return metrics
}

// startUptimeCounter 启动uptime计数器
func (m *Metrics) startUptimeCounter() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		m.uptime.Inc()
	}
}

// GinMiddleware 返回Gin中间件用于收集HTTP指标
func (m *Metrics) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过metrics端点本身
		if c.Request.URL.Path == m.config.MetricsPath {
			c.Next()
			return
		}

		start := time.Now()

		// 记录请求大小
		if c.Request.ContentLength > 0 {
			m.requestSize.WithLabelValues(
				c.Request.Method,
				c.FullPath(),
			).Observe(float64(c.Request.ContentLength))
		}

		c.Next()

		// 计算响应时间
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())

		// 记录指标
		m.requestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			statusCode,
		).Inc()

		m.requestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			statusCode,
		).Observe(duration)

		m.responseSize.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			statusCode,
		).Observe(float64(c.Writer.Size()))
	}
}

// Handler 返回Prometheus指标处理器
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// RegisterMetrics 注册自定义指标
func (m *Metrics) RegisterMetrics(collectors ...prometheus.Collector) error {
	for _, collector := range collectors {
		if err := m.registry.Register(collector); err != nil {
			m.logger.Error("Failed to register metric", "error", err)
			return err
		}
	}
	return nil
}

// GetRegistry 获取Prometheus注册器
func (m *Metrics) GetRegistry() *prometheus.Registry {
	return m.registry
}

// RecordCustomMetric 记录自定义指标的辅助方法
func (m *Metrics) RecordCustomMetric(name string, value float64, labels map[string]string) {
	// 这个方法可以用于记录自定义业务指标
	// 具体实现可以根据需要扩展
	m.logger.Debug("Recording custom metric", "name", name, "value", value, "labels", labels)
}

// DefaultConfig 返回默认配置
func DefaultConfig(serviceName string) *PrometheusConfig {
	return &PrometheusConfig{
		Namespace:   "go_template",
		Subsystem:   serviceName,
		MetricsPath: "/metrics",
	}
}
