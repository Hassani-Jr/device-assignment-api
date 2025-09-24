package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// TLSConfig holds TLS/mTLS configuration
type TLSConfig struct {
	CertFile   string
	KeyFile    string
	CAFile     string
	RequireSSL bool
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	SecretKey     string
	TokenDuration time.Duration
	Issuer        string
}

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	TLS      TLSConfig
	JWT      JWTConfig
}

// Load reads configuration from environment variables with sensible defaults
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8443"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", "15s"),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", "15s"),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", "60s"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "device_assignment"),
			SSLMode:  getEnv("DB_SSL_MODE", "prefer"),
		},
		TLS: TLSConfig{
			CertFile:   getEnv("TLS_CERT_FILE", ""),
			KeyFile:    getEnv("TLS_KEY_FILE", ""),
			CAFile:     getEnv("TLS_CA_FILE", ""),
			RequireSSL: getBoolEnv("TLS_REQUIRE_SSL", true),
		},
		JWT: JWTConfig{
			SecretKey:     getEnv("JWT_SECRET_KEY", ""),
			TokenDuration: getDurationEnv("JWT_TOKEN_DURATION", "24h"),
			Issuer:        getEnv("JWT_ISSUER", "device-assignment-api"),
		},
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// validate ensures required configuration values are present
func (c *Config) validate() error {
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}

	if c.TLS.CertFile == "" || c.TLS.KeyFile == "" {
		return fmt.Errorf("TLS_CERT_FILE and TLS_KEY_FILE are required")
	}

	if c.TLS.CAFile == "" {
		return fmt.Errorf("TLS_CA_FILE is required for client certificate verification")
	}

	if c.JWT.SecretKey == "" {
		return fmt.Errorf("JWT_SECRET_KEY is required")
	}

	return nil
}

// ConnectionString returns the PostgreSQL connection string
func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return parsed
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue string) time.Duration {
	if value := os.Getenv(key); value != "" {
		parsed, err := time.ParseDuration(value)
		if err != nil {
			// If parsing fails, fall back to default
			parsed, _ = time.ParseDuration(defaultValue)
			return parsed
		}
		return parsed
	}
	parsed, _ := time.ParseDuration(defaultValue)
	return parsed
}
