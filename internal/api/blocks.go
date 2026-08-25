package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/nfeldt/inhale-with-me/internal/model"
	"github.com/nfeldt/inhale-with-me/internal/store"
)

type blockRequest struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// handleBlockUser blocks another user (hides their content, severs friendship).
func (a *API) handleBlockUser(w http.ResponseWriter, r *http.Request) {
	var req blockRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body", nil)
		return
	}
	me := currentUserID(r)
	target, err := a.lookupUser(req.UserID, req.Username)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "user not found", nil)
			return
		}
		writeStoreError(w, err)
		return
	}
	if target.ID == me {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "you cannot block yourself", nil)
		return
	}
	if err := a.store.BlockUser(me, target.ID); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleUnblockUser removes a block.
func (a *API) handleUnblockUser(w http.ResponseWriter, r *http.Request) {
	if err := a.store.UnblockUser(currentUserID(r), chi.URLParam(r, "userId")); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleListBlocks lists the users the caller has blocked.
func (a *API) handleListBlocks(w http.ResponseWriter, r *http.Request) {
	users, err := a.store.ListBlockedUsers(currentUserID(r))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	out := make([]model.PublicUser, 0, len(users))
	for _, u := range users {
		out = append(out, u.Public())
	}
	writeJSON(w, http.StatusOK, map[string]any{"blocked": out})
}

// lookupUser resolves a user by id or username.
func (a *API) lookupUser(id, username string) (*model.User, error) {
	if id != "" {
		return a.store.GetUserByID(id)
	}
	if username != "" {
		return a.store.GetUserByUsername(username)
	}
	return nil, store.ErrNotFound
}
