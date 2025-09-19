package jaeger

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	zhlog "codeup.aliyun.com/chevalierteam/zhanhai-kit/plugins/logger/zap"
)

// JaegerConfig Jaeger配置
type JaegerConfig struct {
	ServiceName string  // 服务名称
	Environment string  // 环境名称 (dev, staging, prod)
	JaegerURL   string  // Jaeger Collector URL
	SampleRatio float64 // 采样率 (0.0-1.0)
	Disabled    bool    // 是否禁用追踪
}

// TracingProvider 追踪提供者
type TracingProvider struct {
	tracer   oteltrace.Tracer
	provider *trace.TracerProvider
	config   *JaegerConfig
	logger   *zhlog.Helper
}

// NewTracingProvider 创建链路追踪提供者
func NewTracingProvider(config *JaegerConfig, logger *zhlog.Helper) (*TracingProvider, error) {
	if config.Disabled {
		logger.Info("Jaeger tracing is disabled")
		return &TracingProvider{
			config: config,
			logger: logger,
		}, nil
	}

	// 创建Jaeger导出器
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.JaegerURL)))
	if err != nil {
		logger.Error("Failed to create Jaeger exporter", "error", err)
		return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	// 创建资源
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion("1.0.0"),
			semconv.DeploymentEnvironment(config.Environment),
		),
	)
	if err != nil {
		logger.Error("Failed to create resource", "error", err)
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 创建TracerProvider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
		trace.WithSampler(trace.TraceIDRatioBased(config.SampleRatio)),
	)

	// 设置全局TracerProvider
	otel.SetTracerProvider(tp)

	// 设置全局传播器
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer := tp.Tracer(config.ServiceName)

	logger.Info("Jaeger tracing initialized",
		"service", config.ServiceName,
		"environment", config.Environment,
		"jaeger_url", config.JaegerURL,
		"sample_ratio", config.SampleRatio,
	)

	return &TracingProvider{
		tracer:   tracer,
		provider: tp,
		config:   config,
		logger:   logger,
	}, nil
}

// GinMiddleware 返回Gin中间件用于自动追踪HTTP请求
func (tp *TracingProvider) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if tp.config.Disabled || tp.tracer == nil {
			c.Next()
			return
		}

		// 从请求头中提取trace context
		ctx := otel.GetTextMapPropagator().Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

		// 开始新的span
		ctx, span := tp.tracer.Start(ctx, fmt.Sprintf("%s %s", c.Request.Method, c.FullPath()),
			oteltrace.WithAttributes(
				semconv.HTTPMethod(c.Request.Method),
				semconv.HTTPRoute(c.FullPath()),
				semconv.HTTPURL(c.Request.URL.String()),
				semconv.HTTPUserAgent(c.Request.UserAgent()),
				semconv.HTTPClientIP(c.ClientIP()),
			),
			oteltrace.WithSpanKind(oteltrace.SpanKindServer),
		)
		defer span.End()

		// 将context传递给请求
		c.Request = c.Request.WithContext(ctx)

		// 处理请求
		c.Next()

		// 记录响应信息
		span.SetAttributes(
			semconv.HTTPStatusCode(c.Writer.Status()),
			semconv.HTTPResponseContentLength(int64(c.Writer.Size())),
		)

		// 如果有错误，记录错误信息
		if len(c.Errors) > 0 {
			span.SetAttributes(attribute.String("error.message", c.Errors.String()))
			span.RecordError(fmt.Errorf("request error: %s", c.Errors.String()))
		}

		// 设置span状态
		if c.Writer.Status() >= 400 {
			span.SetAttributes(attribute.Bool("error", true))
		}
	}
}

// StartSpan 手动开始一个新的span
func (tp *TracingProvider) StartSpan(ctx context.Context, operationName string, attrs ...attribute.KeyValue) (context.Context, oteltrace.Span) {
	if tp.config.Disabled || tp.tracer == nil {
		return ctx, oteltrace.NoopSpan{}
	}

	return tp.tracer.Start(ctx, operationName, oteltrace.WithAttributes(attrs...))
}

// StartSpanWithOptions 使用自定义选项开始span
func (tp *TracingProvider) StartSpanWithOptions(ctx context.Context, operationName string, opts ...oteltrace.SpanStartOption) (context.Context, oteltrace.Span) {
	if tp.config.Disabled || tp.tracer == nil {
		return ctx, oteltrace.NoopSpan{}
	}

	return tp.tracer.Start(ctx, operationName, opts...)
}

// AddSpanAttributes 向当前span添加属性
func (tp *TracingProvider) AddSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	if tp.config.Disabled {
		return
	}

	span := oteltrace.SpanFromContext(ctx)
	if span != nil {
		span.SetAttributes(attrs...)
	}
}

// RecordError 记录错误到当前span
func (tp *TracingProvider) RecordError(ctx context.Context, err error, attrs ...attribute.KeyValue) {
	if tp.config.Disabled {
		return
	}

	span := oteltrace.SpanFromContext(ctx)
	if span != nil {
		span.RecordError(err, oteltrace.WithAttributes(attrs...))
		span.SetAttributes(attribute.Bool("error", true))
	}
}

// SetSpanStatus 设置span状态
func (tp *TracingProvider) SetSpanStatus(ctx context.Context, code oteltrace.StatusCode, description string) {
	if tp.config.Disabled {
		return
	}

	span := oteltrace.SpanFromContext(ctx)
	if span != nil {
		span.SetStatus(code, description)
	}
}

// GetTracer 获取tracer实例
func (tp *TracingProvider) GetTracer() oteltrace.Tracer {
	return tp.tracer
}

// InjectHeaders 将trace context注入到HTTP请求头中
func (tp *TracingProvider) InjectHeaders(ctx context.Context, req *http.Request) {
	if tp.config.Disabled {
		return
	}

	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
}

// ExtractContext 从HTTP请求头中提取trace context
func (tp *TracingProvider) ExtractContext(req *http.Request) context.Context {
	if tp.config.Disabled {
		return req.Context()
	}

	return otel.GetTextMapPropagator().Extract(req.Context(), propagation.HeaderCarrier(req.Header))
}

// Shutdown 关闭追踪提供者
func (tp *TracingProvider) Shutdown(ctx context.Context) error {
	if tp.provider == nil {
		return nil
	}

	// 设置5秒超时
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := tp.provider.Shutdown(shutdownCtx); err != nil {
		tp.logger.Error("Failed to shutdown tracing provider", "error", err)
		return err
	}

	tp.logger.Info("Tracing provider shutdown successfully")
	return nil
}

// DefaultConfig 返回默认配置
func DefaultConfig(serviceName string) *JaegerConfig {
	return &JaegerConfig{
		ServiceName: serviceName,
		Environment: "development",
		JaegerURL:   "http://localhost:14268/api/traces",
		SampleRatio: 1.0, // 开发环境100%采样
		Disabled:    false,
	}
}

// ProductionConfig 返回生产环境配置
func ProductionConfig(serviceName string) *JaegerConfig {
	return &JaegerConfig{
		ServiceName: serviceName,
		Environment: "production",
		JaegerURL:   "http://localhost:14268/api/traces",
		SampleRatio: 0.1, // 生产环境10%采样
		Disabled:    false,
	}
}
