package main

import (
	"log/slog"
	"os"

	"doozip/internal/logger"
)

func main() {
	log := logger.SetupLogger(os.Getenv("ENV"))
	slog.SetDefault(log)

	log.Info("application started",
		slog.String("version", "1.0.0"),
		slog.String("env", os.Getenv("ENV")),
	)

	
}
