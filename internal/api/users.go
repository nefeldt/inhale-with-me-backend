package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/nfeldt/inhale-with-me/internal/model"
	"github.com/nfeldt/inhale-with-me/internal/store"
)

// handleMe returns the authenticated user's full profile.
func (a *API) handleMe(w http.ResponseWriter, r *http.Request) {
	u, err := a.store.GetUserByID(currentUserID(r))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

type updateMeRequest struct {
	DisplayName *string `json:"display_name"`
	Bio         *string `json:"bio"`
	AvatarURL   *string `json:"avatar_url"`
	Username    *string `json:"username"`
	Currency    *string `json:"currency"`
}

// handleUpdateMe applies a partial update to the authenticated user's profile.
func (a *API) handleUpdateMe(w http.ResponseWriter, r *http.Request) {
	var req updateMeRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body", nil)
		return
	}

	u, err := a.store.GetUserByID(currentUserID(r))
	if err != nil {
		writeStoreError(w, err)
		return
	}

	if req.DisplayName != nil {
		u.DisplayName = cleanPtr(req.DisplayName)
	}
	if req.Bio != nil {
		u.Bio = cleanPtr(req.Bio)
	}
	if req.AvatarURL != nil {
		u.AvatarURL = cleanPtr(req.AvatarURL)
	}
	if req.Currency != nil {
		if c := strings.ToUpper(strings.TrimSpace(*req.Currency)); len(c) == 3 {
			u.Currency = c
		}
	}
	if req.Username != nil {
		un := strings.ToLower(strings.TrimSpace(*req.Username))
		if len(un) < 3 || len(un) > 30 || !validUsername(un) {
			writeValidation(w, map[string]string{"username": "username must be 3-30 characters using letters, numbers, . _ -"})
			return
		}
		u.Username = un
	}

	if err := a.store.UpdateUser(u); err != nil {
		if errors.Is(err, store.ErrConflict) {
			writeError(w, http.StatusConflict, "conflict", "that username is already taken", nil)
			return
		}
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// handleSearchUsers finds users by a username/email prefix (excluding self).
func (a *API) handleSearchUsers(w http.ResponseWriter, r *http.Request) {
	users, err := a.store.SearchUsers(r.URL.Query().Get("query"), queryInt(r, "limit", 20))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	me := currentUserID(r)
	out := make([]model.PublicUser, 0, len(users))
	for _, u := range users {
		if u.ID == me {
			continue
		}
		pu := u.Public()
		if st, err := a.store.FriendStatus(me, u.ID); err == nil {
			pu.FriendStatus = st
		}
		out = append(out, pu)
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": out})
}

// handleGetUser returns a public profile with the viewer's friend status.
func (a *API) handleGetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	u, err := a.store.GetUserByID(id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	pu := u.Public()
	if me := currentUserID(r); me != id {
		if st, err := a.store.FriendStatus(me, id); err == nil {
			pu.FriendStatus = st
		}
	}
	writeJSON(w, http.StatusOK, pu)
}

type costSettingsRequest struct {
	Settings []model.CostSetting `json:"settings"`
}

// handleGetCostSettings returns the user's currency and per-type unit costs.
func (a *API) handleGetCostSettings(w http.ResponseWriter, r *http.Request) {
	me := currentUserID(r)
	cs, err := a.store.GetCostSettings(me)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	u, err := a.store.GetUserByID(me)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"currency": u.Currency, "settings": cs})
}

// handlePutCostSettings replaces the user's full set of cost settings.
func (a *API) handlePutCostSettings(w http.ResponseWriter, r *http.Request) {
	var req costSettingsRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid request body", nil)
		return
	}
	seen := make(map[model.SessionType]bool, len(req.Settings))
	for _, c := range req.Settings {
		if !c.Type.Valid() {
			writeValidation(w, map[string]string{"settings": "invalid session type: " + string(c.Type)})
			return
		}
		if c.UnitCostCents < 0 {
			writeValidation(w, map[string]string{"settings": "unit_cost_cents must be zero or greater"})
			return
		}
		if seen[c.Type] {
			writeValidation(w, map[string]string{"settings": "duplicate session type: " + string(c.Type)})
			return
		}
		seen[c.Type] = true
	}
	me := currentUserID(r)
	if err := a.store.ReplaceCostSettings(me, req.Settings); err != nil {
		writeStoreError(w, err)
		return
	}
	cs, err := a.store.GetCostSettings(me)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	u, err := a.store.GetUserByID(me)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"currency": u.Currency, "settings": cs})
}
