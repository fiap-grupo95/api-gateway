package handler

import (
	"net/http"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/errors"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/authenticate"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthHandler é o adapter HTTP para o use case de autenticação
type AuthHandler struct {
	authenticateUseCase *authenticate.UseCase
	log                 *zap.Logger
}

// NewAuthHandler cria uma nova instância
func NewAuthHandler(authenticateUseCase *authenticate.UseCase, log *zap.Logger) *AuthHandler {
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

	// Criar input do use case
	input := authenticate.Input{
		Username: req.Username,
		Password: req.Password,
	}

	// Executar use case
	output, err := h.authenticateUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		if err == errors.ErrInvalidCredentials {
			h.log.Warn("failed login attempt", zap.String("remote_addr", c.ClientIP()))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		h.log.Error("authentication use case failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Formatar resposta HTTP
	c.JSON(http.StatusOK, gin.H{
		"token":      output.Token,
		"expires_in": output.ExpiresIn,
	})
}
