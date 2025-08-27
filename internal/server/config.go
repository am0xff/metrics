package server

import (
	"encoding/json"
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddr      string `env:"ADDRESS" envDefault:":8080"`
	StoreInterval   int    `env:"STORE_INTERVAL" envDefault:"300"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"storage_file"`
	Restore         bool   `env:"RESTORE" envDefault:"false"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	Key             string `env:"KEY" envDefault:""`
	PprofEnabled    bool   `env:"PPROF_ENABLED" envDefault:"true"`
	PprofAddr       string `env:"PPROF_PORT" envDefault:":6060"`
	CryptoKey       string `env:"CRYPTO_KEY" envDefault:""`
	ConfigFile      string `env:"CONFIG" envDefault:""`
	TrustedSubnet   string `env:"TRUSTED_SUBNET" envDefault:""`
	Protocol        string `env:"PROTOCOL" envDefault:"http"`
	GRPCAddr        string `env:"GRPC_ADDR" envDefault:":9090"`
}

func LoadConfig() (Config, error) {
	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	// флаги
	serverAddr := flag.String("a", cfg.ServerAddr, "HTTP сервер адрес")
	storeInterval := flag.Int("i", cfg.StoreInterval, "Пауза/Интервал между сохранениями")
	fileStoragePath := flag.String("f", cfg.FileStoragePath, "Путь до файла, куда сохраняются текущие значения")
	restore := flag.Bool("r", cfg.Restore, "Загружать или нет ранее сохранённые значения")
	databaseDSN := flag.String("d", cfg.DatabaseDSN, "DNS database host")
	fKey := flag.String("k", cfg.Key, "HashSHA256 ключ")
	pprofEnabled := flag.Bool("pe", cfg.PprofEnabled, "pprof Enabled")
	pprofAddr := flag.String("pp", cfg.PprofAddr, "pprof address")
	fCryptoKey := flag.String("crypto-key", cfg.CryptoKey, "Путь к файлу с приватным ключом для расшифровки")
	fConfigFile := flag.String("c", cfg.ConfigFile, "Путь к файлу конфигурации")
	trustedSubnet := flag.String("t", cfg.TrustedSubnet, "Доверенная подсеть в формате CIDR")
	fProtocol := flag.String("p", cfg.Protocol, "Имя протокола")
	fGRPCAddr := flag.String("g", cfg.GRPCAddr, "GRPC сервер адрес")
	flag.Parse()

	cfg.ServerAddr = *serverAddr
	cfg.StoreInterval = *storeInterval
	cfg.FileStoragePath = *fileStoragePath
	cfg.Restore = *restore
	cfg.DatabaseDSN = *databaseDSN
	cfg.Key = *fKey
	cfg.PprofEnabled = *pprofEnabled
	cfg.PprofAddr = *pprofAddr
	cfg.CryptoKey = *fCryptoKey
	cfg.ConfigFile = *fConfigFile
	cfg.TrustedSubnet = *trustedSubnet
	cfg.Protocol = *fProtocol
	cfg.GRPCAddr = *fGRPCAddr

	if *fConfigFile != "" && *fConfigFile != cfg.ConfigFile {
		tempCfg := cfg
		if err := loadFromJSON(*fConfigFile, &tempCfg); err != nil {
			return cfg, err
		}

		tempCfg.ServerAddr = *serverAddr
		tempCfg.StoreInterval = *storeInterval
		tempCfg.FileStoragePath = *fileStoragePath
		tempCfg.Restore = *restore
		tempCfg.DatabaseDSN = *databaseDSN
		tempCfg.Key = *fKey
		tempCfg.PprofEnabled = *pprofEnabled
		tempCfg.PprofAddr = *pprofAddr
		tempCfg.CryptoKey = *fCryptoKey
		tempCfg.ConfigFile = *fConfigFile
		tempCfg.TrustedSubnet = *trustedSubnet

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
		Address       string `json:"address"`
		Restore       *bool  `json:"restore"`
		StoreInterval string `json:"store_interval"`
		StoreFile     string `json:"store_file"`
		DatabaseDSN   string `json:"database_dsn"`
		CryptoKey     string `json:"crypto_key"`
		TrustedSubnet string `json:"trusted_subnet"`
	}

	if err := json.Unmarshal(data, &jsonConfig); err != nil {
		return err
	}

	if jsonConfig.Address != "" {
		cfg.ServerAddr = jsonConfig.Address
	}
	if jsonConfig.Restore != nil {
		cfg.Restore = *jsonConfig.Restore
	}
	if jsonConfig.StoreFile != "" {
		cfg.FileStoragePath = jsonConfig.StoreFile
	}
	if jsonConfig.DatabaseDSN != "" {
		cfg.DatabaseDSN = jsonConfig.DatabaseDSN
	}
	if jsonConfig.CryptoKey != "" {
		cfg.CryptoKey = jsonConfig.CryptoKey
	}
	if jsonConfig.StoreInterval != "" {
		if duration, err := time.ParseDuration(jsonConfig.StoreInterval); err == nil {
			cfg.StoreInterval = int(duration.Seconds())
		}
	}
	if jsonConfig.TrustedSubnet != "" {
		cfg.TrustedSubnet = jsonConfig.TrustedSubnet
	}

	return nil
}
