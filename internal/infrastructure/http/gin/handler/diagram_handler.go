package handler

import (
	"io"
	"net/http"

	"github.com/fiap/secure-systems/api-gateway/internal/infrastructure/http/gin/middleware"
	"github.com/fiap/secure-systems/api-gateway/internal/logging"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/get_process_status"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/get_report"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/upload_diagram"
	"github.com/gin-gonic/gin"
)

type DiagramHandler struct {
	uploadUseCase    *upload_diagram.UseCase
	getStatusUseCase *get_process_status.UseCase
	getReportUseCase *get_report.UseCase
	maxUploadBytes   int64
}

func NewDiagramHandler(
	uploadUseCase *upload_diagram.UseCase,
	getStatusUseCase *get_process_status.UseCase,
	getReportUseCase *get_report.UseCase,
	maxUploadBytes int64,
) *DiagramHandler {
	return &DiagramHandler{
		uploadUseCase:    uploadUseCase,
		getStatusUseCase: getStatusUseCase,
		getReportUseCase: getReportUseCase,
		maxUploadBytes:   maxUploadBytes,
	}
}

func (h *DiagramHandler) Upload(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	log := logging.LoggerWithContext(c.Request.Context())

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxUploadBytes)

	file, header, err := c.Request.FormFile("diagram")
	if err != nil {
		log.Warn().
			Str("request_id", requestID).
			Err(err).
			Msg("missing or unreadable 'diagram' field in multipart form")
		c.JSON(http.StatusBadRequest, gin.H{"error": "field 'diagram' is required (multipart/form-data)"})
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		log.Error().
			Str("request_id", requestID).
			Err(err).
			Msg("failed to read uploaded file")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read file"})
		return
	}

	input := upload_diagram.Input{
		Content:  content,
		Filename: header.Filename,
	}

	output, err := h.uploadUseCase.Execute(c.Request.Context(), input, requestID)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("filename", header.Filename).
			Msg("upload use case failed")
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"process_id": output.ProcessID.String(),
		"status":     output.Status.String(),
		"created_at": output.CreatedAt,
	})
}

func (h *DiagramHandler) GetStatus(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	processID := c.Param("processId")
	log := logging.LoggerWithContext(c.Request.Context())

	output, err := h.getStatusUseCase.Execute(c.Request.Context(), get_process_status.Input{ProcessID: processID}, requestID)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("process_id", processID).
			Msg("get status use case failed")
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

func (h *DiagramHandler) GetReport(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	reportID := c.Param("reportId")
	log := logging.LoggerWithContext(c.Request.Context())

	output, err := h.getReportUseCase.Execute(c.Request.Context(), get_report.Input{ReportID: reportID}, requestID)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("report_id", reportID).
			Msg("get report use case failed")
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
