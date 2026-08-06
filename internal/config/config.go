package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

type Config struct {
	DB  DatabaseConfig
	JWT JWTConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

func (dc DatabaseConfig) BuildDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dc.Host, dc.Port, dc.User, dc.Password, dc.DBName, dc.SSLMode,
	)
}

func Load() (*Config, error) {
	cfg := &Config{
		DB: DatabaseConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5432"),
			User:     os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			DBName:   os.Getenv("POSTGRES_DB"),
			SSLMode:  getEnv("POSTGRES_SSL_MODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:     os.Getenv("JWT_SECRET"),
			Expiration: getEnvAsDuration("JWT_EXPIRATION", time.Hour),
		},
	}

	required := []struct {
		key   string
		value string
	}{
		{"POSTGRES_USER", cfg.DB.User},
		{"POSTGRES_PASSWORD", cfg.DB.Password},
		{"POSTGRES_DB", cfg.DB.DBName},
		{"JWT_SECRET", cfg.JWT.Secret},
	}
	var missing []string
	for _, r := range required {
		if r.value == "" {
			missing = append(missing, r.key)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}
	return cfg, nil
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	d, err := time.ParseDuration(os.Getenv(key))
	if err != nil {
		slog.Warn("failed to parse env as duration", "key", key, "value", os.Getenv(key))
		return defaultValue
	}
	return d
}

func getEnv(key, defaultValue string) string {
	if value, found := os.LookupEnv(key); found && value != "" {
		return value
	}
	return defaultValue
}
