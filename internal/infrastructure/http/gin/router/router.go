package router

import (
	"net/http"

	"github.com/fiap/secure-systems/api-gateway/internal/infrastructure/di"
	"github.com/fiap/secure-systems/api-gateway/internal/infrastructure/http/gin/middleware"
	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent/v3/integrations/nrgin"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// Router configura todas as rotas da aplicação
type Router struct {
	engine    *gin.Engine
	container *di.Container
}

// NewRouter cria um novo router configurado
func NewRouter(container *di.Container, jwtSecret []byte, nrApp *newrelic.Application) *Router {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Middlewares globais (ordem importa)
	r.Use(gin.Recovery())
	r.Use(nrgin.Middleware(nrApp))
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.RateLimiterMiddleware(container.RateLimiter))

	// Health check
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Auth routes (sem autenticação)
	r.POST("/auth/login", container.AuthHandler.Login)

	// API routes (com autenticação JWT)
	api := r.Group("/api")
	api.Use(middleware.JWTAuth(jwtSecret))
	{
		api.POST("/diagrams", container.DiagramHandler.Upload)
		api.GET("/process/:processId/status", container.DiagramHandler.GetStatus)
		api.GET("/reports/:reportId", container.DiagramHandler.GetReport)
	}

	return &Router{
		engine:    r,
		container: container,
	}
}

// Engine retorna a engine do Gin para uso no servidor HTTP
func (r *Router) Engine() *gin.Engine {
	return r.engine
}
