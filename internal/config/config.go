package config

import (
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)


type Config struct {
	Env     string `yaml:"env" env-default:"local"`
	Listen  struct {
		BindIP string `yaml:"bind_ip" env-default:"127.0.0.1"`
		Port   string `yaml:"port" env-default:"8080"`
		Timeout     time.Duration `yaml:"timeout" env-default:"4s"`	
		IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
		} `yaml:"listen"`
	Storage StorageConfig `yaml:"storage"`
	Auth AuthConfig `yaml:"auth"`
}

type StorageConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type AuthConfig struct {
	SigningKey string `yaml:"signing_key"`
	AccessTokenTTL time.Duration `yaml:"access_token_ttl"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl"`
}

func GetConfig(pathConfig string) *Config {
	var instance *Config
	var once sync.Once

	once.Do(func() {
		instance = &Config{}
		if err := cleanenv.ReadConfig(pathConfig, instance); err != nil {
			panic(err)
		}
	})
	return instance
}