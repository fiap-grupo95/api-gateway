package handler

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"slices"
	"strings"

	"github.com/fiap/secure-systems/api-gateway/internal/client"
	"github.com/fiap/secure-systems/api-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DiagramHandler struct {
	orchestrator   *client.OrchestratorClient
	reportClient   *client.ReportClient
	maxUploadBytes int64
	allowedMIME    []string
	log            *zap.Logger
}

func NewDiagramHandler(
	orchestrator *client.OrchestratorClient,
	reportClient *client.ReportClient,
	maxMB int64,
	allowed []string,
	log *zap.Logger,
) *DiagramHandler {
	return &DiagramHandler{
		orchestrator:   orchestrator,
		reportClient:   reportClient,
		maxUploadBytes: maxMB * 1024 * 1024,
		allowedMIME:    allowed,
		log:            log,
	}
}

// POST /api/diagrams
func (h *DiagramHandler) Upload(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	// Limita o corpo antes de qualquer leitura para prevenir ataques de payload gigante.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxUploadBytes)

	file, header, err := c.Request.FormFile("diagram")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "field 'diagram' is required (multipart/form-data)"})
		return
	}
	defer file.Close()

	// Detecta MIME pelos primeiros 512 bytes do conteúdo binário — nunca pela extensão.
	sniff := make([]byte, 512)
	n, _ := file.Read(sniff)
	detectedMIME := http.DetectContentType(sniff[:n])

	if !slices.ContainsFunc(h.allowedMIME, func(m string) bool { return strings.EqualFold(m, detectedMIME) }) {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"error":         "file type not allowed",
			"detected_type": detectedMIME,
			"allowed_types": h.allowedMIME,
		})
		return
	}

	// Lê o restante do arquivo (após os bytes de sniff) e monta o conteúdo completo.
	rest, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read file"})
		return
	}
	fullContent := append(sniff[:n], rest...)

	out, statusCode, err := h.orchestrator.Upload(
		c.Request.Context(),
		fullContent,
		sanitizeFilename(header.Filename),
		detectedMIME,
		requestID,
	)
	if err != nil {
		h.log.Error("orchestrator upload failed",
			zap.String("request_id", requestID),
			zap.Int("upstream_status", statusCode),
			zap.Error(err),
		)
		c.JSON(http.StatusBadGateway, gin.H{"error": "upstream error"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"process_id": out.ProcessID,
		"status":     out.Status,
		"created_at": out.CreatedAt,
	})
}

// GET /api/process/:processId/status
func (h *DiagramHandler) GetStatus(c *gin.Context) {
	processID := c.Param("processId")
	requestID := middleware.GetRequestID(c)

	out, statusCode, err := h.orchestrator.GetStatus(c.Request.Context(), processID, requestID)
	if err != nil {
		h.log.Error("orchestrator get status failed",
			zap.String("process_id", processID),
			zap.String("request_id", requestID),
			zap.Int("upstream_status", statusCode),
			zap.Error(err),
		)
		c.JSON(mapUpstreamStatus(statusCode), gin.H{"error": "upstream error"})
		return
	}

	c.JSON(http.StatusOK, out)
}

// GET /api/reports/:reportId
func (h *DiagramHandler) GetReport(c *gin.Context) {
	reportID := c.Param("reportId")
	requestID := middleware.GetRequestID(c)

	out, statusCode, err := h.reportClient.GetReport(c.Request.Context(), reportID, requestID)
	if err != nil {
		h.log.Error("report service get report failed",
			zap.String("report_id", reportID),
			zap.String("request_id", requestID),
			zap.Int("upstream_status", statusCode),
			zap.Error(err),
		)
		c.JSON(mapUpstreamStatus(statusCode), gin.H{"error": "upstream error"})
		return
	}

	c.JSON(http.StatusOK, out)
}

// mapUpstreamStatus mapeia códigos de erro upstream para o cliente externo.
// Evita vazar detalhes internos enquanto propaga 404 e 400 corretamente.
func mapUpstreamStatus(upstreamCode int) int {
	switch upstreamCode {
	case http.StatusNotFound:
		return http.StatusNotFound
	case http.StatusBadRequest:
		return http.StatusBadRequest
	case http.StatusUnprocessableEntity:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusBadGateway
	}
}

// sanitizeFilename remove componentes de caminho e caracteres perigosos do nome do arquivo.
func sanitizeFilename(name string) string {
	base := filepath.Base(name)
	base = strings.NewReplacer("..", "", "/", "", "\\", "").Replace(base)
	if base == "" || base == "." {
		return "diagram"
	}
	return fmt.Sprintf("%s", base)
}
