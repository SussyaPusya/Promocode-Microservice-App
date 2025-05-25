package redis

import (
	"context"
	"fmt"

	"gitlab.com/pisya-dev/account-service/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewRedis(config *config.Redis, ctx context.Context) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       0,
	})

	pong := client.Ping(ctx)
	if pong.Err() != nil {
		return nil, pong.Err()
	}

	return client, nil
}
