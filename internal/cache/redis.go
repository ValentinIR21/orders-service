package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

// Инициализация соединения с redis
func NewConnectionRedis(ctx context.Context, addr string) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisClient{
		Client: rdb,
	}, nil
}

func (r *RedisClient) CloseConnectionRedis() error {

	return r.Client.Close()

}
