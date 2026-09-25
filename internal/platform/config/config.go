package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
}

type AppConfig struct {
	Environment string
	Port        int
}

type DatabaseConfig struct {
	URL string
}

type JWTConfig struct {
	Secret string
	TTL    time.Duration
}

type CORSConfig struct {
	AllowedOrigins []string
}

func Load() (Config, error) {
	port, err := intEnv("APP_PORT", 8080)
	if err != nil {
		return Config{}, err
	}

	ttl, err := durationEnv("JWT_TTL", 24*time.Hour)
	if err != nil {
		return Config{}, err
	}

	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	return Config{
		App: AppConfig{
			Environment: envOrDefault("APP_ENV", "development"),
			Port:        port,
		},
		Database: DatabaseConfig{URL: databaseURL},
		JWT: JWTConfig{
			Secret: secret,
			TTL:    ttl,
		},
		CORS: CORSConfig{AllowedOrigins: csvEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")},
	}, nil
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func intEnv(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}

func csvEnv(key, fallback string) []string {
	values := strings.Split(envOrDefault(key, fallback), ",")
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}
