package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type ReportResponse struct {
	ReportID        string    `json:"report_id"`
	ProcessID       string    `json:"process_id"`
	Components      []string  `json:"components"`
	Risks           []string  `json:"risks"`
	Recommendations []string  `json:"recommendations"`
	CreatedAt       time.Time `json:"created_at"`
}

type ReportClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewReportClient(baseURL string) *ReportClient {
	return &ReportClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// GetReport busca o relatório pelo ID.
func (c *ReportClient) GetReport(ctx context.Context, reportID, requestID string) (*ReportResponse, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/internal/reports/%s", c.baseURL, reportID), nil)
	if err != nil {
		return nil, 0, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Request-ID", requestID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("report request: %w", err)
	}
	defer resp.Body.Close()

	var out ReportResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode response: %w", err)
	}
	return &out, resp.StatusCode, nil
}
