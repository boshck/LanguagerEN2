package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		setEnv       bool
		envValue     string
		expected     string
	}{
		{
			name:         "env variable set",
			key:          "TEST_KEY",
			defaultValue: "default",
			setEnv:       true,
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "env variable not set",
			key:          "TEST_KEY_NOT_SET",
			defaultValue: "default",
			setEnv:       false,
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfig_DSN(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "testuser",
			Password: "testpass",
			Name:     "testdb",
		},
	}

	dsn := cfg.DSN()
	expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	assert.Equal(t, expected, dsn)
}

func TestLoad_MissingRequiredFields(t *testing.T) {
	// Save original env
	originalBotToken := os.Getenv("BOT_TOKEN")
	originalBotPassword := os.Getenv("BOT_PASSWORD")
	originalDBPassword := os.Getenv("DB_PASSWORD")

	// Clean up after test
	defer func() {
		if originalBotToken != "" {
			os.Setenv("BOT_TOKEN", originalBotToken)
		} else {
			os.Unsetenv("BOT_TOKEN")
		}
		if originalBotPassword != "" {
			os.Setenv("BOT_PASSWORD", originalBotPassword)
		} else {
			os.Unsetenv("BOT_PASSWORD")
		}
		if originalDBPassword != "" {
			os.Setenv("DB_PASSWORD", originalDBPassword)
		} else {
			os.Unsetenv("DB_PASSWORD")
		}
	}()

	// Test missing BOT_TOKEN
	os.Unsetenv("BOT_TOKEN")
	os.Unsetenv("BOT_PASSWORD")
	os.Unsetenv("DB_PASSWORD")

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "BOT_TOKEN")
}

func TestLoad_WithDefaults(t *testing.T) {
	// Save original env
	originalBotToken := os.Getenv("BOT_TOKEN")
	originalBotPassword := os.Getenv("BOT_PASSWORD")
	originalDBPassword := os.Getenv("DB_PASSWORD")
	originalDBHost := os.Getenv("DB_HOST")
	originalDBPort := os.Getenv("DB_PORT")
	originalDBName := os.Getenv("DB_NAME")
	originalDBUser := os.Getenv("DB_USER")

	// Clean up after test
	defer func() {
		if originalBotToken != "" {
			os.Setenv("BOT_TOKEN", originalBotToken)
		}
		if originalBotPassword != "" {
			os.Setenv("BOT_PASSWORD", originalBotPassword)
		}
		if originalDBPassword != "" {
			os.Setenv("DB_PASSWORD", originalDBPassword)
		}
		if originalDBHost != "" {
			os.Setenv("DB_HOST", originalDBHost)
		} else {
			os.Unsetenv("DB_HOST")
		}
		if originalDBPort != "" {
			os.Setenv("DB_PORT", originalDBPort)
		} else {
			os.Unsetenv("DB_PORT")
		}
		if originalDBName != "" {
			os.Setenv("DB_NAME", originalDBName)
		} else {
			os.Unsetenv("DB_NAME")
		}
		if originalDBUser != "" {
			os.Setenv("DB_USER", originalDBUser)
		} else {
			os.Unsetenv("DB_USER")
		}
	}()

	// Set required fields
	os.Setenv("BOT_TOKEN", "test_token")
	os.Setenv("BOT_PASSWORD", "test_password")
	os.Setenv("DB_PASSWORD", "test_db_password")

	// Unset optional fields to test defaults
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_USER")

	cfg, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "test_token", cfg.BotToken)
	assert.Equal(t, "test_password", cfg.BotPassword)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "5432", cfg.Database.Port)
	assert.Equal(t, "languager", cfg.Database.Name)
	assert.Equal(t, "languager", cfg.Database.User)
}

func TestLoad_MissingBotPassword(t *testing.T) {
	// Save original env
	originalBotToken := os.Getenv("BOT_TOKEN")
	originalBotPassword := os.Getenv("BOT_PASSWORD")
	originalDBPassword := os.Getenv("DB_PASSWORD")

	// Clean up after test
	defer func() {
		if originalBotToken != "" {
			os.Setenv("BOT_TOKEN", originalBotToken)
		} else {
			os.Unsetenv("BOT_TOKEN")
		}
		if originalBotPassword != "" {
			os.Setenv("BOT_PASSWORD", originalBotPassword)
		} else {
			os.Unsetenv("BOT_PASSWORD")
		}
		if originalDBPassword != "" {
			os.Setenv("DB_PASSWORD", originalDBPassword)
		} else {
			os.Unsetenv("DB_PASSWORD")
		}
	}()

	// Test missing BOT_PASSWORD
	os.Setenv("BOT_TOKEN", "test_token")
	os.Unsetenv("BOT_PASSWORD")
	os.Setenv("DB_PASSWORD", "test_db_password")

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "BOT_PASSWORD")
}

func TestLoad_MissingDBPassword(t *testing.T) {
	// Save original env
	originalBotToken := os.Getenv("BOT_TOKEN")
	originalBotPassword := os.Getenv("BOT_PASSWORD")
	originalDBPassword := os.Getenv("DB_PASSWORD")

	// Clean up after test
	defer func() {
		if originalBotToken != "" {
			os.Setenv("BOT_TOKEN", originalBotToken)
		} else {
			os.Unsetenv("BOT_TOKEN")
		}
		if originalBotPassword != "" {
			os.Setenv("BOT_PASSWORD", originalBotPassword)
		} else {
			os.Unsetenv("BOT_PASSWORD")
		}
		if originalDBPassword != "" {
			os.Setenv("DB_PASSWORD", originalDBPassword)
		} else {
			os.Unsetenv("DB_PASSWORD")
		}
	}()

	// Test missing DB_PASSWORD
	os.Setenv("BOT_TOKEN", "test_token")
	os.Setenv("BOT_PASSWORD", "test_password")
	os.Unsetenv("DB_PASSWORD")

	cfg, err := Load()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "DB_PASSWORD")
}
