package main

import (
	"os"

	"keyraccoon/internal/cli"
	"keyraccoon/pkg/logger"
)

func main() {
	logger.Init()

	if err := cli.Execute(); err != nil {
		logger.Fatal("cli execution failed", "error", err)
		os.Exit(1)
	}
}
