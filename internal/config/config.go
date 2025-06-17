package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"time"
)

type Config struct {
	Env         string        `yaml:"env" env-default:"local"`
	AccessTTL   time.Duration `yaml:"access_ttl" env-required:"true"`
	RefreshTTL  time.Duration `yaml:"refresh_ttl" env-required:"true"`
	PostgresDSN string        `yaml:"postgres-dsn" env-required:"true"`
	JWTSecret   string        `yaml:"jwt_secret" env-required:"true"`
	LogLevel    string        `yaml:"log_level" env-required:"true"`
	GRPC        GRPCConfig    `yaml:"grpc" env-required:"true"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-required:"true"`
}

func MustLoad() *Config {
	var cfg Config
	// TODO: change to .env file or add path init by flag --config
	err := cleanenv.ReadConfig("C:/github.com/AronditFire/User-Service/config/local.yaml", &cfg)
	if err != nil {
		panic(err)
	}

	return &cfg
}
