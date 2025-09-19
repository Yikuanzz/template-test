package etcd

import (
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	zhlog "codeup.aliyun.com/chevalierteam/zhanhai-kit/plugins/logger/zap"
)

type EtcdConfig struct {
	Endpoints   []string
	DialTimeout int
}

func NewEtcd(config *EtcdConfig, logger *zhlog.Helper) *clientv3.Client {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   config.Endpoints,                                // Etcd服务器地址
		DialTimeout: time.Duration(config.DialTimeout) * time.Second, // 连接超时时间
	})

	if err != nil {
		logger.Error("Etcd 连接失败", err)
		panic(err)
	}

	logger.Info("Etcd 连接成功")

	return etcdClient
}
