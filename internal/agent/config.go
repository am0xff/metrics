package agent

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddr     string `env:"ADDRESS" envDefault:"localhost:8080"`
	PollInterval   int    `env:"POLL_INTERVAL" envDefault:"2"`
	ReportInterval int    `env:"REPORT_INTERVAL" envDefault:"10"`
}

func LoadConfig() (Config, error) {
	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	// флаги
	fAddr := flag.String("a", cfg.ServerAddr, "HTTP сервер адрес")
	fPoll := flag.Int("p", cfg.PollInterval, "Интервал опроса метрик (сек)")
	fReport := flag.Int("r", cfg.ReportInterval, "Интервал отправки метрик (сек)")
	flag.Parse()

	cfg.ServerAddr = *fAddr
	cfg.PollInterval = *fPoll
	cfg.ReportInterval = *fReport

	return cfg, nil
}
