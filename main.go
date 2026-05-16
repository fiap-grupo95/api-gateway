package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/fiap/secure-systems/api-gateway/internal/client"
	"github.com/fiap/secure-systems/api-gateway/internal/config"
	"github.com/fiap/secure-systems/api-gateway/internal/handler"
	"github.com/fiap/secure-systems/api-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent/v3/integrations/nrgin"
	"github.com/newrelic/go-agent/v3/newrelic"
	"go.uber.org/zap"
)

func main() {
	log, _ := zap.NewProduction()
	defer log.Sync()

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

	// ─── Clients para serviços internos ───────────────────────────────────────
	orchestratorClient := client.NewOrchestratorClient(cfg.UploadOrchestratorURL)
	reportClient := client.NewReportClient(cfg.ReportServiceURL)

	// ─── Handler ──────────────────────────────────────────────────────────────
	diagramHandler := handler.NewDiagramHandler(
		orchestratorClient,
		reportClient,
		cfg.MaxUploadSizeMB,
		cfg.AllowedMIMETypes,
		log,
	)

	// ─── Roteador Gin ─────────────────────────────────────────────────────────
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Middlewares globais (ordem importa)
	r.Use(gin.Recovery())
	r.Use(nrgin.Middleware(nrApp))
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(log))
	// 100 req/min por IP, burst de 20
	r.Use(middleware.RateLimiter(100.0/60.0, 20))

	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })

	api := r.Group("/api")
	api.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		// Timeout maior para upload (LLM pode demorar no downstream)
		api.POST("/diagrams", diagramHandler.Upload)
		api.GET("/process/:processId/status", diagramHandler.GetStatus)
		api.GET("/reports/:reportId", diagramHandler.GetReport)
	}

	// ─── Servidor HTTP ────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
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
