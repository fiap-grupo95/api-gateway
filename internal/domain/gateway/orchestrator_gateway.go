package gateway

import (
	"context"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/entity"
)

// OrchestratorGateway é a interface para comunicação com o serviço de orquestração
// (Port da Clean Architecture)
type OrchestratorGateway interface {
	// SubmitDiagram envia um diagrama para processamento
	SubmitDiagram(ctx context.Context, diagram *entity.Diagram, requestID string) (entity.ProcessID, error)

	// GetProcessStatus consulta o status de um processo
	GetProcessStatus(ctx context.Context, processID entity.ProcessID, requestID string) (*ProcessStatusDTO, error)
}

// ProcessStatusDTO é o DTO retornado pelo gateway
type ProcessStatusDTO struct {
	ProcessID entity.ProcessID
	Status    entity.ProcessStatus
	ReportID  *entity.ReportID
	Error     string
}
