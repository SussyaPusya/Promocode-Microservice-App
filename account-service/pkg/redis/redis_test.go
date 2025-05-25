package redis_test

import (
	"strconv"

	"gitlab.com/pisya-dev/account-service/internal/config"
	"gitlab.com/pisya-dev/account-service/pkg/redis"

	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewRedis_Success(t *testing.T) {
	// Запускаем in-memory Redis
	s, err := miniredis.Run()
	assert.NoError(t, err)
	defer s.Close()

	ctx := context.Background()

	port, _ := strconv.Atoi(s.Port())
	// Конфиг для Redis
	cfg := &config.Redis{
		Host:     s.Host(),
		Port:     port,
		Password: "", // miniredis не требует пароля
	}

	client, err := redis.NewRedis(cfg, ctx)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// Пробуем записать и прочитать значение
	err = client.Set(ctx, "key", "value", 0).Err()
	assert.NoError(t, err)

	val, err := client.Get(ctx, "key").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value", val)
}
