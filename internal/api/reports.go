package api

import (
	"errors"
	"log/slog"
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
	reason := strings.TrimSpace(req.Reason)
	rep, err := a.store.CreateReport(currentUserID(r), sess.UserID, &sid, reason)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	// Surface the report in the server logs so an operator can act on it (there
	// is no admin UI); also retrievable via `server list-reports`.
	slog.Warn("content reported",
		"report_id", rep.ID,
		"reporter_id", currentUserID(r),
		"reported_user_id", sess.UserID,
		"session_id", sessionID,
		"reason", reason)
	w.WriteHeader(http.StatusNoContent)
}
