package handler

import (
	"net/http"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/errors"
	"github.com/gin-gonic/gin"
)

// ErrorResponse é a resposta padrão para erros
type ErrorResponse struct {
	Error      string `json:"error"`
	Message    string `json:"message,omitempty"`
	StatusCode int    `json:"status_code"`
}

// MapErrorToResponse converte um erro do domain para uma resposta HTTP
// Centraliza o mapeamento de erros para evitar duplicação nos handlers
func MapErrorToResponse(err error) (int, ErrorResponse) {
	switch err {
	// Erros de arquivo/diagrama
	case errors.ErrInvalidMIMEType:
		return http.StatusUnsupportedMediaType, ErrorResponse{
			Error:      "invalid_file_type",
			Message:    "file type not allowed",
			StatusCode: http.StatusUnsupportedMediaType,
		}
	case errors.ErrFileTooLarge:
		return 413, ErrorResponse{
			Error:      "file_too_large",
			Message:    "file size exceeds maximum allowed",
			StatusCode: 413,
		}
	case errors.ErrEmptyFilename, errors.ErrEmptyContent, errors.ErrEmptyDiagramID:
		return http.StatusBadRequest, ErrorResponse{
			Error:      "invalid_diagram",
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}

	// Erros de processo
	case errors.ErrEmptyProcessID, errors.ErrInvalidProcessIDFormat:
		return http.StatusBadRequest, ErrorResponse{
			Error:      "invalid_process_id",
			Message:    err.Error(),
			StatusCode: http.StatusBadRequest,
		}
	case errors.ErrInvalidStatusTransition, errors.ErrCannotFailCompletedProcess:
		return http.StatusConflict, ErrorResponse{
			Error:      "invalid_state",
			Message:    err.Error(),
			StatusCode: http.StatusConflict,
		}

	// Erros de autenticação
	case errors.ErrInvalidCredentials:
		return http.StatusUnauthorized, ErrorResponse{
			Error:      "invalid_credentials",
			Message:    "username or password is incorrect",
			StatusCode: http.StatusUnauthorized,
		}
	case errors.ErrExpiredToken:
		return http.StatusUnauthorized, ErrorResponse{
			Error:      "expired_token",
			Message:    "token has expired",
			StatusCode: http.StatusUnauthorized,
		}

	// Erros de gateway
	case errors.ErrGatewayUnavailable:
		return http.StatusServiceUnavailable, ErrorResponse{
			Error:      "service_unavailable",
			Message:    "external service is unavailable",
			StatusCode: http.StatusServiceUnavailable,
		}
	case errors.ErrInvalidResponse:
		return http.StatusBadGateway, ErrorResponse{
			Error:      "bad_gateway",
			Message:    "invalid response from external service",
			StatusCode: http.StatusBadGateway,
		}

	// Erro genérico
	default:
		return http.StatusInternalServerError, ErrorResponse{
			Error:      "internal_error",
			Message:    "an unexpected error occurred",
			StatusCode: http.StatusInternalServerError,
		}
	}
}

// RespondWithError entrega uma resposta de erro padronizada ao cliente
func RespondWithError(c *gin.Context, err error) {
	statusCode, errResponse := MapErrorToResponse(err)
	c.JSON(statusCode, errResponse)
}
