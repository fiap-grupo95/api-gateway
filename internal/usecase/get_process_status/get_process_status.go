package get_process_status

import (
	"context"
	"fmt"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/entity"
	"github.com/fiap/secure-systems/api-gateway/internal/domain/gateway"
)

// UseCase representa o caso de uso de consulta de status de processo
type UseCase struct {
	orchestratorGateway gateway.OrchestratorGateway
}

// New cria uma nova instância do use case
func New(orchestratorGateway gateway.OrchestratorGateway) *UseCase {
	return &UseCase{
		orchestratorGateway: orchestratorGateway,
	}
}

// Execute executa o caso de uso
func (uc *UseCase) Execute(ctx context.Context, input Input, requestID string) (*Output, error) {
	// 1. Validar input
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	processID := entity.ProcessID(input.ProcessID)

	// 2. Consultar status via gateway
	dto, err := uc.orchestratorGateway.GetProcessStatus(ctx, processID, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get process status: %w", err)
	}

	// 3. Retornar output
	return &Output{
		ProcessID: dto.ProcessID,
		Status:    dto.Status,
		ReportID:  dto.ReportID,
		Error:     dto.Error,
	}, nil
}
