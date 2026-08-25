package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/nfeldt/inhale-with-me/internal/api"
	"github.com/nfeldt/inhale-with-me/internal/auth"
	"github.com/nfeldt/inhale-with-me/internal/config"
	"github.com/nfeldt/inhale-with-me/internal/database"
	"github.com/nfeldt/inhale-with-me/internal/store"
)

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "test.db"), false)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	cfg := config.Config{
		JWTSecret:          "test-secret",
		TokenTTL:           time.Hour,
		BcryptCost:         4,
		CORSAllowedOrigins: []string{"*"},
	}
	h := api.New(store.New(db), auth.NewManager(cfg.JWTSecret, cfg.TokenTTL), cfg).Router()
	return httptest.NewServer(h)
}

func doJSON(t *testing.T, method, url, token string, body any) (*http.Response, map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}
	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	var out map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	_ = resp.Body.Close()
	return resp, out
}

func register(t *testing.T, srv *httptest.Server, email, username string) string {
	t.Helper()
	resp, out := doJSON(t, http.MethodPost, srv.URL+"/api/v1/auth/register", "", map[string]any{
		"email": email, "username": username, "password": "password1234",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register %s: status %d body %v", username, resp.StatusCode, out)
	}
	token, _ := out["token"].(string)
	if token == "" {
		t.Fatalf("register %s: no token in %v", username, out)
	}
	return token
}

func TestAuthSessionAndStats(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	token := register(t, srv, "alice@example.com", "alice")

	// Duplicate registration is a conflict.
	if resp, _ := doJSON(t, http.MethodPost, srv.URL+"/api/v1/auth/register", "", map[string]any{
		"email": "alice@example.com", "username": "alice", "password": "password1234",
	}); resp.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate register: want 409 got %d", resp.StatusCode)
	}

	// Protected route without a token is unauthorized.
	if resp, _ := doJSON(t, http.MethodGet, srv.URL+"/api/v1/users/me", "", nil); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("no token: want 401 got %d", resp.StatusCode)
	}

	// Create a valid session.
	if resp, body := doJSON(t, http.MethodPost, srv.URL+"/api/v1/sessions", token, map[string]any{
		"type": "cigarette", "quantity": 1,
	}); resp.StatusCode != http.StatusCreated {
		t.Fatalf("create session: want 201 got %d body %v", resp.StatusCode, body)
	}

	// Invalid type is a validation error.
	if resp, _ := doJSON(t, http.MethodPost, srv.URL+"/api/v1/sessions", token, map[string]any{
		"type": "bong",
	}); resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("invalid type: want 422 got %d", resp.StatusCode)
	}

	// Stats reflect the one session.
	resp, stats := doJSON(t, http.MethodGet, srv.URL+"/api/v1/stats/summary?tz=UTC", token, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("stats: want 200 got %d", resp.StatusCode)
	}
	today, _ := stats["today"].(map[string]any)
	if today == nil || today["total_count"].(float64) != 1 {
		t.Fatalf("today total_count: want 1 got %v", stats["today"])
	}
}

func TestPrivateSessionHiddenFromOthers(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()

	alice := register(t, srv, "alice@example.com", "alice")
	bob := register(t, srv, "bob@example.com", "bob")

	_, sess := doJSON(t, http.MethodPost, srv.URL+"/api/v1/sessions", alice, map[string]any{
		"type": "vape", "visibility": "private",
	})
	id, _ := sess["id"].(string)
	if id == "" {
		t.Fatalf("no session id in %v", sess)
	}

	// Bob cannot see Alice's private session; the API returns 404 (not 403) to
	// avoid leaking its existence.
	if resp, _ := doJSON(t, http.MethodGet, srv.URL+"/api/v1/sessions/"+id, bob, nil); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("private session for other user: want 404 got %d", resp.StatusCode)
	}
}
