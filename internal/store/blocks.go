package store

import "github.com/nfeldt/inhale-with-me/internal/model"

// BlockUser blocks blockedID for blockerID (idempotent) and severs any existing
// friendship/requests between them.
func (s *Store) BlockUser(blockerID, blockedID string) error {
	if blockerID == blockedID {
		return ErrConflict
	}
	var count int64
	if err := s.db.Model(&model.Block{}).
		Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil // already blocked
	}
	if err := s.db.Where(
		"(requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?)",
		blockerID, blockedID, blockedID, blockerID,
	).Delete(&model.Friendship{}).Error; err != nil {
		return err
	}
	b := &model.Block{ID: newID(), BlockerID: blockerID, BlockedID: blockedID}
	return translate(s.db.Create(b).Error)
}

// UnblockUser removes a block.
func (s *Store) UnblockUser(blockerID, blockedID string) error {
	res := s.db.Where("blocker_id = ? AND blocked_id = ?", blockerID, blockedID).Delete(&model.Block{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ListBlockedUsers returns the users blockerID has blocked.
func (s *Store) ListBlockedUsers(blockerID string) ([]model.User, error) {
	var blocks []model.Block
	if err := s.db.Where("blocker_id = ?", blockerID).Find(&blocks).Error; err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		return []model.User{}, nil
	}
	ids := make([]string, 0, len(blocks))
	for _, b := range blocks {
		ids = append(ids, b.BlockedID)
	}
	var users []model.User
	err := s.db.Where("id IN ?", ids).Order("username ASC").Find(&users).Error
	return users, err
}

// IsBlockedEitherWay reports whether a and b have blocked each other (in either
// direction).
func (s *Store) IsBlockedEitherWay(a, b string) (bool, error) {
	var count int64
	err := s.db.Model(&model.Block{}).Where(
		"(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)",
		a, b, b, a,
	).Count(&count).Error
	return count > 0, err
}
