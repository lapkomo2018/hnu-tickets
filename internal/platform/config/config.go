package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// EnvironmentDevelopment is the environment in which the API regenerates its
// own specification on startup.
const EnvironmentDevelopment = "development"

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
	Admin    AdminConfig
}

type AppConfig struct {
	Environment string
	Port        int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

// URL builds the connection string, escaping credentials so passwords with
// reserved characters such as @ or / stay valid.
func (c DatabaseConfig) URL() string {
	return (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(c.User, c.Password),
		Host:     net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		Path:     "/" + c.Name,
		RawQuery: url.Values{"sslmode": {c.SSLMode}}.Encode(),
	}).String()
}

type JWTConfig struct {
	Secret string
	TTL    time.Duration
}

type CORSConfig struct {
	AllowedOrigins []string
}

// AdminConfig describes the administrator created on startup. It is optional:
// without an email no account is created.
type AdminConfig struct {
	Name     string
	Email    string
	Password string
}

func (c AdminConfig) Enabled() bool {
	return c.Email != ""
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

	database, err := loadDatabase()
	if err != nil {
		return Config{}, err
	}

	admin := AdminConfig{
		Name:     envOrDefault("ADMIN_NAME", "Administrator"),
		Email:    strings.TrimSpace(os.Getenv("ADMIN_EMAIL")),
		Password: os.Getenv("ADMIN_PASSWORD"),
	}
	if (admin.Email == "") != (admin.Password == "") {
		return Config{}, errors.New("ADMIN_EMAIL and ADMIN_PASSWORD must be set together")
	}

	return Config{
		App: AppConfig{
			Environment: envOrDefault("APP_ENV", EnvironmentDevelopment),
			Port:        port,
		},
		Database: database,
		JWT: JWTConfig{
			Secret: secret,
			TTL:    ttl,
		},
		CORS:  CORSConfig{AllowedOrigins: csvEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")},
		Admin: admin,
	}, nil
}

func loadDatabase() (DatabaseConfig, error) {
	port, err := intEnv("POSTGRES_PORT", 5432)
	if err != nil {
		return DatabaseConfig{}, err
	}
	database := DatabaseConfig{
		Host:     envOrDefault("POSTGRES_HOST", "localhost"),
		Port:     port,
		User:     strings.TrimSpace(os.Getenv("POSTGRES_USER")),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Name:     strings.TrimSpace(os.Getenv("POSTGRES_DB")),
		SSLMode:  envOrDefault("POSTGRES_SSLMODE", "disable"),
	}
	for _, required := range []struct{ key, value string }{
		{"POSTGRES_USER", database.User},
		{"POSTGRES_PASSWORD", database.Password},
		{"POSTGRES_DB", database.Name},
	} {
		if required.value == "" {
			return DatabaseConfig{}, fmt.Errorf("%s is required", required.key)
		}
	}
	return database, nil
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
