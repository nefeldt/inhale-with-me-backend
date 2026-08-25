// Package store is the data-access layer over GORM.
package store

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Sentinel errors mapped to HTTP status codes by the API layer.
var (
	ErrNotFound  = errors.New("not found")
	ErrConflict  = errors.New("conflict")
	ErrForbidden = errors.New("forbidden")
)

// Store wraps a *gorm.DB and exposes typed data-access methods.
type Store struct {
	db *gorm.DB
}

// New returns a Store backed by db.
func New(db *gorm.DB) *Store { return &Store{db: db} }

func newID() string { return uuid.NewString() }

// translate converts GORM errors into store sentinels.
func translate(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return ErrConflict
	default:
		return err
	}
}
