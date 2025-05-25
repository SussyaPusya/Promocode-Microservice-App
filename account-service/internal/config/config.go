package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	GRPCPort int `env:"GRPC_PORT"`

	Gateway
	Postgres
	Redis
}

type Postgres struct {
	Host     string `env:"PG_HOST"`
	Port     int    `env:"PG_PORT"`
	Database string `env:"PGPG_DATABASE_DATA"`
	User     string `env:"PG_USER"`
	Password string `env:"PG_PASS"`
	MaxConn  int32  `env:"PG_MAXCONN"`
	MinConn  int32  `env:"PG_MINCONN"`
}

type Redis struct {
	Host     string `env:"REDIS_HOST"`
	Password string `env:"REDIS_PASS"`
	Port     int    `env:"REDIS_PORT"`
}

type Gateway struct {
	Port int `env:"GATEWAY_PORT"`
}

func NewConfig() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Println(err)
	}

	return &cfg, nil
}
