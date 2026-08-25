package store

import (
	"strings"

	"github.com/nfeldt/inhale-with-me/internal/model"
	"gorm.io/gorm"
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

// DeleteUser permanently deletes a user and all their associated data (App Store
// account-deletion requirement). Runs in one transaction.
func (s *Store) DeleteUser(id string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Reactions the user made anywhere.
		if err := tx.Where("user_id = ?", id).Delete(&model.Reaction{}).Error; err != nil {
			return err
		}
		// Reactions others made on the user's sessions.
		var sessionIDs []string
		if err := tx.Model(&model.SmokeSession{}).Where("user_id = ?", id).Pluck("id", &sessionIDs).Error; err != nil {
			return err
		}
		if len(sessionIDs) > 0 {
			if err := tx.Where("session_id IN ?", sessionIDs).Delete(&model.Reaction{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("user_id = ?", id).Delete(&model.SmokeSession{}).Error; err != nil {
			return err
		}
		if err := tx.Where("requester_id = ? OR addressee_id = ?", id, id).Delete(&model.Friendship{}).Error; err != nil {
			return err
		}
		if err := tx.Where("blocker_id = ? OR blocked_id = ?", id, id).Delete(&model.Block{}).Error; err != nil {
			return err
		}
		if err := tx.Where("reporter_id = ? OR reported_user_id = ?", id, id).Delete(&model.Report{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", id).Delete(&model.CostSetting{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", id).Delete(&model.Device{}).Error; err != nil {
			return err
		}
		res := tx.Delete(&model.User{}, "id = ?", id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}
