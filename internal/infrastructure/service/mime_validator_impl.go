package service

import (
	"net/http"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/entity"
	domainService "github.com/fiap/secure-systems/api-gateway/internal/domain/service"
)

// MIMEValidatorImpl é a implementação do MIMEValidator
type MIMEValidatorImpl struct {
	allowedTypes []entity.MIMEType
}

// NewMIMEValidator cria uma nova instância
func NewMIMEValidator(allowedTypes []entity.MIMEType) *MIMEValidatorImpl {
	return &MIMEValidatorImpl{
		allowedTypes: allowedTypes,
	}
}

// ValidateContent valida o conteúdo detectando o MIME type pelos magic bytes
func (v *MIMEValidatorImpl) ValidateContent(content []byte) (entity.MIMEType, error) {
	// Detecta pelos primeiros 512 bytes
	sniff := content
	if len(content) > 512 {
		sniff = content[:512]
	}

	detected := http.DetectContentType(sniff)
	mimeType := entity.MIMEType(detected)

	if err := mimeType.Validate(); err != nil {
		return "", err
	}

	return mimeType, nil
}

// IsAllowed verifica se o MIME type é permitido
func (v *MIMEValidatorImpl) IsAllowed(mimeType entity.MIMEType) bool {
	for _, allowed := range v.allowedTypes {
		if allowed == mimeType {
			return true
		}
	}
	return false
}

// Compile-time check
var _ domainService.MIMEValidator = (*MIMEValidatorImpl)(nil)
