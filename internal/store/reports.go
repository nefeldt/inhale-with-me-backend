package store

import "github.com/nfeldt/inhale-with-me/internal/model"

// CreateReport stores a user's report of objectionable content for later review.
func (s *Store) CreateReport(reporterID, reportedUserID string, sessionID *string, reason string) (*model.Report, error) {
	r := &model.Report{
		ID:             newID(),
		ReporterID:     reporterID,
		ReportedUserID: reportedUserID,
		SessionID:      sessionID,
		Reason:         reason,
	}
	if err := s.db.Create(r).Error; err != nil {
		return nil, err
	}
	return r, nil
}
