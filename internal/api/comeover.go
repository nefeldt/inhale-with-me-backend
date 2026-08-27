package api

import (
	"errors"
	"net/http"

	"github.com/nfeldt/inhale-with-me/internal/store"
)

type comeOverRequest struct {
	RecipientID *string `json:"recipient_id"`
	Username    *string `json:"username"`
}

// handleComeOver sends a friend a "Bleib sitzen, ich komme vorbei" nudge. It is
// a notification only — no location is shared or stored.
func (a *API) handleComeOver(w http.ResponseWriter, r *http.Request) {
	var req comeOverRequest
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
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "you cannot nudge yourself", nil)
		return
	}
	friends, err := a.store.AreFriends(me, recipient.ID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if !friends {
		writeError(w, http.StatusForbidden, "not_friends", "you can only do this with friends", nil)
		return
	}

	sender, _ := a.store.GetUserByID(me)
	a.pushToUser(recipient.ID, "Inhale With Me", displayName(sender)+": Bleib sitzen, ich komme vorbei", map[string]string{
		"kind": "come_over",
	})
	w.WriteHeader(http.StatusNoContent)
}
