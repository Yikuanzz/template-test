package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	zhlog "codeup.aliyun.com/chevalierteam/zhanhai-kit/plugins/logger/zap"
)

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func NewRedis(config *RedisConfig, logger *zhlog.Helper) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", config.Host, config.Port), // Redis服务器地址
		Password:     config.Password,                                // Redis密码
		DB:           config.DB,                                      // 使用默认数据库
		PoolSize:     100,                                            // 连接池大小
		MinIdleConns: 10,                                             // 最小空闲连接数
		MaxIdleConns: 20,                                             // 最大空闲连接数
		MaxRetries:   3,                                              // 最大重试次数
		DialTimeout:  5 * time.Second,                                // 连接超时时间
		ReadTimeout:  3 * time.Second,                                // 读取超时时间
		WriteTimeout: 3 * time.Second,                                // 写入超时时间
		PoolTimeout:  4 * time.Second,                                // 连接池超时时间
	})

	// 检查Redis连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Error("Redis 缓存连接失败", err)
		panic(err)
	}

	logger.Info("Redis 缓存连接成功")

	return redisClient
}
