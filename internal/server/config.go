package server

type Config struct {
	ServerAddr string `env:"ADDRESS" envDefault:"localhost:8080"`
}
