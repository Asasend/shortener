package sequence

import (
	"context"
	"time"

	// "github.com/cloudwego/kitex/client"
	"github.com/redis/go-redis/v9"
)

// 基于redis实现一个发号器

const (
	sequenceKey  = "shortener:sequence"
	redisTimeout = 5 * time.Second
)

type Redis struct {
	// redis连接
	client *redis.Client
	ctx    context.Context
}

func NewRedis(redisAddr string) Sequence {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	return &Redis{
		client: client,
		ctx:    context.Background(),
	}
}

// Next 实现 Sequence 接口， 生成下一个序列号
func (r *Redis) Next() (uint64, error) {
	// 使用Redis的INCR命令原子性地递增序列号
	ctx, cancel := context.WithTimeout(r.ctx, redisTimeout)
	defer cancel()

	result, err := r.client.Incr(ctx, sequenceKey).Result()
	if err != nil {
		return 0, err
	}

	return uint64(result), nil
}

// Close 关闭Redis连接
func (r *Redis) Close() error {
	return r.client.Close()
}

// Ping 测试Redis连接
func (r *Redis) Ping() error {
	ctx, cancel := context.WithTimeout(r.ctx, redisTimeout)
	defer cancel()

	return r.client.Ping(ctx).Err()
}
