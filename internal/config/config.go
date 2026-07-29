package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
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
	Expiration int64
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
			Expiration: getEnvAsInt64("JWT_EXPIRATION", 86400000),
		},
	}

	if cfg.JWT.Secret == "" {
		return nil, errors.New("JWT_SECRET is required")
	}
	if cfg.DB.User == "" || cfg.DB.Password == "" || cfg.DB.DBName == "" {
		return nil, errors.New("database env vars are required")
	}
	return cfg, nil
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	num, err := strconv.ParseInt(os.Getenv(key), 10, 64)
	if err != nil {
		slog.Warn("failed to parse env as int64", "key", key, "value", os.Getenv(key))
		return defaultValue
	}
	return num
}

func getEnv(key, defaultValue string) string {
	if value, found := os.LookupEnv(key); found && value != "" {
		return value
	}
	return defaultValue
}
