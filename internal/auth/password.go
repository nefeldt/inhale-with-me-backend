// Package auth provides password hashing, JWT issuance/verification and the
// HTTP middleware that authenticates requests.
package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword returns a bcrypt hash of the plaintext password at the given cost.
func HashPassword(password string, cost int) (string, error) {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	b, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CheckPassword reports whether password matches the stored bcrypt hash.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
