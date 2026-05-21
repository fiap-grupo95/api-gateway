package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/entity"
	domainGateway "github.com/fiap/secure-systems/api-gateway/internal/domain/gateway"
)

// ReportHTTPGateway é a implementação HTTP do ReportGateway
type ReportHTTPGateway struct {
	baseURL    string
	httpClient *http.Client
}

// NewReportHTTPGateway cria uma nova instância
func NewReportHTTPGateway(baseURL string) *ReportHTTPGateway {
	return &ReportHTTPGateway{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// GetReport implementa a interface ReportGateway
func (g *ReportHTTPGateway) GetReport(ctx context.Context, reportID entity.ReportID, requestID string) (*domainGateway.ReportDTO, error) {
	url := fmt.Sprintf("%s/internal/reports/%s", g.baseURL, reportID)

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
		ReportID        string    `json:"report_id"`
		ProcessID       string    `json:"process_id"`
		Components      []string  `json:"components"`
		Risks           []string  `json:"risks"`
		Recommendations []string  `json:"recommendations"`
		CreatedAt       time.Time `json:"created_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &domainGateway.ReportDTO{
		ReportID:        entity.ReportID(response.ReportID),
		ProcessID:       entity.ProcessID(response.ProcessID),
		Components:      response.Components,
		Risks:           response.Risks,
		Recommendations: response.Recommendations,
		CreatedAt:       response.CreatedAt,
	}, nil
}

// Garantir que implementa a interface (compile-time check)
var _ domainGateway.ReportGateway = (*ReportHTTPGateway)(nil)
