package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Postgres
	Rest
	Redis
	GrpcConfig
}

type Postgres struct {
	Host     string `env:"PG_HOST"`
	Port     int    `env:"PG_PORT"`
	Database string `env:"PG_DATABASE_DATA"`
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

type Rest struct {
	Port int `env:"REST_PORT_INP"`
}

type GrpcConfig struct {
	AccountClientAddr string `env:"ACCOUNT_CLIENTGRPCADDR"`
	PromoClientAddr   string `env:"PROMO_CLIENTGRPCADDR"`
}

func NewConfig() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Println(err)
	}

	return &cfg, nil
}
