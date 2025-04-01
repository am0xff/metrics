package main

import "flag"

type options struct {
	addr           string
	reportInterval int
	pollInterval   int
}

// parseFlags разбирает аргументы командной строки и возвращает структуру с опциями.
func parseFlags() options {
	opt := options{}
	flag.StringVar(&opt.addr, "a", "http://localhost:8080", "address of the HTTP server")
	flag.IntVar(&opt.reportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&opt.pollInterval, "p", 2, "poll interval in seconds")
	flag.Parse()
	return opt
}
