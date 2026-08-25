// Package database opens the GORM/SQLite database and runs schema migrations.
package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite" // pure-Go SQLite driver (no cgo), GORM compatible
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/nfeldt/inhale-with-me/internal/model"
)

// Open opens (creating if needed) the SQLite database at path using the pure-Go
// driver, so the server builds and runs with CGO_ENABLED=0. Pragmas enable
// foreign keys, a busy timeout and WAL journaling for concurrent reads.
func Open(path string, debug bool) (*gorm.DB, error) {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db directory: %w", err)
		}
	}

	dsn := "file:" + path + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)"

	level := gormlogger.Warn
	if debug {
		level = gormlogger.Info
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		TranslateError: true, // surfaces gorm.ErrDuplicatedKey / ErrRecordNotFound
		Logger:         gormlogger.Default.LogMode(level),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("access sql.DB: %w", err)
	}
	// SQLite has a single writer; a small pool with WAL is the right tradeoff.
	sqlDB.SetMaxOpenConns(4)

	return db, nil
}

// Migrate creates or updates the schema for every entity via GORM AutoMigrate.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.SmokeSession{},
		&model.Friendship{},
		&model.Reaction{},
		&model.CostSetting{},
		&model.Device{},
		&model.Block{},
		&model.Report{},
	)
}
