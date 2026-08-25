package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/nfeldt/inhale-with-me/internal/model"
	"github.com/nfeldt/inhale-with-me/internal/store"
)

// How long a shared location stays visible to the recipient.
const locationShareTTL = time.Hour

type locationShareRequest struct {
	RecipientID *string  `json:"recipient_id"`
	Username    *string  `json:"username"`
	Lat         float64  `json:"lat"`
	Lng         float64  `json:"lng"`
	Message     *string  `json:"message"`
}

// handleCreateLocationShare shares the caller's current location with ONE friend
// for a limited time and pushes them ("I'm coming over").
func (a *API) handleCreateLocationShare(w http.ResponseWriter, r *http.Request) {
	var req locationShareRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body", nil)
		return
	}
	me := currentUserID(r)

	recipient, err := a.lookupUser(deref(req.RecipientID), deref(req.Username))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "user not found", nil)
			return
		}
		writeStoreError(w, err)
		return
	}
	if recipient.ID == me {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "you cannot share with yourself", nil)
		return
	}
	if req.Lat < -90 || req.Lat > 90 || req.Lng < -180 || req.Lng > 180 {
		writeValidation(w, map[string]string{"lat": "invalid coordinates"})
		return
	}
	// Only friends can receive your location.
	friends, err := a.store.AreFriends(me, recipient.ID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if !friends {
		writeError(w, http.StatusForbidden, "not_friends", "you can only share your location with friends", nil)
		return
	}

	ls := &model.LocationShare{
		SenderID:    me,
		RecipientID: recipient.ID,
		Lat:         req.Lat,
		Lng:         req.Lng,
		Message:     cleanPtr(req.Message),
		ExpiresAt:   time.Now().UTC().Add(locationShareTTL),
	}
	if err := a.store.CreateLocationShare(ls); err != nil {
		writeStoreError(w, err)
		return
	}

	sender, _ := a.store.GetUserByID(me)
	body := displayName(sender) + " shared their location"
	if ls.Message != nil {
		body = displayName(sender) + ": " + *ls.Message
	}
	a.pushToUser(recipient.ID, "Inhale With Me", body, map[string]string{
		"kind":    "location_share",
		"shareId": ls.ID,
	})

	writeJSON(w, http.StatusCreated, ls)
}

type locationShareItem struct {
	Share  model.LocationShare `json:"share"`
	Sender model.PublicUser    `json:"sender"`
}

// handleListLocationShares returns active location shares sent to the caller.
func (a *API) handleListLocationShares(w http.ResponseWriter, r *http.Request) {
	shares, err := a.store.ActiveLocationSharesForRecipient(currentUserID(r))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	out := make([]locationShareItem, 0, len(shares))
	for _, ls := range shares {
		sender, err := a.store.GetUserByID(ls.SenderID)
		if err != nil {
			continue
		}
		out = append(out, locationShareItem{Share: ls, Sender: sender.Public()})
	}
	writeJSON(w, http.StatusOK, map[string]any{"shares": out})
}
