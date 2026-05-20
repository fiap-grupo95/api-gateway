package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/entity"
	domainGateway "github.com/fiap/secure-systems/api-gateway/internal/domain/gateway"
)

// OrchestratorHTTPGateway é a implementação HTTP do OrchestratorGateway
type OrchestratorHTTPGateway struct {
	baseURL    string
	httpClient *http.Client
}

// NewOrchestratorHTTPGateway cria uma nova instância
func NewOrchestratorHTTPGateway(baseURL string) *OrchestratorHTTPGateway {
	return &OrchestratorHTTPGateway{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// SubmitDiagram implementa a interface OrchestratorGateway
func (g *OrchestratorHTTPGateway) SubmitDiagram(ctx context.Context, diagram *entity.Diagram, requestID string) (entity.ProcessID, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		g.baseURL+"/internal/diagrams",
		bytes.NewReader(diagram.Content()),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.ContentLength = diagram.Size()
	req.Header.Set("Content-Type", diagram.MIMEType().String())
	req.Header.Set("X-Filename", diagram.Filename())
	req.Header.Set("X-Request-ID", requestID)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response struct {
		ProcessID string    `json:"process_id"`
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"created_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return entity.ProcessID(response.ProcessID), nil
}

// GetProcessStatus implementa a interface OrchestratorGateway
func (g *OrchestratorHTTPGateway) GetProcessStatus(ctx context.Context, processID entity.ProcessID, requestID string) (*domainGateway.ProcessStatusDTO, error) {
	url := fmt.Sprintf("%s/internal/process/%s/status", g.baseURL, processID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Request-ID", requestID)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var response struct {
		ProcessID string  `json:"process_id"`
		Status    string  `json:"status"`
		ReportID  *string `json:"report_id,omitempty"`
		Error     string  `json:"error,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	dto := &domainGateway.ProcessStatusDTO{
		ProcessID: entity.ProcessID(response.ProcessID),
		Status:    entity.ProcessStatus(response.Status),
		Error:     response.Error,
	}

	if response.ReportID != nil {
		reportID := entity.ReportID(*response.ReportID)
		dto.ReportID = &reportID
	}

	return dto, nil
}

// Garantir que implementa a interface (compile-time check)
var _ domainGateway.OrchestratorGateway = (*OrchestratorHTTPGateway)(nil)
