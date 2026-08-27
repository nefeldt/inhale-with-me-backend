package api

import (
	"net/http"
	"time"

	"github.com/nfeldt/inhale-with-me/internal/model"
)

// presenceTTL is how long after a located session a user stays on their
// friends' map. After it, they drop off until they log another session.
const presenceTTL = 90 * time.Minute

// mapPin is one person visible on the caller's map: a friend who is actively
// smoking (logged a located session within the presence window).
type mapPin struct {
	User      model.PublicUser `json:"user"`
	Lat       float64          `json:"lat"`
	Lng       float64          `json:"lng"`
	Kind      string           `json:"kind"` // always "smoking"
	Type      *string          `json:"type"`
	Subtype   *string          `json:"subtype"`
	At        time.Time        `json:"at"`
	ExpiresAt time.Time        `json:"expires_at"`
}

// handleMap returns the friends currently on the caller's map: those who logged
// a located session within the presence window. At most one pin per person (the
// most recent session wins).
func (a *API) handleMap(w http.ResponseWriter, r *http.Request) {
	me := currentUserID(r)
	pins := map[string]mapPin{}

	friendIDs, err := a.store.FriendIDs(me)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if len(friendIDs) > 0 {
		since := time.Now().UTC().Add(-presenceTTL)
		sessions, err := a.store.LocatedSessionsSince(friendIDs, since)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		for _, sess := range sessions {
			if sess.Lat == nil || sess.Lng == nil {
				continue
			}
			// Sessions are newest-first; keep the first (latest) per user.
			if _, seen := pins[sess.UserID]; seen {
				continue
			}
			user, err := a.store.GetUserByID(sess.UserID)
			if err != nil {
				continue
			}
			typ := string(sess.Type)
			pins[sess.UserID] = mapPin{
				User:      user.Public(),
				Lat:       *sess.Lat,
				Lng:       *sess.Lng,
				Kind:      "smoking",
				Type:      &typ,
				Subtype:   sess.Subtype,
				At:        sess.OccurredAt,
				ExpiresAt: sess.OccurredAt.Add(presenceTTL),
			}
		}
	}

	out := make([]mapPin, 0, len(pins))
	for _, p := range pins {
		out = append(out, p)
	}
	writeJSON(w, http.StatusOK, map[string]any{"pins": out})
}
