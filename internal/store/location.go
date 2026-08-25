package store

import (
	"time"

	"github.com/nfeldt/inhale-with-me/internal/model"
)

// CreateLocationShare stores a one-to-one, time-limited location share.
func (s *Store) CreateLocationShare(ls *model.LocationShare) error {
	if ls.ID == "" {
		ls.ID = newID()
	}
	return s.db.Create(ls).Error
}

// ActiveLocationSharesForRecipient returns non-expired shares sent to the user.
func (s *Store) ActiveLocationSharesForRecipient(recipientID string) ([]model.LocationShare, error) {
	var out []model.LocationShare
	err := s.db.Where("recipient_id = ? AND expires_at > ?", recipientID, time.Now().UTC()).
		Order("created_at DESC").Find(&out).Error
	return out, err
}
