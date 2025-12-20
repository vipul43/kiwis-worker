package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL        string
	PollInterval       int // seconds
	MaxRetries         int
	ShutdownTimeout    int // seconds
	GoogleClientID     string
	GoogleClientSecret string
	OpenRouterAPIKey   string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists (ignore error in production)
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if googleClientID == "" || googleClientSecret == "" {
		fmt.Println("Warning: GOOGLE_CLIENT_ID or GOOGLE_CLIENT_SECRET not set, Gmail API will not work")
	}

	openRouterAPIKey := os.Getenv("OPENROUTER_API_KEY")
	if openRouterAPIKey == "" {
		fmt.Println("Warning: OPENROUTER_API_KEY not set, LLM payment extraction will not work")
	}

	return &Config{
		DatabaseURL:        dbURL,
		PollInterval:       10, // poll every 10 seconds
		MaxRetries:         3,
		ShutdownTimeout:    30,
		GoogleClientID:     googleClientID,
		GoogleClientSecret: googleClientSecret,
		OpenRouterAPIKey:   openRouterAPIKey,
	}, nil
}
