package upload_diagram

import (
	"context"
	"fmt"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/entity"
	"github.com/fiap/secure-systems/api-gateway/internal/domain/errors"
	"github.com/fiap/secure-systems/api-gateway/internal/domain/gateway"
	"github.com/fiap/secure-systems/api-gateway/internal/domain/service"
)

// UseCase representa o caso de uso de upload de diagrama
type UseCase struct {
	orchestratorGateway gateway.OrchestratorGateway
	mimeValidator       service.MIMEValidator
	maxSizeBytes        int64
}

// New cria uma nova instância do use case
func New(
	orchestratorGateway gateway.OrchestratorGateway,
	mimeValidator service.MIMEValidator,
	maxSizeBytes int64,
) *UseCase {
	return &UseCase{
		orchestratorGateway: orchestratorGateway,
		mimeValidator:       mimeValidator,
		maxSizeBytes:        maxSizeBytes,
	}
}

// Execute executa o caso de uso
func (uc *UseCase) Execute(ctx context.Context, input Input, requestID string) (*Output, error) {
	// 1. Validar input
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// 2. Validar tamanho (regra de negócio)
	if int64(len(input.Content)) > uc.maxSizeBytes {
		return nil, fmt.Errorf("%w: %d bytes (max: %d)", errors.ErrFileTooLarge, len(input.Content), uc.maxSizeBytes)
	}

	// 3. Validar MIME type (domain service)
	detectedMIME, err := uc.mimeValidator.ValidateContent(input.Content)
	if err != nil {
		return nil, fmt.Errorf("invalid file content: %w", err)
	}

	if !uc.mimeValidator.IsAllowed(detectedMIME) {
		return nil, fmt.Errorf("MIME type not allowed: %s", detectedMIME)
	}

	// 4. Criar entidade Diagram (validação de domínio)
	diagram, err := entity.NewDiagram(input.Filename, input.Content, detectedMIME)
	if err != nil {
		return nil, fmt.Errorf("invalid diagram: %w", err)
	}

	// 5. Criar entidade Process
	process, err := entity.NewProcess(diagram.ID())
	if err != nil {
		return nil, fmt.Errorf("failed to create process: %w", err)
	}

	// 6. Submeter ao orchestrator (via gateway)
	processID, err := uc.orchestratorGateway.SubmitDiagram(ctx, diagram, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to submit diagram: %w", err)
	}

	// 7. Retornar output
	return &Output{
		ProcessID: processID,
		Status:    process.Status(),
		CreatedAt: process.CreatedAt(),
	}, nil
}
