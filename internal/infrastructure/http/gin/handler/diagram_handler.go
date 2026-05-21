package handler

import (
	"io"
	"net/http"

	domsvc "github.com/fiap/secure-systems/api-gateway/internal/domain/service"
	"github.com/fiap/secure-systems/api-gateway/internal/infrastructure/http/gin/middleware"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/get_process_status"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/get_report"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/upload_diagram"
	"github.com/gin-gonic/gin"
)

// DiagramHandler é o adapter HTTP para os use cases de diagrama
type DiagramHandler struct {
	uploadUseCase    *upload_diagram.UseCase
	getStatusUseCase *get_process_status.UseCase
	getReportUseCase *get_report.UseCase
	maxUploadBytes   int64
	log              domsvc.Logger
}

// NewDiagramHandler cria uma nova instância
func NewDiagramHandler(
	uploadUseCase *upload_diagram.UseCase,
	getStatusUseCase *get_process_status.UseCase,
	getReportUseCase *get_report.UseCase,
	maxUploadBytes int64,
	log domsvc.Logger,
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
func (h *DiagramHandler) Upload(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	// Limita o corpo antes de qualquer leitura
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxUploadBytes)

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

	input := upload_diagram.Input{
		Content:  content,
		Filename: header.Filename,
	}

	output, err := h.uploadUseCase.Execute(c.Request.Context(), input, requestID)
	if err != nil {
		h.log.Error("upload use case failed", "request_id", requestID, "error", err)
		RespondWithError(c, err)
		return
	}

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
		h.log.Error("get status use case failed", "request_id", requestID, "process_id", processID, "error", err)
		RespondWithError(c, err)
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
		h.log.Error("get report use case failed", "request_id", requestID, "report_id", reportID, "error", err)
		RespondWithError(c, err)
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
