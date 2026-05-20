package get_report

import (
	"context"
	"fmt"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/entity"
	"github.com/fiap/secure-systems/api-gateway/internal/domain/gateway"
)

// UseCase representa o caso de uso de consulta de relatório
type UseCase struct {
	reportGateway gateway.ReportGateway
}

// New cria uma nova instância do use case
func New(reportGateway gateway.ReportGateway) *UseCase {
	return &UseCase{
		reportGateway: reportGateway,
	}
}

// Execute executa o caso de uso
func (uc *UseCase) Execute(ctx context.Context, input Input, requestID string) (*Output, error) {
	// 1. Validar input
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	reportID := entity.ReportID(input.ReportID)

	// 2. Buscar relatório via gateway
	dto, err := uc.reportGateway.GetReport(ctx, reportID, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	// 3. Retornar output
	return &Output{
		ReportID:        dto.ReportID,
		ProcessID:       dto.ProcessID,
		Components:      dto.Components,
		Risks:           dto.Risks,
		Recommendations: dto.Recommendations,
		CreatedAt:       dto.CreatedAt,
	}, nil
}
