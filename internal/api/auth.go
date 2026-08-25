package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/nfeldt/inhale-with-me/internal/auth"
	"github.com/nfeldt/inhale-with-me/internal/model"
	"github.com/nfeldt/inhale-with-me/internal/store"
)

type registerRequest struct {
	Email       string  `json:"email"`
	Username    string  `json:"username"`
	Password    string  `json:"password"`
	DisplayName *string `json:"display_name"`
}

func (a *API) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body", nil)
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	username := strings.ToLower(strings.TrimSpace(req.Username))

	fields := map[string]string{}
	if !validEmail(email) {
		fields["email"] = "a valid email is required"
	}
	if len(username) < 3 || len(username) > 30 || !validUsername(username) {
		fields["username"] = "username must be 3-30 characters using letters, numbers, . _ -"
	}
	if len(req.Password) < 8 {
		fields["password"] = "password must be at least 8 characters"
	}
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}

	hash, err := auth.HashPassword(req.Password, a.cfg.BcryptCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "could not hash password", nil)
		return
	}

	u := &model.User{
		Email:        email,
		Username:     username,
		PasswordHash: hash,
		DisplayName:  cleanPtr(req.DisplayName),
	}
	if err := a.store.CreateUser(u); err != nil {
		if errors.Is(err, store.ErrConflict) {
			writeError(w, http.StatusConflict, "conflict", "that email or username is already in use", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "could not create account", nil)
		return
	}

	a.issueToken(w, http.StatusCreated, u)
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (a *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body", nil)
		return
	}

	u, err := a.store.GetUserByLogin(req.Login)
	if err != nil || !auth.CheckPassword(u.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "invalid login or password", nil)
		return
	}

	a.issueToken(w, http.StatusOK, u)
}

// issueToken generates a JWT for u and writes the standard auth response.
func (a *API) issueToken(w http.ResponseWriter, status int, u *model.User) {
	token, expiresAt, err := a.tokens.Generate(u.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal", "could not issue token", nil)
		return
	}
	writeJSON(w, status, map[string]any{
		"token":      token,
		"expires_at": expiresAt,
		"user":       u,
	})
}
