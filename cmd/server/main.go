// Command server runs the Inhale With Me REST API.
package main

import (
	"context"
	"errors"
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
	handler := api.New(st, tokens, cfg).Router()

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
