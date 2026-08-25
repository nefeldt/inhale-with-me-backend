package api

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/nfeldt/inhale-with-me/internal/model"
	"github.com/nfeldt/inhale-with-me/internal/store"
)

type createFriendRequestBody struct {
	Username    *string `json:"username"`
	AddresseeID *string `json:"addressee_id"`
}

func (a *API) handleCreateFriendRequest(w http.ResponseWriter, r *http.Request) {
	var body createFriendRequestBody
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body", nil)
		return
	}

	me := currentUserID(r)
	var addressee *model.User
	var err error
	switch {
	case body.AddresseeID != nil && *body.AddresseeID != "":
		addressee, err = a.store.GetUserByID(*body.AddresseeID)
	case body.Username != nil && *body.Username != "":
		addressee, err = a.store.GetUserByUsername(*body.Username)
	default:
		writeValidation(w, map[string]string{"username": "provide a username or addressee_id"})
		return
	}
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "user not found", nil)
			return
		}
		writeStoreError(w, err)
		return
	}
	if addressee.ID == me {
		writeError(w, http.StatusUnprocessableEntity, "validation_error", "you cannot add yourself", nil)
		return
	}
	if blocked, err := a.store.IsBlockedEitherWay(me, addressee.ID); err == nil && blocked {
		writeError(w, http.StatusForbidden, "blocked", "you can't send a request to this user", nil)
		return
	}

	f, err := a.store.CreateRequest(me, addressee.ID)
	if err != nil {
		if errors.Is(err, store.ErrConflict) {
			writeError(w, http.StatusConflict, "conflict", "a friend request or friendship already exists", nil)
			return
		}
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, f)
}

type friendRequestItem struct {
	Friendship model.Friendship `json:"friendship"`
	User       model.PublicUser `json:"user"`
}

func (a *API) handleListFriendRequests(w http.ResponseWriter, r *http.Request) {
	me := currentUserID(r)
	direction := r.URL.Query().Get("direction")
	if direction != "outgoing" {
		direction = "incoming"
	}

	fs, err := a.store.ListPendingRequests(me, direction)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	out := make([]friendRequestItem, 0, len(fs))
	for _, f := range fs {
		otherID := f.RequesterID
		if direction == "outgoing" {
			otherID = f.AddresseeID
		}
		u, err := a.store.GetUserByID(otherID)
		if err != nil {
			continue
		}
		out = append(out, friendRequestItem{Friendship: f, User: u.Public()})
	}
	writeJSON(w, http.StatusOK, map[string]any{"requests": out})
}

func (a *API) handleAcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	a.respondToRequest(w, r, model.FriendshipAccepted)
}

func (a *API) handleDeclineFriendRequest(w http.ResponseWriter, r *http.Request) {
	a.respondToRequest(w, r, model.FriendshipDeclined)
}

// respondToRequest lets the addressee accept or decline a pending request.
func (a *API) respondToRequest(w http.ResponseWriter, r *http.Request, status model.FriendshipStatus) {
	me := currentUserID(r)
	f, err := a.store.GetFriendship(chi.URLParam(r, "id"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if f.AddresseeID != me || f.Status != model.FriendshipPending {
		writeError(w, http.StatusNotFound, "not_found", "friend request not found", nil)
		return
	}
	updated, err := a.store.SetFriendshipStatus(f.ID, status)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (a *API) handleCancelFriendRequest(w http.ResponseWriter, r *http.Request) {
	me := currentUserID(r)
	f, err := a.store.GetFriendship(chi.URLParam(r, "id"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if f.RequesterID != me || f.Status != model.FriendshipPending {
		writeError(w, http.StatusNotFound, "not_found", "friend request not found", nil)
		return
	}
	if err := a.store.DeleteFriendship(f.ID); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) handleListFriends(w http.ResponseWriter, r *http.Request) {
	users, err := a.store.ListFriends(currentUserID(r))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	out := make([]model.PublicUser, 0, len(users))
	for _, u := range users {
		pu := u.Public()
		pu.FriendStatus = model.FriendStatusFriends
		out = append(out, pu)
	}
	writeJSON(w, http.StatusOK, map[string]any{"friends": out})
}

func (a *API) handleRemoveFriend(w http.ResponseWriter, r *http.Request) {
	if err := a.store.RemoveFriend(currentUserID(r), chi.URLParam(r, "userId")); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
