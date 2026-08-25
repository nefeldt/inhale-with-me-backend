package store

import (
	"errors"

	"github.com/nfeldt/inhale-with-me/internal/model"
	"gorm.io/gorm"
)

// ListCustomTypes returns a user's custom session types.
func (s *Store) ListCustomTypes(userID string) ([]model.CustomType, error) {
	var out []model.CustomType
	err := s.db.Where("user_id = ?", userID).Order("created_at ASC").Find(&out).Error
	return out, err
}

// UpsertCustomType creates or updates a user's custom type (keyed by name).
func (s *Store) UpsertCustomType(userID, name, color string) (*model.CustomType, error) {
	var ct model.CustomType
	err := s.db.Where("user_id = ? AND name = ?", userID, name).First(&ct).Error
	if err == nil {
		ct.Color = color
		if e := s.db.Save(&ct).Error; e != nil {
			return nil, e
		}
		return &ct, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	ct = model.CustomType{UserID: userID, Name: name, Color: color}
	if e := s.db.Create(&ct).Error; e != nil {
		return nil, translate(e)
	}
	return &ct, nil
}

// DeleteCustomType removes a user's custom type.
func (s *Store) DeleteCustomType(userID, name string) error {
	res := s.db.Where("user_id = ? AND name = ?", userID, name).Delete(&model.CustomType{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
