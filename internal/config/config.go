// Package config loads runtime configuration from environment variables.
package config

import (
	"bufio"
	"encoding/base64"
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

	// APNs push (optional). Push is enabled only when APNsKeyP8 + APNsKeyID +
	// APNsTeamID are all set; otherwise the server runs with a no-op notifier.
	APNsKeyP8      string // the .p8 signing key contents (PEM)
	APNsKeyID      string
	APNsTeamID     string
	APNsBundleID   string
	APNsProduction bool // true for TestFlight/App Store builds
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
		APNsKeyP8:          apnsKey(),
		APNsKeyID:          getenv("APNS_KEY_ID", ""),
		APNsTeamID:         getenv("APNS_TEAM_ID", ""),
		APNsBundleID:       getenv("APNS_BUNDLE_ID", "feldt.systems.Inhale-With-Me"),
		APNsProduction:     getenv("APNS_PRODUCTION", "false") == "true",
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

// apnsKey returns the .p8 signing key contents, preferring a base64-encoded
// value (APNS_KEY_P8_BASE64) since a multi-line PEM is awkward to store in an
// env var; falls back to the raw PEM in APNS_KEY_P8.
func apnsKey() string {
	if b64 := getenv("APNS_KEY_P8_BASE64", ""); b64 != "" {
		// Strip ALL whitespace: `base64` output is often wrapped across lines,
		// and StdEncoding rejects embedded newlines.
		clean := strings.NewReplacer("\n", "", "\r", "", "\t", "", " ", "").Replace(b64)
		if decoded, err := base64.StdEncoding.DecodeString(clean); err == nil {
			return string(decoded)
		}
	}
	return getenv("APNS_KEY_P8", "")
}

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
