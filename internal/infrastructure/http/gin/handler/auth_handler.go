package handler

import (
	"errors"
	"net/http"

	domainErrors "github.com/fiap/secure-systems/api-gateway/internal/domain/errors"
	domsvc "github.com/fiap/secure-systems/api-gateway/internal/domain/service"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/authenticate"
	"github.com/gin-gonic/gin"
)

// AuthHandler é o adapter HTTP para o use case de autenticação
type AuthHandler struct {
	authenticateUseCase *authenticate.UseCase
	log                 domsvc.Logger
}

// NewAuthHandler cria uma nova instância
func NewAuthHandler(authenticateUseCase *authenticate.UseCase, log domsvc.Logger) *AuthHandler {
	return &AuthHandler{
		authenticateUseCase: authenticateUseCase,
		log:                 log,
	}
}

// Login é o handler HTTP para POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=1,max=64"`
		Password string `json:"password" binding:"required,min=4,max=128"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	input := authenticate.Input{
		Username: req.Username,
		Password: req.Password,
	}

	output, err := h.authenticateUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, domainErrors.ErrInvalidCredentials) {
			h.log.Warn("failed login attempt", "remote_addr", c.ClientIP())
		} else {
			h.log.Error("authentication use case failed", "error", err)
		}
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      output.Token,
		"expires_in": output.ExpiresIn,
	})
}
