package gateway

import (
	"context"
	"time"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/entity"
)

// ReportGateway é a interface para comunicação com o serviço de relatórios
// (Port da Clean Architecture)
type ReportGateway interface {
	// GetReport busca um relatório pelo ID
	GetReport(ctx context.Context, reportID entity.ReportID, requestID string) (*ReportDTO, error)
}

// ReportDTO é o DTO retornado pelo gateway
type ReportDTO struct {
	ReportID        entity.ReportID
	ProcessID       entity.ProcessID
	Components      []string
	Risks           []string
	Recommendations []string
	CreatedAt       time.Time
}
