package server

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddr      string `env:"ADDRESS" envDefault:"localhost:8080"`
	StoreInterval   int    `env:"STORE_INTERVAL" envDefault:"300"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"storage_file"`
	Restore         bool   `env:"RESTORE" envDefault:"false"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	Key             string `env:"KEY" envDefault:""`
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

	flag.Parse()

	cfg.ServerAddr = *serverAddr
	cfg.StoreInterval = *storeInterval
	cfg.FileStoragePath = *fileStoragePath
	cfg.Restore = *restore
	cfg.DatabaseDSN = *databaseDSN
	cfg.Key = *fKey

	return cfg, nil
}
