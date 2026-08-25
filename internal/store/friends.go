package store

import (
	"errors"

	"github.com/nfeldt/inhale-with-me/internal/model"
)

// GetFriendship returns a friendship row by id.
func (s *Store) GetFriendship(id string) (*model.Friendship, error) {
	var f model.Friendship
	if err := s.db.First(&f, "id = ?", id).Error; err != nil {
		return nil, translate(err)
	}
	return &f, nil
}

// GetFriendshipBetween returns the friendship row for the unordered pair (a,b).
func (s *Store) GetFriendshipBetween(a, b string) (*model.Friendship, error) {
	var f model.Friendship
	err := s.db.Where(
		"(requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?)",
		a, b, b, a,
	).First(&f).Error
	if err != nil {
		return nil, translate(err)
	}
	return &f, nil
}

// CreateRequest creates a pending friend request from requester to addressee. A
// pending/accepted relationship yields ErrConflict; a previously declined row is
// revived as a new pending request in the requested direction.
func (s *Store) CreateRequest(requesterID, addresseeID string) (*model.Friendship, error) {
	if requesterID == addresseeID {
		return nil, ErrConflict
	}
	existing, err := s.GetFriendshipBetween(requesterID, addresseeID)
	switch {
	case err == nil:
		switch existing.Status {
		case model.FriendshipPending, model.FriendshipAccepted:
			return nil, ErrConflict
		default: // declined -> revive
			existing.RequesterID = requesterID
			existing.AddresseeID = addresseeID
			existing.PairKey = pairKey(requesterID, addresseeID)
			existing.Status = model.FriendshipPending
			if err := s.db.Save(existing).Error; err != nil {
				return nil, err
			}
			return existing, nil
		}
	case errors.Is(err, ErrNotFound):
		f := &model.Friendship{
			ID:          newID(),
			RequesterID: requesterID,
			AddresseeID: addresseeID,
			PairKey:     pairKey(requesterID, addresseeID),
			Status:      model.FriendshipPending,
		}
		// The unique PairKey index turns a lost race (concurrent mutual request)
		// into ErrConflict instead of a duplicate row.
		if err := s.db.Create(f).Error; err != nil {
			return nil, translate(err)
		}
		return f, nil
	default:
		return nil, err
	}
}

// pairKey returns an order-independent key for a pair of user ids, so (A,B) and
// (B,A) map to the same value.
func pairKey(a, b string) string {
	if a <= b {
		return a + "|" + b
	}
	return b + "|" + a
}

// SetFriendshipStatus updates the status of a friendship row.
func (s *Store) SetFriendshipStatus(id string, status model.FriendshipStatus) (*model.Friendship, error) {
	f, err := s.GetFriendship(id)
	if err != nil {
		return nil, err
	}
	f.Status = status
	if err := s.db.Save(f).Error; err != nil {
		return nil, err
	}
	return f, nil
}

// DeleteFriendship removes a friendship row by id.
func (s *Store) DeleteFriendship(id string) error {
	res := s.db.Delete(&model.Friendship{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ListPendingRequests returns pending requests in the given direction for user.
// direction is "incoming" (requests to accept) or "outgoing" (sent requests).
func (s *Store) ListPendingRequests(userID, direction string) ([]model.Friendship, error) {
	q := s.db.Where("status = ?", model.FriendshipPending)
	if direction == "outgoing" {
		q = q.Where("requester_id = ?", userID)
	} else {
		q = q.Where("addressee_id = ?", userID)
	}
	var out []model.Friendship
	err := q.Order("created_at DESC").Find(&out).Error
	return out, err
}

// ListFriends returns the accepted-friend users of userID.
func (s *Store) ListFriends(userID string) ([]model.User, error) {
	ids, err := s.FriendIDs(userID)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []model.User{}, nil
	}
	var users []model.User
	if err := s.db.Where("id IN ?", ids).Order("username ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// FriendIDs returns the ids of userID's accepted friends.
func (s *Store) FriendIDs(userID string) ([]string, error) {
	var fs []model.Friendship
	if err := s.db.Where(
		"status = ? AND (requester_id = ? OR addressee_id = ?)",
		model.FriendshipAccepted, userID, userID,
	).Find(&fs).Error; err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(fs))
	for _, f := range fs {
		if f.RequesterID == userID {
			ids = append(ids, f.AddresseeID)
		} else {
			ids = append(ids, f.RequesterID)
		}
	}
	return ids, nil
}

// AreFriends reports whether a and b are accepted friends.
func (s *Store) AreFriends(a, b string) (bool, error) {
	var count int64
	err := s.db.Model(&model.Friendship{}).Where(
		"status = ? AND ((requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?))",
		model.FriendshipAccepted, a, b, b, a,
	).Count(&count).Error
	return count > 0, err
}

// RemoveFriend deletes the accepted friendship between userID and otherID.
func (s *Store) RemoveFriend(userID, otherID string) error {
	res := s.db.Where(
		"status = ? AND ((requester_id = ? AND addressee_id = ?) OR (requester_id = ? AND addressee_id = ?))",
		model.FriendshipAccepted, userID, otherID, otherID, userID,
	).Delete(&model.Friendship{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// FriendStatus returns the relationship of other relative to viewer:
// "none", "friends", "outgoing" (viewer requested) or "incoming" (other requested).
func (s *Store) FriendStatus(viewerID, otherID string) (string, error) {
	f, err := s.GetFriendshipBetween(viewerID, otherID)
	if errors.Is(err, ErrNotFound) {
		return model.FriendStatusNone, nil
	}
	if err != nil {
		return "", err
	}
	switch f.Status {
	case model.FriendshipAccepted:
		return model.FriendStatusFriends, nil
	case model.FriendshipPending:
		if f.RequesterID == viewerID {
			return model.FriendStatusOutgoing, nil
		}
		return model.FriendStatusIncoming, nil
	default:
		return model.FriendStatusNone, nil
	}
}
