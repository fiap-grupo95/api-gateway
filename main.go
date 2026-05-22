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
	"github.com/fiap/secure-systems/api-gateway/internal/logging"
	"github.com/newrelic/go-agent/v3/newrelic"
)

func main() {
	// ─── Load Configuration ────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	// ─── New Relic ────────────────────────────────────────────────────────────
	nrApp, err := newrelic.NewApplication(
		newrelic.ConfigFromEnvironment(),
		newrelic.ConfigDistributedTracerEnabled(true),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)
	if err != nil {
		nrApp, _ = newrelic.NewApplication(newrelic.ConfigEnabled(false))
	} else {
		_ = nrApp.WaitForConnection(5 * time.Second)
	}

	// ─── Logging (deve ser inicializado após o New Relic) ─────────────────────
	logging.Init(nrApp)
	log := logging.Logger()

	// ─── Dependency Injection Container ───────────────────────────────────────
	container := di.NewContainer(&di.Config{
		UploadOrchestratorURL: cfg.UploadOrchestratorURL,
		ReportServiceURL:      cfg.ReportServiceURL,
		MaxUploadSizeBytes:    cfg.MaxUploadSizeMB * 1024 * 1024,
		AuthUsername:          cfg.AuthUsername,
		AuthPasswordHash:      cfg.AuthPasswordHash,
		JWTSecret:             cfg.JWTSecret,
		AllowedMIMETypes:      cfg.AllowedMIMETypes,
	})

	// ─── Router ───────────────────────────────────────────────────────────────
	r := router.NewRouter(container, cfg.JWTSecret, nrApp)

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
		log.Info().Str("port", cfg.Port).Msg("api-gateway started")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("shutdown error")
	}
	log.Info().Msg("api-gateway stopped")
}
