package handler

import (
	"crypto/subtle"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	username     string
	passwordHash []byte
	jwtSecret    []byte
	log          *zap.Logger
}

func NewAuthHandler(username, passwordHash string, jwtSecret []byte, log *zap.Logger) *AuthHandler {
	return &AuthHandler{
		username:     username,
		passwordHash: []byte(passwordHash),
		jwtSecret:    jwtSecret,
		log:          log,
	}
}

type loginRequest struct {
	Username string `json:"username" binding:"required,min=1,max=64"`
	Password string `json:"password" binding:"required,min=4,max=128"`
}

// POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Ambas as checagens sempre executam para prevenir enumeração por timing (CWE-208).
	usernameMatch := subtle.ConstantTimeCompare([]byte(req.Username), []byte(h.username)) == 1
	passwordErr := bcrypt.CompareHashAndPassword(h.passwordHash, []byte(req.Password))

	if !usernameMatch || passwordErr != nil {
		h.log.Warn("failed login attempt", zap.String("remote_addr", c.ClientIP()))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"sub": req.Username,
		"iat": now.Unix(),
		"exp": now.Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(h.jwtSecret)
	if err != nil {
		h.log.Error("failed to sign jwt", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      signed,
		"expires_in": 3600,
	})
}
