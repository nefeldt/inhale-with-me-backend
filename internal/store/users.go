package store

import (
	"strings"

	"github.com/nfeldt/inhale-with-me/internal/model"
)

// CreateUser inserts a new user, assigning an ID if unset. Duplicate email or
// username yields ErrConflict.
func (s *Store) CreateUser(u *model.User) error {
	if u.ID == "" {
		u.ID = newID()
	}
	if u.Currency == "" {
		u.Currency = "EUR"
	}
	return translate(s.db.Create(u).Error)
}

// GetUserByID returns the user with the given id, or ErrNotFound.
func (s *Store) GetUserByID(id string) (*model.User, error) {
	var u model.User
	if err := s.db.First(&u, "id = ?", id).Error; err != nil {
		return nil, translate(err)
	}
	return &u, nil
}

// GetUserByUsername returns the user with the given (case-insensitive) username.
func (s *Store) GetUserByUsername(username string) (*model.User, error) {
	var u model.User
	username = strings.ToLower(strings.TrimSpace(username))
	if err := s.db.First(&u, "username = ?", username).Error; err != nil {
		return nil, translate(err)
	}
	return &u, nil
}

// GetUserByLogin looks a user up by email OR username (case-insensitive).
func (s *Store) GetUserByLogin(login string) (*model.User, error) {
	var u model.User
	login = strings.ToLower(strings.TrimSpace(login))
	if err := s.db.First(&u, "email = ? OR username = ?", login, login).Error; err != nil {
		return nil, translate(err)
	}
	return &u, nil
}

// UpdateUser persists all columns of u. Duplicate username yields ErrConflict.
func (s *Store) UpdateUser(u *model.User) error {
	return translate(s.db.Save(u).Error)
}

// SearchUsers returns up to limit users whose username or email starts with the
// query (case-insensitive). An empty query returns no rows.
func (s *Store) SearchUsers(query string, limit int) ([]model.User, error) {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return []model.User{}, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	// Match usernames only — emails are private (never returned in PublicUser),
	// so searching them would turn this into an email->account oracle. Escape
	// LIKE wildcards so a query like "%" can't enumerate every account.
	like := escapeLike(query) + "%"
	var users []model.User
	err := s.db.Where(`username LIKE ? ESCAPE '\'`, like).
		Order("username ASC").Limit(limit).Find(&users).Error
	return users, err
}

// escapeLike escapes the LIKE metacharacters so user input is matched literally
// (used with an ESCAPE '\' clause).
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
