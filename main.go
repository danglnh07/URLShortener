// @title URL Shortener API
// @version 1.0
// @description This is the API for URL Shortener service
// @host localhost:8080
// @BasePath /
package main

import (
	"database/sql"
	"log/slog"
	"os"

	_ "github.com/lib/pq"

	"github.com/danglnh07/URLShortener/api"
	"github.com/danglnh07/URLShortener/service"
)

func main() {
	// Initialize logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Load config
	err := service.LoadConfig(".env", logger)
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
	server := api.NewServer(&config, conn, logger)
	err = server.Start()
	if err != nil {
		logger.Error("Server failed to start or unexpectedly shutdown", "error", err)
		os.Exit(1)
	}
}
