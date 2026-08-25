package store

import (
	"errors"

	"github.com/nfeldt/inhale-with-me/internal/model"
	"gorm.io/gorm"
)

// AddReaction adds (idempotently) a reaction of typ by userID on sessionID.
// Re-adding the same reaction returns the existing row.
func (s *Store) AddReaction(sessionID, userID, typ string) (*model.Reaction, error) {
	if typ == "" {
		typ = model.DefaultReactionType
	}
	var r model.Reaction
	err := s.db.Where("session_id = ? AND user_id = ? AND type = ?", sessionID, userID, typ).First(&r).Error
	if err == nil {
		return &r, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	r = model.Reaction{ID: newID(), SessionID: sessionID, UserID: userID, Type: typ}
	if err := s.db.Create(&r).Error; err != nil {
		return nil, translate(err)
	}
	return &r, nil
}

// RemoveReaction removes userID's reaction of typ on sessionID.
func (s *Store) RemoveReaction(sessionID, userID, typ string) error {
	if typ == "" {
		typ = model.DefaultReactionType
	}
	res := s.db.Where("session_id = ? AND user_id = ? AND type = ?", sessionID, userID, typ).Delete(&model.Reaction{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ListReactions returns all reactions on a session.
func (s *Store) ListReactions(sessionID string) ([]model.Reaction, error) {
	var out []model.Reaction
	err := s.db.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&out).Error
	return out, err
}

// SummariesForSessions returns, for the given sessions, a reaction summary per
// session id and the viewer's own reaction types per session id.
func (s *Store) SummariesForSessions(sessionIDs []string, viewerID string) (map[string]model.ReactionSummary, map[string][]string, error) {
	summaries := make(map[string]model.ReactionSummary)
	mine := make(map[string][]string)
	if len(sessionIDs) == 0 {
		return summaries, mine, nil
	}
	var rs []model.Reaction
	if err := s.db.Where("session_id IN ?", sessionIDs).Find(&rs).Error; err != nil {
		return nil, nil, err
	}
	for _, r := range rs {
		sum := summaries[r.SessionID]
		if sum.Counts == nil {
			sum.Counts = make(map[string]int)
		}
		sum.Counts[r.Type]++
		sum.Total++
		summaries[r.SessionID] = sum
		if r.UserID == viewerID {
			mine[r.SessionID] = append(mine[r.SessionID], r.Type)
		}
	}
	return summaries, mine, nil
}
