package service

import (
	"os"

	"github.com/joho/godotenv"
)

// Config struct to hold environment variables
type Config struct {
	// Server config
	Domain string
	Port   string

	// Database config
	DbDriver string
	DbSource string
}

var config Config

// Load global variable to hold the configuration
func LoadConfig(path string) error {
	// Load .env file
	err := godotenv.Load(path)
	if err != nil {
		return err
	}

	config = Config{
		Domain:   os.Getenv("DOMAIN"),
		Port:     os.Getenv("PORT"),
		DbDriver: os.Getenv("DB_DRIVER"),
		DbSource: os.Getenv("DB_SOURCE"),
	}
	return err
}

// Method to get the configuration
func GetConfig() Config {
	return config
}
