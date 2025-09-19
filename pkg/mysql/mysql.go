package mysql

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	zhlog "codeup.aliyun.com/chevalierteam/zhanhai-kit/plugins/logger/zap"
)

type MySQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func NewMySQL(config *MySQLConfig, logger *zhlog.Helper) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.User, config.Password, config.Host, config.Port, config.DBName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// 获取底层的 sql.DB 对象来配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("获取数据库实例失败", err)
		panic(err)
	}

	// 配置连接池
	sqlDB.SetMaxIdleConns(10)                  // 最大空闲连接数
	sqlDB.SetMaxOpenConns(100)                 // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Hour)        // 连接最大生存时间
	sqlDB.SetConnMaxIdleTime(30 * time.Minute) // 连接最大空闲时间

	// 检查数据库连接
	if err := sqlDB.Ping(); err != nil {
		logger.Error("数据库连接检查失败", err)
		panic(err)
	}

	logger.Info("MySQL 数据库连接成功")
	return db
}
