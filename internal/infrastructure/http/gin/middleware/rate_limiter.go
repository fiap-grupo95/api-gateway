package middleware

import (
	"net/http"

	domsvc "github.com/fiap/secure-systems/api-gateway/internal/domain/service"
	"github.com/gin-gonic/gin"
)

// RateLimiterMiddleware retorna um middleware que delega o controle de rate ao serviço de domínio.
// O IP do cliente é usado como identificador de usuário.
func RateLimiterMiddleware(svc domsvc.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !svc.Allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": "60",
			})
			return
		}
		c.Next()
	}
}
