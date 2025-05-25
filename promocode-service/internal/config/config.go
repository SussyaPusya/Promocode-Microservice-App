package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env              string `env:"ENV"`
	PostgresUser     string `env:"POSTGRES_USER"`
	PostgresPassword string `env:"POSTGRES_PASSWORD"`
	PostgresHost     string `env:"POSTGRES_HOST"`
	PostgresPort     int    `env:"POSTGRES_PORT"`
	PostgresDb       string `env:"POSTGRES_DB"`
	GRPCHost         string `env:"GRPC_HOST"`
	GRPCPort         int    `env:"GRPC_PORT"`

	RedisPort          int    `env:"REDIS_PORT"`
	RedisHost          string `env:"REDIS_HOST"`
	AccountServiceAddr string `env:"ACCOUNT_SERVICE_ADDR"`
}

func MustLoad() *Config {

	const op = "config.MustLoad"

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic("failed to read env " + err.Error())
	}

	return &cfg
}
