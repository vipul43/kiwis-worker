package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL     string
	PollInterval    int // seconds
	MaxRetries      int
	ShutdownTimeout int // seconds
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists (ignore error in production)
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return &Config{
		DatabaseURL:     dbURL,
		PollInterval:    10, // poll every 10 seconds
		MaxRetries:      3,
		ShutdownTimeout: 30,
	}, nil
}
