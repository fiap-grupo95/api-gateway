package handler

import (
	"errors"
	"net/http"

	domainErrors "github.com/fiap/secure-systems/api-gateway/internal/domain/errors"
	"github.com/fiap/secure-systems/api-gateway/internal/infrastructure/http/gin/middleware"
	"github.com/fiap/secure-systems/api-gateway/internal/logging"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/authenticate"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authenticateUseCase *authenticate.UseCase
}

func NewAuthHandler(authenticateUseCase *authenticate.UseCase) *AuthHandler {
	return &AuthHandler{authenticateUseCase: authenticateUseCase}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=1,max=64"`
		Password string `json:"password" binding:"required,min=4,max=128"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logging.LoggerWithContext(c.Request.Context()).Warn().
			Str("request_id", middleware.GetRequestID(c)).
			Str("remote_addr", c.ClientIP()).
			Msg("invalid login request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	input := authenticate.Input{
		Username: req.Username,
		Password: req.Password,
	}

	output, err := h.authenticateUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		log := logging.LoggerWithContext(c.Request.Context())
		if errors.Is(err, domainErrors.ErrInvalidCredentials) {
			log.Warn().
				Str("request_id", middleware.GetRequestID(c)).
				Str("remote_addr", c.ClientIP()).
				Msg("failed login attempt")
		} else {
			log.Error().
				Err(err).
				Str("request_id", middleware.GetRequestID(c)).
				Msg("authentication use case failed")
		}
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      output.Token,
		"expires_in": output.ExpiresIn,
	})
}
