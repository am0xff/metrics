package agent

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddr     string `env:"ADDRESS" envDefault:"localhost:8080"`
	PollInterval   int    `env:"POLL_INTERVAL" envDefault:"2"`
	ReportInterval int    `env:"REPORT_INTERVAL" envDefault:"10"`
	Key            string `env:"KEY" envDefault:""`
	RateLimit      int    `env:"RATE_LIMIT" envDefault:"1"`
	CryptoKey      string `env:"CRYPTO_KEY" envDefault:""`
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
	fKey := flag.String("k", cfg.Key, "HashSHA256 ключ")
	fRateLimit := flag.Int("l", cfg.RateLimit, "Количество одновременно исходящих запросов на сервер")
	fCryptoKey := flag.String("crypto-key", cfg.CryptoKey, "Путь к файлу с публичным ключом для шифрования")
	flag.Parse()

	cfg.ServerAddr = *fAddr
	cfg.PollInterval = *fPoll
	cfg.ReportInterval = *fReport
	cfg.Key = *fKey
	cfg.RateLimit = *fRateLimit
	cfg.CryptoKey = *fCryptoKey

	return cfg, nil
}
