// Package config loads runtime configuration from environment variables.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime configuration for the server.
type Config struct {
	Environment        string        // "development" | "production"
	Port               string        // TCP port the HTTP server listens on
	DBPath             string        // filesystem path to the SQLite database file
	JWTSecret          string        // signs and verifies access tokens
	TokenTTL           time.Duration // how long an issued access token stays valid
	BcryptCost         int           // bcrypt work factor for password hashing
	CORSAllowedOrigins []string      // origins permitted by CORS ("*" allows all)
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
}

// Load reads configuration from the environment, applying sensible defaults for
// local development. In development it first loads a local ".env" file (if
// present) so `go run ./cmd/server` works with no exported variables.
func Load() (Config, error) {
	env := getenv("ENVIRONMENT", "development")
	if env != "production" {
		loadDotEnv(".env")
	}

	ttl, err := time.ParseDuration(getenv("JWT_TTL", "720h")) // 30 days
	if err != nil {
		return Config{}, fmt.Errorf("invalid JWT_TTL: %w", err)
	}
	cost, err := strconv.Atoi(getenv("BCRYPT_COST", "12"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid BCRYPT_COST: %w", err)
	}

	cfg := Config{
		Environment:        env,
		Port:               getenv("PORT", "8080"),
		DBPath:             getenv("DB_PATH", "./data/inhale.db"),
		JWTSecret:          getenv("JWT_SECRET", ""),
		TokenTTL:           ttl,
		BcryptCost:         cost,
		CORSAllowedOrigins: splitAndTrim(getenv("CORS_ALLOWED_ORIGINS", "*")),
		ReadTimeout:        10 * time.Second,
		WriteTimeout:       15 * time.Second,
	}

	if cfg.JWTSecret == "" {
		if cfg.IsProduction() {
			return Config{}, fmt.Errorf("JWT_SECRET is required in production")
		}
		cfg.JWTSecret = "dev-insecure-secret-change-me"
	}

	return cfg, nil
}

// IsProduction reports whether the server runs in a production environment.
func (c Config) IsProduction() bool { return c.Environment == "production" }

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func splitAndTrim(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// loadDotEnv reads KEY=VALUE lines from a .env file and sets any variable that
// is not already present in the environment. It silently no-ops if the file is
// missing. This keeps a zero-dependency developer experience.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
}
