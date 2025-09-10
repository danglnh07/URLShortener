package service

import (
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
func LoadConfig(path string) error {
	// Load .env file
	err := godotenv.Load(path)
	if err != nil {
		return err
	}

	// Get and parse max request
	maxRequest, err := strconv.Atoi(os.Getenv("MAX_REQUEST"))
	if err != nil {
		return err
	}

	// Get and parse refill rate
	refileRate, err := strconv.Atoi(os.Getenv("REFILL_RATE"))
	if err != nil {
		return err
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
