package upload_diagram

import (
	"time"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/entity"
)

// Output representa a saída do use case
type Output struct {
	ProcessID entity.ProcessID
	Status    entity.ProcessStatus
	CreatedAt time.Time
}
