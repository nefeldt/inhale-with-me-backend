package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const issuer = "inhale-with-me"

// ErrInvalidToken is returned when a token fails validation.
var ErrInvalidToken = errors.New("invalid token")

// Manager issues and verifies HS256 access tokens.
type Manager struct {
	secret []byte
	ttl    time.Duration
}

// NewManager creates a token manager with the given signing secret and lifetime.
func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl}
}

// Generate signs a new access token for userID and returns it with its expiry.
func (m *Manager) Generate(userID string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.ttl)
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    issuer,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		ID:        newID(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

// Parse validates a token string and returns the subject (user id).
func (m *Manager) Parse(tokenString string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method %v", ErrInvalidToken, t.Header["alg"])
		}
		return m.secret, nil
	}, jwt.WithIssuer(issuer), jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if claims.Subject == "" {
		return "", ErrInvalidToken
	}
	return claims.Subject, nil
}
