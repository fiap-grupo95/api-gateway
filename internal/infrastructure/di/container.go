package di

import (
	"github.com/fiap/secure-systems/api-gateway/internal/domain/entity"
	domsvc "github.com/fiap/secure-systems/api-gateway/internal/domain/service"
	"github.com/fiap/secure-systems/api-gateway/internal/infrastructure/gateway"
	"github.com/fiap/secure-systems/api-gateway/internal/infrastructure/http/gin/handler"
	"github.com/fiap/secure-systems/api-gateway/internal/infrastructure/service"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/authenticate"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/get_process_status"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/get_report"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/upload_diagram"
)

// Container é o container de Dependency Injection
type Container struct {
	DiagramHandler *handler.DiagramHandler
	AuthHandler    *handler.AuthHandler
	RateLimiter    domsvc.RateLimiter
}

// Config é a configuração para o container
type Config struct {
	UploadOrchestratorURL string
	ReportServiceURL      string
	MaxUploadSizeBytes    int64
	AuthUsername          string
	AuthPasswordHash      string
	JWTSecret             []byte
	AllowedMIMETypes      []entity.MIMEType
}

// NewContainer cria um novo container com todas as dependências configuradas
func NewContainer(cfg *Config) *Container {
	// ═══════════════════════════════════════════════════════════════
	// INFRASTRUCTURE LAYER (outer) - Adapters
	// ═══════════════════════════════════════════════════════════════

	// Rate limiter (100 req/min por IP, burst de 20)
	rateLimiter := service.NewTokenBucketRateLimiter(100.0/60.0, 20)

	// Gateways (adapters para serviços externos)
	orchestratorGateway := gateway.NewOrchestratorHTTPGateway(cfg.UploadOrchestratorURL)
	reportGateway := gateway.NewReportHTTPGateway(cfg.ReportServiceURL)

	// Domain Services
	mimeValidator := service.NewMIMEValidator(cfg.AllowedMIMETypes)

	// ═══════════════════════════════════════════════════════════════
	// APPLICATION LAYER - Use Cases
	// ═══════════════════════════════════════════════════════════════

	uploadUseCase := upload_diagram.New(
		orchestratorGateway,
		mimeValidator,
		cfg.MaxUploadSizeBytes,
	)

	getStatusUseCase := get_process_status.New(orchestratorGateway)

	getReportUseCase := get_report.New(reportGateway)

	authenticateUseCase := authenticate.New(
		cfg.AuthUsername,
		cfg.AuthPasswordHash,
		cfg.JWTSecret,
	)

	// ═══════════════════════════════════════════════════════════════
	// INTERFACE ADAPTERS - HTTP Handlers
	// ═══════════════════════════════════════════════════════════════

	diagramHandler := handler.NewDiagramHandler(
		uploadUseCase,
		getStatusUseCase,
		getReportUseCase,
		cfg.MaxUploadSizeBytes,
	)

	authHandler := handler.NewAuthHandler(authenticateUseCase)

	return &Container{
		DiagramHandler: diagramHandler,
		AuthHandler:    authHandler,
		RateLimiter:    rateLimiter,
	}
}
