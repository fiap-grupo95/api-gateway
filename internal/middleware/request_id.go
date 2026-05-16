package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	RequestIDKey    = "request_id"
	RequestIDHeader = "X-Request-ID"
)

// RequestID lê o X-Request-ID do cliente ou gera um UUID novo.
// Sempre injeta o valor no contexto Gin e no header de resposta.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(RequestIDHeader)
		if rid == "" {
			rid = uuid.New().String()
		}
		c.Set(RequestIDKey, rid)
		c.Header(RequestIDHeader, rid)
		c.Next()
	}
}

// GetRequestID recupera o request ID do contexto Gin.
func GetRequestID(c *gin.Context) string {
	v, _ := c.Get(RequestIDKey)
	if rid, ok := v.(string); ok {
		return rid
	}
	return ""
}
