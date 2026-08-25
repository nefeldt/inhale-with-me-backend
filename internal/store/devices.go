package store

import (
	"errors"

	"github.com/nfeldt/inhale-with-me/internal/model"
	"gorm.io/gorm"
)

// RegisterDevice stores (or re-assigns) an APNs device token for a user. A token
// is unique, so re-registering an existing token moves it to the current user.
func (s *Store) RegisterDevice(userID, token, platform string) (*model.Device, error) {
	if platform == "" {
		platform = "ios"
	}
	var d model.Device
	err := s.db.Where("token = ?", token).First(&d).Error
	if err == nil {
		d.UserID = userID
		d.Platform = platform
		if e := s.db.Save(&d).Error; e != nil {
			return nil, e
		}
		return &d, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	d = model.Device{ID: newID(), UserID: userID, Token: token, Platform: platform}
	if e := s.db.Create(&d).Error; e != nil {
		return nil, translate(e)
	}
	return &d, nil
}

// DeleteDevice removes a user's device token.
func (s *Store) DeleteDevice(userID, token string) error {
	res := s.db.Where("user_id = ? AND token = ?", userID, token).Delete(&model.Device{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteDeviceByToken removes a device by token regardless of owner (used to
// prune tokens APNs reports as invalid).
func (s *Store) DeleteDeviceByToken(token string) error {
	return s.db.Where("token = ?", token).Delete(&model.Device{}).Error
}

// DevicesForUsers returns all device registrations for the given user ids.
func (s *Store) DevicesForUsers(userIDs []string) ([]model.Device, error) {
	if len(userIDs) == 0 {
		return []model.Device{}, nil
	}
	var out []model.Device
	err := s.db.Where("user_id IN ?", userIDs).Find(&out).Error
	return out, err
}
