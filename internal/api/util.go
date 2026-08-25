package api

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	emailRe    = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	usernameRe = regexp.MustCompile(`^[a-z0-9._-]+$`)
)

func validEmail(s string) bool    { return len(s) <= 254 && emailRe.MatchString(s) }
func validUsername(s string) bool { return usernameRe.MatchString(s) }

// cleanPtr trims a string pointer, returning nil for empty/whitespace values.
func cleanPtr(p *string) *string {
	if p == nil {
		return nil
	}
	t := strings.TrimSpace(*p)
	if t == "" {
		return nil
	}
	return &t
}

// queryInt reads an integer query parameter with a default.
func queryInt(r *http.Request, key string, def int) int {
	if v := r.URL.Query().Get(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// queryTime parses an RFC3339 query parameter, returning nil when absent/invalid.
func queryTime(r *http.Request, key string) *time.Time {
	v := r.URL.Query().Get(key)
	if v == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return nil
	}
	return &t
}
