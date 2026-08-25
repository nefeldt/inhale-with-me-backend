// Package api wires the HTTP router and request handlers.
package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/nfeldt/inhale-with-me/internal/auth"
	"github.com/nfeldt/inhale-with-me/internal/config"
	"github.com/nfeldt/inhale-with-me/internal/push"
	"github.com/nfeldt/inhale-with-me/internal/store"
)

// API bundles the dependencies shared by every handler.
type API struct {
	store     *store.Store
	tokens    *auth.Manager
	cfg       config.Config
	notifier  push.Notifier
	dummyHash string
}

// New builds an API. A nil notifier defaults to a no-op (push disabled).
func New(st *store.Store, tokens *auth.Manager, cfg config.Config, notifier push.Notifier) *API {
	if notifier == nil {
		notifier = push.Noop{}
	}
	// Precompute a bcrypt hash at the configured cost so a login for a
	// non-existent account still runs a comparison — closing the timing side
	// channel that would otherwise reveal whether an account is registered.
	dummy, _ := auth.HashPassword("inhale-with-me-timing-equalizer", cfg.BcryptCost)
	return &API{store: st, tokens: tokens, cfg: cfg, notifier: notifier, dummyHash: dummy}
}

// Router returns the fully configured HTTP handler.
func (a *API) Router() http.Handler {
	// Route the auth middleware's 401 through our standard error envelope.
	auth.Unauthorized = func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required", nil)
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   a.cfg.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/healthz", a.handleHealth)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", a.handleRegister)
		r.Post("/auth/login", a.handleLogin)

		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(a.tokens))

			r.Get("/users/me", a.handleMe)
			r.Patch("/users/me", a.handleUpdateMe)
			r.Delete("/users/me", a.handleDeleteMe) // App Store account deletion
			r.Get("/users/me/cost-settings", a.handleGetCostSettings)
			r.Put("/users/me/cost-settings", a.handlePutCostSettings)
			r.Get("/users", a.handleSearchUsers)
			r.Get("/users/{id}", a.handleGetUser)

			// Push device registration
			r.Post("/devices", a.handleRegisterDevice)
			r.Delete("/devices/{token}", a.handleUnregisterDevice)

			// UGC moderation (App Store 1.2): block + report
			r.Post("/blocks", a.handleBlockUser)
			r.Get("/blocks", a.handleListBlocks)
			r.Delete("/blocks/{userId}", a.handleUnblockUser)
			r.Post("/sessions/{id}/report", a.handleReportSession)

			r.Post("/sessions", a.handleCreateSession)
			r.Get("/sessions", a.handleListSessions)
			r.Get("/sessions/{id}", a.handleGetSession)
			r.Patch("/sessions/{id}", a.handleUpdateSession)
			r.Delete("/sessions/{id}", a.handleDeleteSession)
			r.Get("/sessions/{id}/reactions", a.handleListReactions)
			r.Post("/sessions/{id}/reactions", a.handleAddReaction)
			r.Delete("/sessions/{id}/reactions/{type}", a.handleDeleteReaction)

			r.Get("/stats/summary", a.handleStats)

			r.Post("/friends/requests", a.handleCreateFriendRequest)
			r.Get("/friends/requests", a.handleListFriendRequests)
			r.Post("/friends/requests/{id}/accept", a.handleAcceptFriendRequest)
			r.Post("/friends/requests/{id}/decline", a.handleDeclineFriendRequest)
			r.Delete("/friends/requests/{id}", a.handleCancelFriendRequest)
			r.Get("/friends", a.handleListFriends)
			r.Delete("/friends/{userId}", a.handleRemoveFriend)

			r.Get("/feed", a.handleFeed)
		})
	})

	return r
}

// currentUserID returns the authenticated user's id from the request context.
func currentUserID(r *http.Request) string {
	id, _ := auth.UserID(r.Context())
	return id
}
