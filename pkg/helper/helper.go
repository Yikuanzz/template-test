package helper

import (
	"os"

	"codeup.aliyun.com/chevalierteam/zhanhai-kit/core/logger"
	zhlog "codeup.aliyun.com/chevalierteam/zhanhai-kit/plugins/logger/zap"
)

// HelperConfig 日志配置结构
type HelperConfig struct {
	Level      logger.Level // 日志级别
	Output     string       // 输出目标: "stdout", "stderr", "file"
	LogFile    string       // 日志文件路径
	MaxSize    int          // 单个文件最大大小(MB)
	MaxAge     int          // 保留天数
	MaxBackups int          // 最大备份数
	Compress   bool         // 是否压缩旧日志
	LocalTime  bool         // 是否使用本地时间
}

// NewLogger 创建日志记录器
func NewLogger(config *HelperConfig) *zhlog.Helper {
	// 设置输出目标
	var writer zhlog.Option
	switch config.Output {
	case "stdout":
		writer = zhlog.WithWriter(os.Stdout)
	case "stderr":
		writer = zhlog.WithWriter(os.Stderr)
	case "file":
		writer = zhlog.WithFileRotation(&zhlog.FileOptions{
			Filename:   config.LogFile,
			MaxSize:    config.MaxSize,
			MaxAge:     config.MaxAge,
			MaxBackups: config.MaxBackups,
			Compress:   config.Compress,
			LocalTime:  config.LocalTime,
		})
	default:
		writer = zhlog.WithWriter(os.Stderr)
	}

	// 创建日志记录器
	log := zhlog.NewLogger(
		zhlog.WithLevel(config.Level),
		writer,
		zhlog.WithMessageKey("message"),
		zhlog.WithCallerSkip(2),
	)

	// 返回 Helper 实例
	return zhlog.NewHelper(log)
}

// NewSimpleLogger 创建简单日志记录器（使用默认配置）
func NewSimpleLogger() *zhlog.Helper {
	return NewLogger(&HelperConfig{
		Level:      logger.LevelInfo,
		Output:     "stderr",
		LogFile:    "logs/app.log",
		MaxSize:    50,
		MaxAge:     7,
		MaxBackups: 3,
		Compress:   true,
		LocalTime:  true,
	})
}
