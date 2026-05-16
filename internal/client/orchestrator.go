package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type UploadResponse struct {
	ProcessID string    `json:"process_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type StatusResponse struct {
	ProcessID string `json:"process_id"`
	Status    string `json:"status"`
	ReportID  string `json:"report_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

type OrchestratorClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewOrchestratorClient(baseURL string) *OrchestratorClient {
	return &OrchestratorClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // uploads podem ser grandes
		},
	}
}

// Upload envia o diagrama ao upload-orchestrator e retorna o processID.
// Passa Content-Length explícito para que o MinIO receba o tamanho correto.
func (c *OrchestratorClient) Upload(ctx context.Context, content []byte, filename, contentType, requestID string) (*UploadResponse, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/diagrams", bytes.NewReader(content))
	if err != nil {
		return nil, 0, fmt.Errorf("build request: %w", err)
	}
	req.ContentLength = int64(len(content))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Filename", filename)
	req.Header.Set("X-Request-ID", requestID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, resp.StatusCode, fmt.Errorf("orchestrator returned %d: %s", resp.StatusCode, body)
	}

	var out UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode response: %w", err)
	}
	return &out, resp.StatusCode, nil
}

// GetStatus consulta o status de processamento de um diagrama.
func (c *OrchestratorClient) GetStatus(ctx context.Context, processID, requestID string) (*StatusResponse, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/internal/process/%s/status", c.baseURL, processID), nil)
	if err != nil {
		return nil, 0, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("X-Request-ID", requestID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("status request: %w", err)
	}
	defer resp.Body.Close()

	var out StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decode response: %w", err)
	}
	return &out, resp.StatusCode, nil
}
