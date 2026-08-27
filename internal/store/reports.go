package store

import "github.com/nfeldt/inhale-with-me/internal/model"

// ListReports returns filed reports, newest first (for the list-reports CLI).
func (s *Store) ListReports(limit int) ([]model.Report, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var out []model.Report
	err := s.db.Order("created_at DESC").Limit(limit).Find(&out).Error
	return out, err
}

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
