package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/fiap/secure-systems/api-gateway/internal/infrastructure/config"
	"github.com/fiap/secure-systems/api-gateway/internal/infrastructure/di"
	"github.com/fiap/secure-systems/api-gateway/internal/infrastructure/http/gin/router"
	"github.com/newrelic/go-agent/v3/newrelic"
	"go.uber.org/zap"
)

func main() {
	log, _ := zap.NewProduction()
	defer log.Sync()

	// ─── Load Configuration ────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load config", zap.Error(err))
	}

	// ─── New Relic ────────────────────────────────────────────────────────────
	nrApp, err := newrelic.NewApplication(newrelic.ConfigFromEnvironment())
	if err != nil {
		log.Warn("new relic not configured", zap.Error(err))
		nrApp, _ = newrelic.NewApplication(newrelic.ConfigEnabled(false))
	}

	// ─── Dependency Injection Container ───────────────────────────────────────
	container := di.NewContainer(&di.Config{
		UploadOrchestratorURL: cfg.UploadOrchestratorURL,
		ReportServiceURL:      cfg.ReportServiceURL,
		MaxUploadSizeBytes:    cfg.MaxUploadSizeMB * 1024 * 1024,
		AuthUsername:          cfg.AuthUsername,
		AuthPasswordHash:      cfg.AuthPasswordHash,
		JWTSecret:             cfg.JWTSecret,
		AllowedMIMETypes:      cfg.AllowedMIMETypes,
		Logger:                log,
	})

	// ─── Router ───────────────────────────────────────────────────────────────
	r := router.NewRouter(container, cfg.JWTSecret, log, nrApp)

	// ─── HTTP Server ──────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r.Engine(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("api-gateway started", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", zap.Error(err))
	}
	log.Info("api-gateway stopped")
}
