package agent

import (
	"encoding/json"
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddr     string `env:"ADDRESS" envDefault:"localhost:8080"`
	PollInterval   int    `env:"POLL_INTERVAL" envDefault:"2"`
	ReportInterval int    `env:"REPORT_INTERVAL" envDefault:"10"`
	Key            string `env:"KEY" envDefault:""`
	RateLimit      int    `env:"RATE_LIMIT" envDefault:"1"`
	CryptoKey      string `env:"CRYPTO_KEY" envDefault:""`
	ConfigFile     string `env:"CONFIG" envDefault:""`
	Protocol       string `env:"PROTOCOL" envDefault:"http"`
	GRPCAddr       string `env:"GRPC_ADDR" envDefault:":9090"`
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
	fConfigFile := flag.String("c", cfg.ConfigFile, "Путь к файлу конфигурации")
	fProtocol := flag.String("p", cfg.Protocol, "Имя протокола")
	fGRPCAddr := flag.String("g", cfg.GRPCAddr, "GRPC сервер адрес")
	flag.Parse()

	cfg.ServerAddr = *fAddr
	cfg.PollInterval = *fPoll
	cfg.ReportInterval = *fReport
	cfg.Key = *fKey
	cfg.RateLimit = *fRateLimit
	cfg.CryptoKey = *fCryptoKey
	cfg.ConfigFile = *fConfigFile
	cfg.Protocol = *fProtocol
	cfg.GRPCAddr = *fGRPCAddr

	if *fConfigFile != "" && *fConfigFile != cfg.ConfigFile {
		tempCfg := cfg
		if err := loadFromJSON(*fConfigFile, &tempCfg); err != nil {
			return cfg, err
		}

		tempCfg.ServerAddr = *fAddr
		tempCfg.PollInterval = *fPoll
		tempCfg.ReportInterval = *fReport
		tempCfg.Key = *fKey
		tempCfg.RateLimit = *fRateLimit
		tempCfg.CryptoKey = *fCryptoKey
		tempCfg.ConfigFile = *fConfigFile

		cfg = tempCfg
	}

	return cfg, nil
}

func loadFromJSON(configPath string, cfg *Config) error {
	if configPath == "" {
		return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var jsonConfig struct {
		Address        string `json:"address"`
		ReportInterval string `json:"report_interval"`
		PollInterval   string `json:"poll_interval"`
		CryptoKey      string `json:"crypto_key"`
	}

	if err := json.Unmarshal(data, &jsonConfig); err != nil {
		return err
	}

	if jsonConfig.Address != "" {
		cfg.ServerAddr = jsonConfig.Address
	}
	if jsonConfig.CryptoKey != "" {
		cfg.CryptoKey = jsonConfig.CryptoKey
	}
	if jsonConfig.ReportInterval != "" {
		if duration, err := time.ParseDuration(jsonConfig.ReportInterval); err == nil {
			cfg.ReportInterval = int(duration.Seconds())
		}
	}
	if jsonConfig.PollInterval != "" {
		if duration, err := time.ParseDuration(jsonConfig.PollInterval); err == nil {
			cfg.PollInterval = int(duration.Seconds())
		}
	}

	return nil
}
