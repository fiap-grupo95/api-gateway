package middleware

import (
	"time"

	"github.com/fiap/secure-systems/api-gateway/internal/logging"
	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			path += "?" + c.Request.URL.RawQuery
		}

		c.Next()

		logging.LoggerWithContext(c.Request.Context()).Info().
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", c.Writer.Status()).
			Dur("latency", time.Since(start)).
			Str("ip", c.ClientIP()).
			Str("request_id", GetRequestID(c)).
			Int("bytes_out", c.Writer.Size()).
			Msg("request")
	}
}
