package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/nfeldt/inhale-with-me/internal/model"
)

type createSessionRequest struct {
	Type       model.SessionType `json:"type"`
	Quantity   *float64          `json:"quantity"`
	Note       *string           `json:"note"`
	Mood       *string           `json:"mood"`
	Location   *string           `json:"location"`
	CostCents  *int64            `json:"cost_cents"`
	Visibility *model.Visibility `json:"visibility"`
	OccurredAt *time.Time        `json:"occurred_at"`
}

func (a *API) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	var req createSessionRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body", nil)
		return
	}

	fields := map[string]string{}
	if !req.Type.Valid() {
		fields["type"] = "invalid session type"
	}
	quantity := 1.0
	if req.Quantity != nil {
		quantity = *req.Quantity
	}
	if quantity <= 0 {
		fields["quantity"] = "quantity must be greater than 0"
	}
	visibility := model.VisibilityFriends
	if req.Visibility != nil {
		if !req.Visibility.Valid() {
			fields["visibility"] = "invalid visibility"
		} else {
			visibility = *req.Visibility
		}
	}
	if req.Mood != nil && strings.TrimSpace(*req.Mood) != "" && !model.Mood(strings.TrimSpace(*req.Mood)).Valid() {
		fields["mood"] = "invalid mood"
	}
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}

	occurredAt := time.Now().UTC()
	if req.OccurredAt != nil {
		occurredAt = req.OccurredAt.UTC()
	}

	sess := &model.SmokeSession{
		UserID:     currentUserID(r),
		Type:       req.Type,
		Quantity:   quantity,
		Note:       cleanPtr(req.Note),
		Mood:       cleanPtr(req.Mood),
		Location:   cleanPtr(req.Location),
		CostCents:  req.CostCents,
		Visibility: visibility,
		OccurredAt: occurredAt,
	}
	if err := a.store.CreateSession(sess); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sess)
}

func (a *API) handleListSessions(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 20)
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	before := queryTime(r, "before")

	var typ *model.SessionType
	if t := r.URL.Query().Get("type"); t != "" {
		st := model.SessionType(t)
		if st.Valid() {
			typ = &st
		}
	}

	sessions, err := a.store.ListSessionsByUser(currentUserID(r), before, typ, limit)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	resp := map[string]any{"sessions": sessions, "next_before": nil}
	if len(sessions) == limit {
		resp["next_before"] = sessions[len(sessions)-1].OccurredAt
	}
	writeJSON(w, http.StatusOK, resp)
}

func (a *API) handleGetSession(w http.ResponseWriter, r *http.Request) {
	sess, err := a.store.GetSession(chi.URLParam(r, "id"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	ok, err := a.store.CanView(currentUserID(r), sess)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "resource not found", nil)
		return
	}
	writeJSON(w, http.StatusOK, sess)
}

type updateSessionRequest struct {
	Type       *model.SessionType `json:"type"`
	Quantity   *float64           `json:"quantity"`
	Note       *string            `json:"note"`
	Mood       *string            `json:"mood"`
	Location   *string            `json:"location"`
	CostCents  *int64             `json:"cost_cents"`
	Visibility *model.Visibility  `json:"visibility"`
	OccurredAt *time.Time         `json:"occurred_at"`
}

func (a *API) handleUpdateSession(w http.ResponseWriter, r *http.Request) {
	sess, err := a.store.GetSession(chi.URLParam(r, "id"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if sess.UserID != currentUserID(r) {
		writeError(w, http.StatusNotFound, "not_found", "resource not found", nil)
		return
	}

	var req updateSessionRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body", nil)
		return
	}

	fields := map[string]string{}
	if req.Type != nil {
		if !req.Type.Valid() {
			fields["type"] = "invalid session type"
		} else {
			sess.Type = *req.Type
		}
	}
	if req.Quantity != nil {
		if *req.Quantity <= 0 {
			fields["quantity"] = "quantity must be greater than 0"
		} else {
			sess.Quantity = *req.Quantity
		}
	}
	if req.Visibility != nil {
		if !req.Visibility.Valid() {
			fields["visibility"] = "invalid visibility"
		} else {
			sess.Visibility = *req.Visibility
		}
	}
	if req.Mood != nil {
		if m := strings.TrimSpace(*req.Mood); m != "" && !model.Mood(m).Valid() {
			fields["mood"] = "invalid mood"
		} else {
			sess.Mood = cleanPtr(req.Mood)
		}
	}
	if len(fields) > 0 {
		writeValidation(w, fields)
		return
	}

	if req.Note != nil {
		sess.Note = cleanPtr(req.Note)
	}
	if req.Location != nil {
		sess.Location = cleanPtr(req.Location)
	}
	if req.CostCents != nil {
		sess.CostCents = req.CostCents
	}
	if req.OccurredAt != nil {
		sess.OccurredAt = req.OccurredAt.UTC()
	}

	if err := a.store.UpdateSession(sess); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sess)
}

func (a *API) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
	sess, err := a.store.GetSession(chi.URLParam(r, "id"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if sess.UserID != currentUserID(r) {
		writeError(w, http.StatusNotFound, "not_found", "resource not found", nil)
		return
	}
	if err := a.store.DeleteSession(sess.ID); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
