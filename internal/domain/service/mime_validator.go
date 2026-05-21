package service

import (
	"github.com/fiap/secure-systems/api-gateway/internal/domain/entity"
)

// MIMEValidator valida o tipo MIME de um arquivo
// (Domain Service - lógica de domínio que não pertence a uma entidade específica)
type MIMEValidator interface {
	// ValidateContent valida o conteúdo verificando os magic bytes
	ValidateContent(content []byte) (entity.MIMEType, error)

	// IsAllowed verifica se o MIME type é permitido
	IsAllowed(mimeType entity.MIMEType) bool
}
