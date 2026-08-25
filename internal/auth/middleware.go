package auth

import (
	"net/http"
	"strings"
)

// Unauthorized is the hook the API layer sets so this package can write its
// standard error envelope without importing the api package.
var Unauthorized = func(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":{"code":"unauthorized","message":"authentication required"}}`, http.StatusUnauthorized)
	w.Header().Set("Content-Type", "application/json")
}

// Middleware returns a chi-compatible middleware that requires a valid bearer
// token and injects the user id into the request context.
func Middleware(m *Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || strings.TrimSpace(token) == "" {
				Unauthorized(w, r)
				return
			}
			userID, err := m.Parse(strings.TrimSpace(token))
			if err != nil {
				Unauthorized(w, r)
				return
			}
			next.ServeHTTP(w, r.WithContext(WithUserID(r.Context(), userID)))
		})
	}
}
