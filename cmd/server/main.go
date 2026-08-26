// Command server runs the Inhale With Me REST API.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nfeldt/inhale-with-me/internal/api"
	"github.com/nfeldt/inhale-with-me/internal/auth"
	"github.com/nfeldt/inhale-with-me/internal/config"
	"github.com/nfeldt/inhale-with-me/internal/database"
	"github.com/nfeldt/inhale-with-me/internal/push"
	"github.com/nfeldt/inhale-with-me/internal/store"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "err", err)
		os.Exit(1)
	}

	// Admin CLI: `server reset-password <login> <new-password>` (and future
	// one-off commands). Runs against the same DB, then exits without serving.
	if len(os.Args) > 1 {
		if err := runCommand(os.Args[1:], cfg); err != nil {
			slog.Error("command failed", "err", err)
			os.Exit(1)
		}
		return
	}

	db, err := database.Open(cfg.DBPath, !cfg.IsProduction())
	if err != nil {
		slog.Error("open database", "err", err)
		os.Exit(1)
	}
	if err := database.Migrate(db); err != nil {
		slog.Error("run migrations", "err", err)
		os.Exit(1)
	}

	st := store.New(db)
	tokens := auth.NewManager(cfg.JWTSecret, cfg.TokenTTL)

	var notifier push.Notifier = push.Noop{}
	if cfg.APNsKeyP8 != "" && cfg.APNsKeyID != "" && cfg.APNsTeamID != "" {
		n, err := push.NewAPNs([]byte(cfg.APNsKeyP8), cfg.APNsKeyID, cfg.APNsTeamID, cfg.APNsBundleID, cfg.APNsProduction)
		if err != nil {
			slog.Error("APNs init failed; push disabled", "err", err)
		} else {
			notifier = n
			slog.Info("APNs push enabled", "production", cfg.APNsProduction, "bundle", cfg.APNsBundleID)
		}
	} else {
		// Log which piece is missing so a misconfigured secret is easy to spot.
		slog.Info("APNs push disabled",
			"hasKey", cfg.APNsKeyP8 != "",
			"hasKeyID", cfg.APNsKeyID != "",
			"hasTeamID", cfg.APNsTeamID != "")
	}

	handler := api.New(st, tokens, cfg, notifier).Router()

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	go func() {
		slog.Info("server listening", "port", cfg.Port, "env", cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
}

// runCommand dispatches admin subcommands invoked as `server <cmd> ...`.
func runCommand(args []string, cfg config.Config) error {
	switch args[0] {
	case "reset-password":
		if len(args) != 3 {
			return fmt.Errorf("usage: server reset-password <login> <new-password>")
		}
		return resetPassword(cfg, args[1], args[2])
	default:
		return fmt.Errorf("unknown command %q (available: reset-password)", args[0])
	}
}

// resetPassword sets a user's password to newPassword, identified by username or
// email. Opens the same database the server uses.
func resetPassword(cfg config.Config, login, newPassword string) error {
	if len(newPassword) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	db, err := database.Open(cfg.DBPath, !cfg.IsProduction())
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}()

	st := store.New(db)
	u, err := st.GetUserByLogin(login)
	if err != nil {
		return fmt.Errorf("find user %q: %w", login, err)
	}
	hash, err := auth.HashPassword(newPassword, cfg.BcryptCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	u.PasswordHash = hash
	if err := st.UpdateUser(u); err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	slog.Info("password reset", "login", login, "username", u.Username, "user_id", u.ID)
	return nil
}
