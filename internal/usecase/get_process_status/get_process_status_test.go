package get_process_status_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/entity"
	"github.com/fiap/secure-systems/api-gateway/internal/domain/gateway"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/get_process_status"
)

// Mock do OrchestratorGateway
type mockOrchestratorGateway struct {
	getProcessStatusFunc func(ctx context.Context, processID entity.ProcessID, requestID string) (*gateway.ProcessStatusDTO, error)
}

func (m *mockOrchestratorGateway) SubmitDiagram(ctx context.Context, diagram *entity.Diagram, requestID string) (entity.ProcessID, error) {
	return "", nil
}

func (m *mockOrchestratorGateway) GetProcessStatus(ctx context.Context, processID entity.ProcessID, requestID string) (*gateway.ProcessStatusDTO, error) {
	if m.getProcessStatusFunc != nil {
		return m.getProcessStatusFunc(ctx, processID, requestID)
	}
	return nil, nil
}

func TestGetProcessStatus_Success(t *testing.T) {
	// Arrange
	reportID := entity.ReportID("report-456")
	mockGateway := &mockOrchestratorGateway{
		getProcessStatusFunc: func(ctx context.Context, processID entity.ProcessID, requestID string) (*gateway.ProcessStatusDTO, error) {
			return &gateway.ProcessStatusDTO{
				ProcessID: processID,
				Status:    entity.StatusCompleted,
				ReportID:  &reportID,
				Error:     "",
			}, nil
		},
	}

	uc := get_process_status.New(mockGateway)

	// Act
	output, err := uc.Execute(context.Background(), get_process_status.Input{
		ProcessID: "550e8400-e29b-41d4-a716-446655440000",
	}, "request-123")

	// Assert
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if output == nil {
		t.Fatal("Expected output, got nil")
	}
	if output.Status != entity.StatusCompleted {
		t.Errorf("Status = %v, want %v", output.Status, entity.StatusCompleted)
	}
	if output.ReportID == nil {
		t.Error("Expected report ID, got nil")
	}
}

func TestGetProcessStatus_Pending(t *testing.T) {
	mockGateway := &mockOrchestratorGateway{
		getProcessStatusFunc: func(ctx context.Context, processID entity.ProcessID, requestID string) (*gateway.ProcessStatusDTO, error) {
			return &gateway.ProcessStatusDTO{
				ProcessID: processID,
				Status:    entity.StatusPending,
				ReportID:  nil,
				Error:     "",
			}, nil
		},
	}

	uc := get_process_status.New(mockGateway)

	output, err := uc.Execute(context.Background(), get_process_status.Input{
		ProcessID: "550e8400-e29b-41d4-a716-446655440000",
	}, "request-123")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if output.Status != entity.StatusPending {
		t.Errorf("Status = %v, want %v", output.Status, entity.StatusPending)
	}
	if output.ReportID != nil {
		t.Error("Expected nil report ID for pending status")
	}
}

func TestGetProcessStatus_Failed(t *testing.T) {
	mockGateway := &mockOrchestratorGateway{
		getProcessStatusFunc: func(ctx context.Context, processID entity.ProcessID, requestID string) (*gateway.ProcessStatusDTO, error) {
			return &gateway.ProcessStatusDTO{
				ProcessID: processID,
				Status:    entity.StatusFailed,
				ReportID:  nil,
				Error:     "processing failed",
			}, nil
		},
	}

	uc := get_process_status.New(mockGateway)

	output, err := uc.Execute(context.Background(), get_process_status.Input{
		ProcessID: "550e8400-e29b-41d4-a716-446655440000",
	}, "request-123")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if output.Status != entity.StatusFailed {
		t.Errorf("Status = %v, want %v", output.Status, entity.StatusFailed)
	}
	if output.Error != "processing failed" {
		t.Errorf("Error = %v, want %v", output.Error, "processing failed")
	}
}

func TestGetProcessStatus_GatewayError(t *testing.T) {
	mockGateway := &mockOrchestratorGateway{
		getProcessStatusFunc: func(ctx context.Context, processID entity.ProcessID, requestID string) (*gateway.ProcessStatusDTO, error) {
			return nil, errors.New("gateway unavailable")
		},
	}

	uc := get_process_status.New(mockGateway)

	output, err := uc.Execute(context.Background(), get_process_status.Input{
		ProcessID: "550e8400-e29b-41d4-a716-446655440000",
	}, "request-123")

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if output != nil {
		t.Errorf("Expected nil output, got %v", output)
	}
}

func TestGetProcessStatus_EmptyProcessID(t *testing.T) {
	mockGateway := &mockOrchestratorGateway{}
	uc := get_process_status.New(mockGateway)

	output, err := uc.Execute(context.Background(), get_process_status.Input{
		ProcessID: "",
	}, "request-123")

	if err == nil {
		t.Error("Expected error for empty process ID, got nil")
	}
	if output != nil {
		t.Errorf("Expected nil output, got %v", output)
	}
}

func TestGetProcessStatus_InvalidProcessIDFormat(t *testing.T) {
	mockGateway := &mockOrchestratorGateway{}
	uc := get_process_status.New(mockGateway)

	output, err := uc.Execute(context.Background(), get_process_status.Input{
		ProcessID: "invalid-uuid-format",
	}, "request-123")

	if err == nil {
		t.Error("Expected error for invalid process ID format, got nil")
	}
	if output != nil {
		t.Errorf("Expected nil output, got %v", output)
	}
}

func TestGetProcessStatusInput_Validate(t *testing.T) {
	tests := []struct {
		name        string
		input       get_process_status.Input
		expectError bool
	}{
		{
			name:        "valid UUID",
			input:       get_process_status.Input{ProcessID: "550e8400-e29b-41d4-a716-446655440000"},
			expectError: false,
		},
		{
			name:        "empty process ID",
			input:       get_process_status.Input{ProcessID: ""},
			expectError: true,
		},
		{
			name:        "invalid UUID format",
			input:       get_process_status.Input{ProcessID: "not-a-uuid"},
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
