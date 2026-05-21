package get_report

import (
	"time"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/entity"
)

// Output representa a saída do use case
type Output struct {
	ReportID        entity.ReportID
	ProcessID       entity.ProcessID
	Components      []string
	Risks           []string
	Recommendations []string
	CreatedAt       time.Time
}
