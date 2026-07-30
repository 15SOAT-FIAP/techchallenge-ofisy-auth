package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseConfig_BuildDSN_Success(t *testing.T) {
	// Arrange
	dbConfig := DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "testuser",
		Password: "testpassword",
		DBName:   "testdb",
		SSLMode:  "disable",
	}
	expectedDSN := "host=localhost port=5432 user=testuser password=testpassword dbname=testdb sslmode=disable"

	// Act
	dsn := dbConfig.BuildDSN()

	// Assert
	assert.Equal(t, expectedDSN, dsn)
}

func TestLoad_Success(t *testing.T) {
	// Arrange
	t.Setenv("POSTGRES_HOST", "localhost")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("POSTGRES_USER", "testuser")
	t.Setenv("POSTGRES_PASSWORD", "testpassword")
	t.Setenv("POSTGRES_DB", "testdb")
	t.Setenv("POSTGRES_SSL_MODE", "disable")
	t.Setenv("JWT_SECRET", "testsecret")
	t.Setenv("JWT_EXPIRATION", "3600000")

	expected := &Config{
		DB: DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "testuser",
			Password: "testpassword",
			DBName:   "testdb",
			SSLMode:  "disable",
		},
		JWT: JWTConfig{
			Secret:     "testsecret",
			Expiration: 3600000,
		},
	}

	// Act
	cfg, err := Load()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expected, cfg)
}

func TestLoad_UsesDefaults_WhenOptionalVarsAreMissing(t *testing.T) {
	// Arrange
	t.Setenv("POSTGRES_USER", "testuser")
	t.Setenv("POSTGRES_PASSWORD", "testpassword")
	t.Setenv("POSTGRES_DB", "testdb")
	t.Setenv("JWT_SECRET", "testsecret")

	// Act
	cfg, err := Load()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "localhost", cfg.DB.Host)
	assert.Equal(t, "5432", cfg.DB.Port)
	assert.Equal(t, "disable", cfg.DB.SSLMode)
	assert.Equal(t, int64(86400000), cfg.JWT.Expiration)
}

func TestLoad_ReturnsError_WhenJWTSecretMissing(t *testing.T) {
	// Arrange
	t.Setenv("POSTGRES_USER", "testuser")
	t.Setenv("POSTGRES_PASSWORD", "testpassword")
	t.Setenv("POSTGRES_DB", "testdb")

	// Act
	cfg, err := Load()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoad_ReturnsError_WhenDatabaseVarsMissing(t *testing.T) {
	// Arrange
	t.Setenv("JWT_SECRET", "testsecret")

	// Act
	cfg, err := Load()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, cfg)
}
