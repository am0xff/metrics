package server

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddr string `env:"ADDRESS" envDefault:"localhost:8080"`
}

func LoadConfig() (Config, error) {
	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	// флаги
	serverAddr := flag.String("a", cfg.ServerAddr, "HTTP сервер адрес")
	flag.Parse()

	cfg.ServerAddr = *serverAddr

	return cfg, nil
}
