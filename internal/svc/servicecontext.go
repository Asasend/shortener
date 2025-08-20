package svc

import (
	// "context"
	"shortener/internal/config"
	"shortener/model"
	"shortener/sequence"

	"github.com/zeromicro/go-zero/core/bloom"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config        config.Config
	ShortUrlModel model.ShortUrlMapModel

	Sequence sequence.Sequence // 这就是基于MySQL的发号器

	ShortUrlBlackList map[string]struct{}

	// bloom filter
	Filter *bloom.Filter
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.ShortUrlDB.DSN)
	// 把配置文件中配置的黑名单加载到map， 方便后续查询判断
	m := make(map[string]struct{}, len(c.ShortUrlBlackList))
	for _, v := range c.ShortUrlBlackList {
		m[v] = struct{}{}
	}

	// 初始化布隆过滤器
	// 初始化 redisBitSet
	store := redis.New(c.CacheRedis[0].Host, func(r *redis.Redis) {
		r.Type = redis.NodeType
	})

	// 声明一个bitSet, key = "bloom_filter"名且bits是20MB
	filter := bloom.New(store, "bloom_filter", 20*(1<<20))

	// 创建 ServiceContext 实例
	svcCtx := &ServiceContext{
		Config:            c,
		ShortUrlModel:     model.NewShortUrlMapModel(conn, c.CacheRedis),
		Sequence:          sequence.NewMySQL(c.Sequence.DSN),
		ShortUrlBlackList: m,
		Filter:            filter,
	}

	// 加载已有的短链接数据到布隆过滤器
	loadDateToBloomFilter(svcCtx)

	return svcCtx
}

// loadDateToBloomFilter 加载已有的短链接数据到布隆过滤器
func loadDateToBloomFilter(svcCtx *ServiceContext) {
	// 直接使用SQL查询所有有效的短链接
	conn := sqlx.NewMysql(svcCtx.Config.ShortUrlDB.DSN)
	query := "SELECT surl FROM short_url_map WHERE is_del = 0 AND surl IS NOT NULL"

	// 使用 QueryRows 方法查询多条记录
	var surls []string
	err := conn.QueryRows(&surls, query)
	if err != nil {
		logx.Errorw("Failed to query existing short urls", logx.LogField{Key: "err", Value: err.Error()})
		return
	}

	// 批量添加到布隆过滤器
	count := 0
	for _, surl := range surls {
		if surl != "" {
			svcCtx.Filter.Add([]byte(surl))
			count++
		}
	}

	logx.Infow("Successfully loaded short urls to bloom filter",
		logx.LogField{Key: "count", Value: count})
}
