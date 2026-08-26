package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/nfeldt/inhale-with-me/internal/model"
)

type createSessionRequest struct {
	Type       model.SessionType `json:"type"`
	Subtype    *string           `json:"subtype"`
	Quantity   *float64          `json:"quantity"`
	Note       *string           `json:"note"`
	Mood       *string           `json:"mood"`
	Location   *string           `json:"location"`
	Lat        *float64          `json:"lat"`
	Lng        *float64          `json:"lng"`
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
	typ := model.SessionType(strings.ToLower(strings.TrimSpace(string(req.Type))))
	if !validTypeName(string(typ)) {
		fields["type"] = "invalid session type"
	}
	var subtype *string
	if req.Subtype != nil {
		if st := strings.ToLower(strings.TrimSpace(*req.Subtype)); st != "" {
			if typ == model.TypeCigarette && !model.ValidSubtype(st) {
				fields["subtype"] = "invalid cigarette subtype"
			} else {
				subtype = &st
			}
		}
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
	var lat, lng *float64
	if req.Lat != nil || req.Lng != nil {
		switch {
		case req.Lat == nil || req.Lng == nil:
			fields["lat"] = "lat and lng must be provided together"
		case *req.Lat < -90 || *req.Lat > 90 || *req.Lng < -180 || *req.Lng > 180:
			fields["lat"] = "invalid coordinates"
		default:
			lat, lng = req.Lat, req.Lng
		}
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
		Type:       typ,
		Subtype:    subtype,
		Quantity:   quantity,
		Note:       cleanPtr(req.Note),
		Mood:       cleanPtr(req.Mood),
		Location:   cleanPtr(req.Location),
		Lat:        lat,
		Lng:        lng,
		CostCents:  req.CostCents,
		Visibility: visibility,
		OccurredAt: occurredAt,
	}
	if err := a.store.CreateSession(sess); err != nil {
		writeStoreError(w, err)
		return
	}
	a.notifyFriendsOfSession(sess)
	writeJSON(w, http.StatusCreated, sess)
}

// notifyFriendsOfSession sends a push to the author's friends (unless the
// session is private). Runs asynchronously so it never blocks the response; a
// no-op when push is not configured.
func (a *API) notifyFriendsOfSession(sess *model.SmokeSession) {
	if sess.Visibility == model.VisibilityPrivate {
		return
	}
	go func() {
		author, err := a.store.GetUserByID(sess.UserID)
		if err != nil {
			return
		}
		friendIDs, err := a.store.FriendIDs(sess.UserID)
		if err != nil || len(friendIDs) == 0 {
			return
		}
		devices, err := a.store.DevicesForUsers(friendIDs)
		if err != nil || len(devices) == 0 {
			return
		}
		tokens := make([]string, 0, len(devices))
		for _, d := range devices {
			tokens = append(tokens, d.Token)
		}
		name := author.Username
		if author.DisplayName != nil && *author.DisplayName != "" {
			name = *author.DisplayName
		}
		body := fmt.Sprintf("%s just had a %s", name, sess.Type)
		invalid := a.notifier.Send(tokens, "Inhale With Me", body, map[string]string{
			"type":      string(sess.Type),
			"sessionId": sess.ID,
			"authorId":  sess.UserID,
		})
		for _, t := range invalid {
			_ = a.store.DeleteDeviceByToken(t)
		}
	}()
}

func (a *API) handleListSessions(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 20)
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var before *time.Time
	var beforeID string
	if t, id, ok := decodeCursor(r.URL.Query().Get("before")); ok {
		before = &t
		beforeID = id
	}

	var typ *model.SessionType
	if t := r.URL.Query().Get("type"); t != "" {
		st := model.SessionType(t)
		if st.Valid() {
			typ = &st
		}
	}

	sessions, err := a.store.ListSessionsByUser(currentUserID(r), before, beforeID, typ, limit)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	resp := map[string]any{"sessions": sessions, "next_before": nil}
	if len(sessions) == limit {
		last := sessions[len(sessions)-1]
		resp["next_before"] = encodeCursor(last.OccurredAt, last.ID)
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
	Subtype    *string            `json:"subtype"`
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
		t := model.SessionType(strings.ToLower(strings.TrimSpace(string(*req.Type))))
		if !validTypeName(string(t)) {
			fields["type"] = "invalid session type"
		} else {
			sess.Type = t
		}
	}
	if req.Subtype != nil {
		st := strings.ToLower(strings.TrimSpace(*req.Subtype))
		switch {
		case st == "":
			sess.Subtype = nil
		case sess.Type == model.TypeCigarette && !model.ValidSubtype(st):
			fields["subtype"] = "invalid cigarette subtype"
		default:
			sess.Subtype = &st
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
