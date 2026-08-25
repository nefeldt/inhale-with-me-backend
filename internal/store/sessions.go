package store

import (
	"time"

	"github.com/nfeldt/inhale-with-me/internal/model"
	"gorm.io/gorm"
)

// CreateSession inserts a new smoke session, assigning an ID if unset.
func (s *Store) CreateSession(sess *model.SmokeSession) error {
	if sess.ID == "" {
		sess.ID = newID()
	}
	if sess.OccurredAt.IsZero() {
		sess.OccurredAt = time.Now().UTC()
	}
	return translate(s.db.Create(sess).Error)
}

// GetSession returns a session by id, or ErrNotFound.
func (s *Store) GetSession(id string) (*model.SmokeSession, error) {
	var sess model.SmokeSession
	if err := s.db.First(&sess, "id = ?", id).Error; err != nil {
		return nil, translate(err)
	}
	return &sess, nil
}

// ListSessionsByUser returns a user's sessions newest-first, optionally filtered
// by type and paginated with a composite (occurred_at, id) cursor. beforeID may
// be empty for a timestamp-only cursor.
func (s *Store) ListSessionsByUser(userID string, before *time.Time, beforeID string, typ *model.SessionType, limit int) ([]model.SmokeSession, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	q := s.db.Where("user_id = ?", userID)
	q = applyBeforeCursor(q, before, beforeID)
	if typ != nil {
		q = q.Where("type = ?", *typ)
	}
	var out []model.SmokeSession
	err := q.Order("occurred_at DESC, id DESC").Limit(limit).Find(&out).Error
	return out, err
}

// applyBeforeCursor adds the keyset predicate matching the (occurred_at DESC,
// id DESC) ordering. The composite form is required so rows sharing a timestamp
// at a page boundary are neither skipped nor duplicated.
func applyBeforeCursor(q *gorm.DB, before *time.Time, beforeID string) *gorm.DB {
	if before == nil {
		return q
	}
	if beforeID == "" {
		return q.Where("occurred_at < ?", *before)
	}
	return q.Where("occurred_at < ? OR (occurred_at = ? AND id < ?)", *before, *before, beforeID)
}

// AllSessionsByUser returns every session for a user, newest-first. Used by the
// stats aggregation which computes buckets and streaks in Go.
func (s *Store) AllSessionsByUser(userID string) ([]model.SmokeSession, error) {
	var out []model.SmokeSession
	err := s.db.Where("user_id = ?", userID).Order("occurred_at DESC").Find(&out).Error
	return out, err
}

// UpdateSession persists all columns of sess.
func (s *Store) UpdateSession(sess *model.SmokeSession) error {
	return translate(s.db.Save(sess).Error)
}

// DeleteSession removes a session by id.
func (s *Store) DeleteSession(id string) error {
	res := s.db.Delete(&model.SmokeSession{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// CanView reports whether viewer may see sess given its visibility.
func (s *Store) CanView(viewerID string, sess *model.SmokeSession) (bool, error) {
	if sess.UserID == viewerID || sess.Visibility == model.VisibilityPublic {
		return true, nil
	}
	if sess.Visibility == model.VisibilityFriends {
		return s.AreFriends(viewerID, sess.UserID)
	}
	return false, nil
}

// FeedSessions returns friends' visible sessions newest-first for the feed,
// paginated with a composite (occurred_at, id) cursor.
func (s *Store) FeedSessions(friendIDs []string, before *time.Time, beforeID string, limit int) ([]model.SmokeSession, error) {
	if len(friendIDs) == 0 {
		return []model.SmokeSession{}, nil
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	q := s.db.Where("user_id IN ?", friendIDs).
		Where("visibility IN ?", []model.Visibility{model.VisibilityPublic, model.VisibilityFriends})
	q = applyBeforeCursor(q, before, beforeID)
	var out []model.SmokeSession
	err := q.Order("occurred_at DESC, id DESC").Limit(limit).Find(&out).Error
	return out, err
}
