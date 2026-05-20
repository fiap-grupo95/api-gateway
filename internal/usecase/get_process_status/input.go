package get_process_status

import (
	"fmt"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/entity"
)

// Input representa a entrada do use case
type Input struct {
	ProcessID string
}

// Validate valida a entrada
func (i Input) Validate() error {
	if i.ProcessID == "" {
		return fmt.Errorf("process ID is required")
	}
	pid := entity.ProcessID(i.ProcessID)
	return pid.Validate()
}
