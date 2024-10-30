package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	BindAddr string `yaml:"bindaddr"`
}

func MustLoad() *Config {
	var cfg Config
	err := cleanenv.ReadConfig("config/config.yaml", &cfg)
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	return &cfg
}