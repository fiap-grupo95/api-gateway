package get_report_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/entity"
	"github.com/fiap/secure-systems/api-gateway/internal/domain/gateway"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/get_report"
)

// Mock do ReportGateway
type mockReportGateway struct {
	getReportFunc func(ctx context.Context, reportID entity.ReportID, requestID string) (*gateway.ReportDTO, error)
}

func (m *mockReportGateway) GetReport(ctx context.Context, reportID entity.ReportID, requestID string) (*gateway.ReportDTO, error) {
	if m.getReportFunc != nil {
		return m.getReportFunc(ctx, reportID, requestID)
	}
	return nil, nil
}

func TestGetReport_Success(t *testing.T) {
	// Arrange
	now := time.Now()
	mockGateway := &mockReportGateway{
		getReportFunc: func(ctx context.Context, reportID entity.ReportID, requestID string) (*gateway.ReportDTO, error) {
			return &gateway.ReportDTO{
				ReportID:        reportID,
				ProcessID:       entity.ProcessID("process-123"),
				Components:      []string{"API Gateway", "Database", "Cache"},
				Risks:           []string{"SQL Injection", "XSS"},
				Recommendations: []string{"Use parameterized queries", "Sanitize input"},
				CreatedAt:       now,
			}, nil
		},
	}

	uc := get_report.New(mockGateway)

	// Act
	output, err := uc.Execute(context.Background(), get_report.Input{
		ReportID: "report-456",
	}, "request-123")

	// Assert
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if output == nil {
		t.Fatal("Expected output, got nil")
	}
	if len(output.Components) != 3 {
		t.Errorf("Components count = %d, want 3", len(output.Components))
	}
	if len(output.Risks) != 2 {
		t.Errorf("Risks count = %d, want 2", len(output.Risks))
	}
	if len(output.Recommendations) != 2 {
		t.Errorf("Recommendations count = %d, want 2", len(output.Recommendations))
	}
	if output.CreatedAt != now {
		t.Errorf("CreatedAt = %v, want %v", output.CreatedAt, now)
	}
}

func TestGetReport_EmptyLists(t *testing.T) {
	mockGateway := &mockReportGateway{
		getReportFunc: func(ctx context.Context, reportID entity.ReportID, requestID string) (*gateway.ReportDTO, error) {
			return &gateway.ReportDTO{
				ReportID:        reportID,
				ProcessID:       entity.ProcessID("process-123"),
				Components:      []string{},
				Risks:           []string{},
				Recommendations: []string{},
				CreatedAt:       time.Now(),
			}, nil
		},
	}

	uc := get_report.New(mockGateway)

	output, err := uc.Execute(context.Background(), get_report.Input{
		ReportID: "report-456",
	}, "request-123")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if output == nil {
		t.Fatal("Expected output, got nil")
	}
	if len(output.Components) != 0 {
		t.Errorf("Components should be empty, got %d items", len(output.Components))
	}
}

func TestGetReport_GatewayError(t *testing.T) {
	mockGateway := &mockReportGateway{
		getReportFunc: func(ctx context.Context, reportID entity.ReportID, requestID string) (*gateway.ReportDTO, error) {
			return nil, errors.New("gateway unavailable")
		},
	}

	uc := get_report.New(mockGateway)

	output, err := uc.Execute(context.Background(), get_report.Input{
		ReportID: "report-456",
	}, "request-123")

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if output != nil {
		t.Errorf("Expected nil output, got %v", output)
	}
}

func TestGetReport_EmptyReportID(t *testing.T) {
	mockGateway := &mockReportGateway{}
	uc := get_report.New(mockGateway)

	output, err := uc.Execute(context.Background(), get_report.Input{
		ReportID: "",
	}, "request-123")

	if err == nil {
		t.Error("Expected error for empty report ID, got nil")
	}
	if output != nil {
		t.Errorf("Expected nil output, got %v", output)
	}
}

func TestGetReportInput_Validate(t *testing.T) {
	tests := []struct {
		name        string
		input       get_report.Input
		expectError bool
	}{
		{
			name:        "valid report ID",
			input:       get_report.Input{ReportID: "report-123"},
			expectError: false,
		},
		{
			name:        "empty report ID",
			input:       get_report.Input{ReportID: ""},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
