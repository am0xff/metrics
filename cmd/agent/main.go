package main

import (
	"github.com/am0xff/metrics/internal/agent"
)

func main() {
	if err := agent.Run(); err != nil {
		panic(err)
	}
}
