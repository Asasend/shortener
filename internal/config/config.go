package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf

	ShortUrlDB struct {
		DSN string
	}

	// 添加 Sequence 配置
	Sequence struct {
		DSN string
	}

	BaseString string // base62指定基础字符串

	ShortUrlBlackList []string // 短链接黑名单

	ShortDomain string // 短链接域名

	CacheRedis cache.CacheConf // redis缓存
}
