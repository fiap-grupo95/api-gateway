package di

import (
	"github.com/fiap/secure-systems/api-gateway/internal/domain/entity"
	"github.com/fiap/secure-systems/api-gateway/internal/infrastructure/gateway"
	"github.com/fiap/secure-systems/api-gateway/internal/infrastructure/http/gin/handler"
	"github.com/fiap/secure-systems/api-gateway/internal/infrastructure/service"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/authenticate"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/get_process_status"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/get_report"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/upload_diagram"
	"go.uber.org/zap"
)

// Container é o container de Dependency Injection
type Container struct {
	DiagramHandler *handler.DiagramHandler
	AuthHandler    *handler.AuthHandler
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
	Logger                *zap.Logger
}

// NewContainer cria um novo container com todas as dependências configuradas
func NewContainer(cfg *Config) *Container {
	// ═══════════════════════════════════════════════════════════════
	// INFRASTRUCTURE LAYER (outer) - Adapters
	// ═══════════════════════════════════════════════════════════════

	// Gateways (adapters para serviços externos)
	orchestratorGateway := gateway.NewOrchestratorHTTPGateway(cfg.UploadOrchestratorURL)
	reportGateway := gateway.NewReportHTTPGateway(cfg.ReportServiceURL)

	// Domain Services
	mimeValidator := service.NewMIMEValidator(cfg.AllowedMIMETypes)

	// ═══════════════════════════════════════════════════════════════
	// APPLICATION LAYER - Use Cases
	// ═══════════════════════════════════════════════════════════════

	// Upload diagram use case
	uploadUseCase := upload_diagram.New(
		orchestratorGateway,
		mimeValidator,
		cfg.MaxUploadSizeBytes,
	)

	// Get process status use case
	getStatusUseCase := get_process_status.New(orchestratorGateway)

	// Get report use case
	getReportUseCase := get_report.New(reportGateway)

	// Authenticate use case
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
		cfg.Logger,
	)

	authHandler := handler.NewAuthHandler(authenticateUseCase, cfg.Logger)

	return &Container{
		DiagramHandler: diagramHandler,
		AuthHandler:    authHandler,
	}
}
