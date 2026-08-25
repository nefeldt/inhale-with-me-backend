package api

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/nfeldt/inhale-with-me/internal/model"
)

type reactionUserItem struct {
	User      model.PublicUser `json:"user"`
	Type      string           `json:"type"`
	CreatedAt time.Time        `json:"created_at"`
}

func (a *API) handleListReactions(w http.ResponseWriter, r *http.Request) {
	me := currentUserID(r)
	sess, err := a.store.GetSession(chi.URLParam(r, "id"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	ok, err := a.store.CanView(me, sess)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "resource not found", nil)
		return
	}

	reactions, err := a.store.ListReactions(sess.ID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	out := make([]reactionUserItem, 0, len(reactions))
	for _, rr := range reactions {
		u, err := a.store.GetUserByID(rr.UserID)
		if err != nil {
			continue
		}
		out = append(out, reactionUserItem{User: u.Public(), Type: rr.Type, CreatedAt: rr.CreatedAt})
	}
	writeJSON(w, http.StatusOK, map[string]any{"reactions": out})
}

type addReactionRequest struct {
	Type *string `json:"type"`
}

func (a *API) handleAddReaction(w http.ResponseWriter, r *http.Request) {
	me := currentUserID(r)
	sess, err := a.store.GetSession(chi.URLParam(r, "id"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	ok, err := a.store.CanView(me, sess)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "resource not found", nil)
		return
	}

	// The body is optional; an empty body means the default "cheers" reaction.
	var req addReactionRequest
	if err := decodeJSON(w, r, &req); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body", nil)
		return
	}
	typ := model.DefaultReactionType
	if req.Type != nil {
		if t := strings.TrimSpace(*req.Type); t != "" {
			typ = t
		}
	}

	reaction, err := a.store.AddReaction(sess.ID, me, typ)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, reaction)
}

func (a *API) handleDeleteReaction(w http.ResponseWriter, r *http.Request) {
	me := currentUserID(r)
	if err := a.store.RemoveReaction(chi.URLParam(r, "id"), me, chi.URLParam(r, "type")); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
