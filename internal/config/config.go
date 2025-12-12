package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	BotToken    string
	BotPassword string
	Database    DatabaseConfig
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file (ignore error if not exists)
	_ = godotenv.Load()

	cfg := &Config{
		BotToken:    os.Getenv("BOT_TOKEN"),
		BotPassword: os.Getenv("BOT_PASSWORD"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Name:     getEnv("DB_NAME", "languager"),
			User:     getEnv("DB_USER", "languager"),
			Password: os.Getenv("DB_PASSWORD"),
		},
	}

	// Validate required fields
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("BOT_TOKEN is required")
	}
	if cfg.BotPassword == "" {
		return nil, fmt.Errorf("BOT_PASSWORD is required")
	}
	if cfg.Database.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}

	return cfg, nil
}

// DSN returns PostgreSQL connection string
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

