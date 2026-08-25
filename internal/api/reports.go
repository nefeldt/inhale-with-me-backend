package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/nfeldt/inhale-with-me/internal/store"
)

type reportRequest struct {
	Reason string `json:"reason"`
}

// handleReportSession files a report about a session (and its author) for review.
func (a *API) handleReportSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	sess, err := a.store.GetSession(sessionID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "session not found", nil)
			return
		}
		writeStoreError(w, err)
		return
	}

	var req reportRequest
	_ = decodeJSON(w, r, &req) // reason is optional; ignore empty/malformed bodies

	sid := sessionID
	if _, err := a.store.CreateReport(currentUserID(r), sess.UserID, &sid, strings.TrimSpace(req.Reason)); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
