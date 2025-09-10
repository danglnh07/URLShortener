package service

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config struct to hold environment variables
type Config struct {
	// Server config
	BaseURL string

	// Database config
	DbDriver string
	DbSource string

	// Rate limiter config
	MaxRequest int
	RefillRate time.Duration
}

var config Config

// Load global variable to hold the configuration
func LoadConfig(path string, logger *slog.Logger) error {
	// Load .env file
	err := godotenv.Load(path)
	if err != nil {
		logger.Warn("Found no .env file. Start using default configuration", "error", err)

		// This value is necessary to connect to database, cannot really have a default value
		if os.Getenv("DB_SOURCE") == "" {
			logger.Error("Found no value for DB_SOURCE")
			return fmt.Errorf("no value for DB_SOURCE, cannot connect to database")
		}

		config = Config{
			BaseURL:    "localhost:8080",
			DbDriver:   "postgres",
			DbSource:   os.Getenv("DB_SOURCE"),
			MaxRequest: 100,
			RefillRate: 10 * time.Second,
		}
	}

	// Get and parse max request
	maxRequest, err := strconv.Atoi(os.Getenv("MAX_REQUEST"))
	if err != nil {
		logger.Warn("Invalid value for MAX_REQUEST. Start using default value", "error", err)
		maxRequest = 100
	}

	// Get and parse refill rate
	refileRate, err := strconv.Atoi(os.Getenv("REFILL_RATE"))
	if err != nil {
		logger.Warn("Invalid value for REFILL_RATE. Start using default value", "error", err)
		refileRate = 10
	}

	config = Config{
		BaseURL:    os.Getenv("BASE_URL"),
		DbDriver:   os.Getenv("DB_DRIVER"),
		DbSource:   os.Getenv("DB_SOURCE"),
		MaxRequest: maxRequest,
		RefillRate: time.Duration(refileRate) * time.Second,
	}
	return err
}

// Method to get the configuration
func GetConfig() Config {
	return config
}
