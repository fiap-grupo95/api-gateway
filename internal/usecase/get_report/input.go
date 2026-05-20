package get_report

import (
	"fmt"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/entity"
)

// Input representa a entrada do use case
type Input struct {
	ReportID string
}

// Validate valida a entrada
func (i Input) Validate() error {
	if i.ReportID == "" {
		return fmt.Errorf("report ID is required")
	}
	rid := entity.ReportID(i.ReportID)
	return rid.Validate()
}
