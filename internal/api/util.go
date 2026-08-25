package api

import (
	"encoding/base64"
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

// encodeCursor builds an opaque keyset pagination cursor from a row's
// (occurred_at, id). Clients treat it as opaque and pass it back as `before`.
func encodeCursor(t time.Time, id string) string {
	raw := t.UTC().Format(time.RFC3339Nano) + "|" + id
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// decodeCursor parses a cursor produced by encodeCursor into (occurred_at, id).
// It also accepts a bare RFC3339 timestamp (id empty) for manual use. ok is
// false when the value is absent or unparseable.
func decodeCursor(raw string) (t time.Time, id string, ok bool) {
	if raw == "" {
		return time.Time{}, "", false
	}
	if decoded, err := base64.RawURLEncoding.DecodeString(raw); err == nil {
		before, rest, found := strings.Cut(string(decoded), "|")
		if parsed, err := time.Parse(time.RFC3339Nano, before); err == nil {
			if found {
				id = rest
			}
			return parsed, id, true
		}
	}
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return parsed, "", true
	}
	return time.Time{}, "", false
}
