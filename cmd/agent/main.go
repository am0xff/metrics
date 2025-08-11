package main

import (
	"fmt"
	"github.com/am0xff/metrics/internal/agent"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()

	if err := agent.Run(); err != nil {
		panic(err)
	}
}
func printBuildInfo() {
	fmt.Printf("Build version: %s\n", getBuildValue(buildVersion))
	fmt.Printf("Build date: %s\n", getBuildValue(buildDate))
	fmt.Printf("Build commit: %s\n", getBuildValue(buildCommit))
}

func getBuildValue(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}
