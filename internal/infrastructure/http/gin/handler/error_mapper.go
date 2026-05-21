package handler

import (
	"errors"
	"net/http"

	domainErrors "github.com/fiap/secure-systems/api-gateway/internal/domain/errors"
	"github.com/gin-gonic/gin"
)

// MapErrorToResponse converte um erro do domain para o HTTP status e mensagem correspondentes
func MapErrorToResponse(err error) (int, gin.H) {
	switch {
	case errors.Is(err, domainErrors.ErrInvalidMIMEType):
		return http.StatusUnsupportedMediaType, gin.H{"error": "file type not allowed"}

	case errors.Is(err, domainErrors.ErrFileTooLarge):
		return http.StatusRequestEntityTooLarge, gin.H{"error": "file size exceeds maximum allowed"}

	case errors.Is(err, domainErrors.ErrEmptyFilename),
		errors.Is(err, domainErrors.ErrEmptyContent),
		errors.Is(err, domainErrors.ErrEmptyDiagramID):
		return http.StatusBadRequest, gin.H{"error": "invalid diagram"}

	case errors.Is(err, domainErrors.ErrEmptyProcessID),
		errors.Is(err, domainErrors.ErrInvalidProcessIDFormat):
		return http.StatusBadRequest, gin.H{"error": "invalid process ID"}

	case errors.Is(err, domainErrors.ErrInvalidStatusTransition),
		errors.Is(err, domainErrors.ErrCannotFailCompletedProcess):
		return http.StatusConflict, gin.H{"error": "invalid state transition"}

	case errors.Is(err, domainErrors.ErrInvalidCredentials):
		return http.StatusUnauthorized, gin.H{"error": "invalid credentials"}

	case errors.Is(err, domainErrors.ErrExpiredToken):
		return http.StatusUnauthorized, gin.H{"error": "token expired"}

	case errors.Is(err, domainErrors.ErrGatewayUnavailable):
		return http.StatusServiceUnavailable, gin.H{"error": "service unavailable"}

	case errors.Is(err, domainErrors.ErrInvalidResponse):
		return http.StatusBadGateway, gin.H{"error": "upstream service returned an invalid response"}

	default:
		return http.StatusInternalServerError, gin.H{"error": "internal server error"}
	}
}

// RespondWithError entrega uma resposta de erro padronizada ao cliente
func RespondWithError(c *gin.Context, err error) {
	statusCode, body := MapErrorToResponse(err)
	c.JSON(statusCode, body)
}
