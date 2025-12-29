package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	LinkUpUsername     string
	LinkUpPassword     string
	LinkUpRegion       string
	LinkUpTimeInterval int
	NightscoutURL      string
	NightscoutAPIToken string // This is the API_SECRET, not a hashed token
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{
		LinkUpUsername:     getEnvOrDefault("LINK_UP_USERNAME", ""),
		LinkUpPassword:     getEnvOrDefault("LINK_UP_PASSWORD", ""),
		LinkUpRegion:       getEnvOrDefault("LINK_UP_REGION", "EU"),
		NightscoutURL:      getEnvOrDefault("NIGHTSCOUT_URL", ""),
		NightscoutAPIToken: getEnvOrDefault("NIGHTSCOUT_API_TOKEN", ""),
	}

	// Parse interval with default
	intervalStr := getEnvOrDefault("LINK_UP_TIME_INTERVAL", "1")
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		return nil, fmt.Errorf("invalid LINK_UP_TIME_INTERVAL: %v", err)
	}
	cfg.LinkUpTimeInterval = interval

	return cfg, nil
}

// Validate checks if all required configuration is present
func (c *Config) Validate() error {
	if c.LinkUpUsername == "" {
		return fmt.Errorf("LINK_UP_USERNAME is required")
	}
	if c.LinkUpPassword == "" {
		return fmt.Errorf("LINK_UP_PASSWORD is required")
	}
	if c.NightscoutURL == "" {
		return fmt.Errorf("NIGHTSCOUT_URL is required")
	}
	if c.NightscoutAPIToken == "" {
		return fmt.Errorf("NIGHTSCOUT_API_TOKEN is required")
	}
	if c.LinkUpTimeInterval < 1 {
		return fmt.Errorf("LINK_UP_TIME_INTERVAL must be at least 1 minute")
	}
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
