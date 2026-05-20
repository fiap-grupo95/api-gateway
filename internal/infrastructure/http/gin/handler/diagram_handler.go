package handler

import (
	"io"
	"net/http"

	"github.com/fiap/secure-systems/api-gateway/internal/infrastructure/http/gin/middleware"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/get_process_status"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/get_report"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/upload_diagram"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DiagramHandler é o adapter HTTP para os use cases de diagrama
type DiagramHandler struct {
	uploadUseCase    *upload_diagram.UseCase
	getStatusUseCase *get_process_status.UseCase
	getReportUseCase *get_report.UseCase
	maxUploadBytes   int64
	log              *zap.Logger
}

// NewDiagramHandler cria uma nova instância
func NewDiagramHandler(
	uploadUseCase *upload_diagram.UseCase,
	getStatusUseCase *get_process_status.UseCase,
	getReportUseCase *get_report.UseCase,
	maxUploadBytes int64,
	log *zap.Logger,
) *DiagramHandler {
	return &DiagramHandler{
		uploadUseCase:    uploadUseCase,
		getStatusUseCase: getStatusUseCase,
		getReportUseCase: getReportUseCase,
		maxUploadBytes:   maxUploadBytes,
		log:              log,
	}
}

// Upload é o handler HTTP para POST /api/diagrams
// Responsabilidade: apenas adaptar HTTP ↔ Use Case
func (h *DiagramHandler) Upload(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	// Limita o corpo antes de qualquer leitura
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxUploadBytes)

	// 1. Parse HTTP request
	file, header, err := c.Request.FormFile("diagram")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "field 'diagram' is required (multipart/form-data)"})
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read file"})
		return
	}

	// 2. Criar input do use case
	input := upload_diagram.Input{
		Content:  content,
		Filename: header.Filename,
	}

	// 3. Executar use case
	output, err := h.uploadUseCase.Execute(c.Request.Context(), input, requestID)
	if err != nil {
		h.log.Error("upload use case failed",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 4. Formatar resposta HTTP
	c.JSON(http.StatusAccepted, gin.H{
		"process_id": output.ProcessID.String(),
		"status":     output.Status.String(),
		"created_at": output.CreatedAt,
	})
}

// GetStatus é o handler HTTP para GET /api/process/:processId/status
func (h *DiagramHandler) GetStatus(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	processID := c.Param("processId")

	input := get_process_status.Input{
		ProcessID: processID,
	}

	output, err := h.getStatusUseCase.Execute(c.Request.Context(), input, requestID)
	if err != nil {
		h.log.Error("get status use case failed",
			zap.String("request_id", requestID),
			zap.String("process_id", processID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"process_id": output.ProcessID.String(),
		"status":     output.Status.String(),
	}
	if output.ReportID != nil {
		response["report_id"] = output.ReportID.String()
	}
	if output.Error != "" {
		response["error"] = output.Error
	}

	c.JSON(http.StatusOK, response)
}

// GetReport é o handler HTTP para GET /api/reports/:reportId
func (h *DiagramHandler) GetReport(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	reportID := c.Param("reportId")

	input := get_report.Input{
		ReportID: reportID,
	}

	output, err := h.getReportUseCase.Execute(c.Request.Context(), input, requestID)
	if err != nil {
		h.log.Error("get report use case failed",
			zap.String("request_id", requestID),
			zap.String("report_id", reportID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"report_id":       output.ReportID.String(),
		"process_id":      output.ProcessID.String(),
		"components":      output.Components,
		"risks":           output.Risks,
		"recommendations": output.Recommendations,
		"created_at":      output.CreatedAt,
	})
}
