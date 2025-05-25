package config_test

import (
	"os"
	"testing"

	"gitlab.com/pisya-dev/account-service/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig_Success(t *testing.T) {
	t.Setenv("GRPC_PORT", "50051")
	t.Setenv("PG_HOST", "localhost")
	t.Setenv("PG_PORT", "5432")
	t.Setenv("PGPG_DATABASE_DATA", "account_db")
	t.Setenv("PG_USER", "postgres")
	t.Setenv("PG_PASS", "password")
	t.Setenv("PG_MAXCONN", "10")
	t.Setenv("PG_MINCONN", "5")
	t.Setenv("REDIS_HOST", "localhost")
	t.Setenv("REDIS_PORT", "6379")
	t.Setenv("REDIS_PASS", "redispass")

	cfg, err := config.NewConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, 50051, cfg.GRPCPort)
	assert.Equal(t, "localhost", cfg.Postgres.Host)
	assert.Equal(t, 5432, cfg.Postgres.Port)
	assert.Equal(t, "account_db", cfg.Postgres.Database)
	assert.Equal(t, "postgres", cfg.Postgres.User)
	assert.Equal(t, "password", cfg.Postgres.Password)
	assert.Equal(t, int32(10), cfg.Postgres.MaxConn)
	assert.Equal(t, int32(5), cfg.Postgres.MinConn)

	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, 6379, cfg.Redis.Port)
	assert.Equal(t, "redispass", cfg.Redis.Password)
}

func TestNewConfig_MissingEnv(t *testing.T) {
	// Не задаём переменных окружения
	os.Clearenv()

	cfg, err := config.NewConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Проверим, что поля пустые
	assert.Equal(t, 0, cfg.GRPCPort)
	assert.Empty(t, cfg.Postgres.Host)
	assert.Empty(t, cfg.Redis.Host)
}
