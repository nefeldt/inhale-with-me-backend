package auth

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey int

const userIDKey ctxKey = 0

// WithUserID returns a copy of ctx carrying the authenticated user id.
func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

// UserID returns the authenticated user id stored in ctx, if any.
func UserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}

// newID returns a new random UUID string (used for token IDs).
func newID() string { return uuid.NewString() }
