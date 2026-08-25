package store

import (
	"github.com/nfeldt/inhale-with-me/internal/model"
	"gorm.io/gorm"
)

// GetCostSettings returns all per-type unit cost settings for a user.
func (s *Store) GetCostSettings(userID string) ([]model.CostSetting, error) {
	var cs []model.CostSetting
	err := s.db.Where("user_id = ?", userID).Order("type ASC").Find(&cs).Error
	return cs, err
}

// ReplaceCostSettings atomically replaces a user's full set of cost settings.
func (s *Store) ReplaceCostSettings(userID string, settings []model.CostSetting) error {
	return translate(s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.CostSetting{}).Error; err != nil {
			return err
		}
		if len(settings) == 0 {
			return nil
		}
		for i := range settings {
			settings[i].UserID = userID
		}
		return tx.Create(&settings).Error
	}))
}

// CostMap returns a user's unit costs keyed by session type.
func (s *Store) CostMap(userID string) (map[model.SessionType]int64, error) {
	cs, err := s.GetCostSettings(userID)
	if err != nil {
		return nil, err
	}
	m := make(map[model.SessionType]int64, len(cs))
	for _, c := range cs {
		m[c.Type] = c.UnitCostCents
	}
	return m, nil
}
