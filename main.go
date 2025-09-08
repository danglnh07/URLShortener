package main

import (
	"database/sql"
	"log/slog"
	"os"

	_ "github.com/lib/pq"

	"github.com/danglnh07/URLShortener/backend"
	"github.com/danglnh07/URLShortener/service"
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Load config
	err := service.LoadConfig(".env")
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}
	config := service.GetConfig()

	// Connect to database
	conn, err := sql.Open(config.DbDriver, config.DbSource)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Initialize and start server
	server := backend.NewServer(&config, conn, logger)
	err = server.Start()
	if err != nil {
		logger.Error("Server failed to start or unexpectedly shutdown", "error", err)
		os.Exit(1)
	}
}
