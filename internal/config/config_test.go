package config

import (
	"maps"
	"strings"
	"testing"
	"time"
)

var allEnvKeys = []string{
	"POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_USER",
	"POSTGRES_PASSWORD", "POSTGRES_DB", "POSTGRES_SSL_MODE",
	"JWT_SECRET", "JWT_EXPIRATION",
}

func TestLoad(t *testing.T) {
	baseEnv := map[string]string{
		"POSTGRES_USER":     "user",
		"POSTGRES_PASSWORD": "pass",
		"POSTGRES_DB":       "dbname",
		"JWT_SECRET":        "secret",
	}

	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{
			name: "carrega com sucesso aplicando defaults",
			env:  baseEnv,
		},
		{
			name:    "erro quando JWT_SECRET ausente",
			env:     withOverride(baseEnv, "JWT_SECRET", ""),
			wantErr: "missing required env vars: JWT_SECRET",
		},
		{
			name:    "erro quando POSTGRES_USER ausente",
			env:     withOverride(baseEnv, "POSTGRES_USER", ""),
			wantErr: "missing required env vars: POSTGRES_USER",
		},
		{
			name:    "erro quando POSTGRES_PASSWORD ausente",
			env:     withOverride(baseEnv, "POSTGRES_PASSWORD", ""),
			wantErr: "missing required env vars: POSTGRES_PASSWORD",
		},
		{
			name:    "erro quando POSTGRES_DB ausente",
			env:     withOverride(baseEnv, "POSTGRES_DB", ""),
			wantErr: "missing required env vars: POSTGRES_DB",
		},
		{
			name:    "erro lista todas as vars ausentes",
			env:     withOverride(withOverride(withOverride(baseEnv, "POSTGRES_USER", ""), "POSTGRES_DB", ""), "JWT_SECRET", ""),
			wantErr: "missing required env vars: POSTGRES_USER, POSTGRES_DB, JWT_SECRET",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnv(t, tt.env)

			cfg, err := Load()

			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("erro = %v, esperado conter %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			if cfg.DB.Host != "localhost" {
				t.Errorf("DB.Host = %q, esperado default %q", cfg.DB.Host, "localhost")
			}
			if cfg.DB.Port != "5432" {
				t.Errorf("DB.Port = %q, esperado default %q", cfg.DB.Port, "5432")
			}
			if cfg.DB.SSLMode != "disable" {
				t.Errorf("DB.SSLMode = %q, esperado default %q", cfg.DB.SSLMode, "disable")
			}
			if cfg.JWT.Expiration != time.Hour {
				t.Errorf("JWT.Expiration = %s, esperado default %s", cfg.JWT.Expiration, time.Hour)
			}
		})
	}
}

func TestLoad_CustomValues(t *testing.T) {
	setEnv(t, map[string]string{
		"POSTGRES_HOST":     "db.internal",
		"POSTGRES_PORT":     "6543",
		"POSTGRES_USER":     "user",
		"POSTGRES_PASSWORD": "pass",
		"POSTGRES_DB":       "dbname",
		"POSTGRES_SSL_MODE": "require",
		"JWT_SECRET":        "secret",
		"JWT_EXPIRATION":    "2h",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if cfg.DB.Host != "db.internal" {
		t.Errorf("DB.Host = %q, esperado %q", cfg.DB.Host, "db.internal")
	}
	if cfg.DB.Port != "6543" {
		t.Errorf("DB.Port = %q, esperado %q", cfg.DB.Port, "6543")
	}
	if cfg.DB.SSLMode != "require" {
		t.Errorf("DB.SSLMode = %q, esperado %q", cfg.DB.SSLMode, "require")
	}
	if cfg.JWT.Expiration != 2*time.Hour {
		t.Errorf("JWT.Expiration = %s, esperado %s", cfg.JWT.Expiration, 2*time.Hour)
	}
}

func TestDatabaseConfig_BuildDSN(t *testing.T) {
	dc := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "user",
		Password: "pass",
		DBName:   "dbname",
		SSLMode:  "disable",
	}

	want := "host=localhost port=5432 user=user password=pass dbname=dbname sslmode=disable"
	if got := dc.BuildDSN(); got != want {
		t.Errorf("BuildDSN() = %q, esperado %q", got, want)
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		envSet       bool
		defaultValue string
		want         string
	}{
		{
			name:         "retorna valor do ambiente quando definido",
			envValue:     "custom",
			envSet:       true,
			defaultValue: "default",
			want:         "custom",
		},
		{
			name:         "retorna default quando variavel nao definida",
			envSet:       false,
			defaultValue: "default",
			want:         "default",
		},
		{
			name:         "retorna default quando variavel vazia",
			envValue:     "",
			envSet:       true,
			defaultValue: "default",
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const key = "CONFIG_TEST_GETENV"
			if tt.envSet {
				t.Setenv(key, tt.envValue)
			}

			if got := getEnv(key, tt.defaultValue); got != tt.want {
				t.Errorf("getEnv() = %q, esperado %q", got, tt.want)
			}
		})
	}
}

func TestGetEnvAsDuration(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		envSet       bool
		defaultValue time.Duration
		want         time.Duration
	}{
		{
			name:         "retorna valor parseado quando definido",
			envValue:     "12h",
			envSet:       true,
			defaultValue: 999 * time.Second,
			want:         12 * time.Hour,
		},
		{
			name:         "retorna default quando variavel nao definida",
			envSet:       false,
			defaultValue: 999 * time.Second,
			want:         999 * time.Second,
		},
		{
			name:         "retorna default quando valor invalido",
			envValue:     "not-a-duration",
			envSet:       true,
			defaultValue: 999 * time.Second,
			want:         999 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const key = "CONFIG_TEST_GETENVASDURATION"
			if tt.envSet {
				t.Setenv(key, tt.envValue)
			}

			if got := getEnvAsDuration(key, tt.defaultValue); got != tt.want {
				t.Errorf("getEnvAsDuration() = %s, esperado %s", got, tt.want)
			}
		})
	}
}

func withOverride(base map[string]string, key, value string) map[string]string {
	env := make(map[string]string, len(base))
	maps.Copy(env, base)
	env[key] = value
	return env
}

func setEnv(t *testing.T, env map[string]string) {
	t.Helper()
	for _, key := range allEnvKeys {
		t.Setenv(key, env[key])
	}
}
