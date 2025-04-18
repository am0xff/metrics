package main

import (
	"github.com/am0xff/metrics/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		panic(err)
	}
}
