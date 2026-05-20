package upload_diagram_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/entity"
	"github.com/fiap/secure-systems/api-gateway/internal/domain/gateway"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/upload_diagram"
)

// Mock do OrchestratorGateway
type mockOrchestratorGateway struct {
	submitDiagramFunc func(ctx context.Context, diagram *entity.Diagram, requestID string) (entity.ProcessID, error)
}

func (m *mockOrchestratorGateway) SubmitDiagram(ctx context.Context, diagram *entity.Diagram, requestID string) (entity.ProcessID, error) {
	if m.submitDiagramFunc != nil {
		return m.submitDiagramFunc(ctx, diagram, requestID)
	}
	return entity.ProcessID("process-123"), nil
}

func (m *mockOrchestratorGateway) GetProcessStatus(ctx context.Context, processID entity.ProcessID, requestID string) (*gateway.ProcessStatusDTO, error) {
	return nil, nil
}

// Mock do MIMEValidator
type mockMIMEValidator struct {
	validateContentFunc func(content []byte) (entity.MIMEType, error)
	isAllowedFunc       func(mimeType entity.MIMEType) bool
}

func (m *mockMIMEValidator) ValidateContent(content []byte) (entity.MIMEType, error) {
	if m.validateContentFunc != nil {
		return m.validateContentFunc(content)
	}
	return entity.MIMETypePNG, nil
}

func (m *mockMIMEValidator) IsAllowed(mimeType entity.MIMEType) bool {
	if m.isAllowedFunc != nil {
		return m.isAllowedFunc(mimeType)
	}
	return true
}

func TestUploadDiagram_Success(t *testing.T) {
	// Arrange
	mockGateway := &mockOrchestratorGateway{
		submitDiagramFunc: func(ctx context.Context, diagram *entity.Diagram, requestID string) (entity.ProcessID, error) {
			return entity.ProcessID("process-123"), nil
		},
	}
	mockValidator := &mockMIMEValidator{
		validateContentFunc: func(content []byte) (entity.MIMEType, error) {
			return entity.MIMETypePNG, nil
		},
		isAllowedFunc: func(mimeType entity.MIMEType) bool {
			return true
		},
	}

	uc := upload_diagram.New(mockGateway, mockValidator, 10*1024*1024)

	// PNG magic bytes
	pngContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	// Act
	output, err := uc.Execute(context.Background(), upload_diagram.Input{
		Content:  pngContent,
		Filename: "test.png",
	}, "request-123")

	// Assert
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if output == nil {
		t.Fatal("Expected output, got nil")
	}
	if output.ProcessID != "process-123" {
		t.Errorf("ProcessID = %v, want %v", output.ProcessID, "process-123")
	}
	if output.Status != entity.StatusPending {
		t.Errorf("Status = %v, want %v", output.Status, entity.StatusPending)
	}
}

func TestUploadDiagram_InvalidMIME(t *testing.T) {
	mockGateway := &mockOrchestratorGateway{}
	mockValidator := &mockMIMEValidator{
		validateContentFunc: func(content []byte) (entity.MIMEType, error) {
			return entity.MIMEType("text/plain"), nil
		},
		isAllowedFunc: func(mimeType entity.MIMEType) bool {
			return false
		},
	}

	uc := upload_diagram.New(mockGateway, mockValidator, 10*1024*1024)

	content := []byte("not a valid image")

	output, err := uc.Execute(context.Background(), upload_diagram.Input{
		Content:  content,
		Filename: "test.txt",
	}, "request-123")

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if output != nil {
		t.Errorf("Expected nil output, got %v", output)
	}
}

func TestUploadDiagram_FileTooLarge(t *testing.T) {
	mockGateway := &mockOrchestratorGateway{}
	mockValidator := &mockMIMEValidator{}

	// Max 1KB
	uc := upload_diagram.New(mockGateway, mockValidator, 1024)

	// 2KB content
	largeContent := make([]byte, 2048)

	output, err := uc.Execute(context.Background(), upload_diagram.Input{
		Content:  largeContent,
		Filename: "large.png",
	}, "request-123")

	if err == nil {
		t.Error("Expected error for file too large, got nil")
	}
	if output != nil {
		t.Errorf("Expected nil output, got %v", output)
	}
}

func TestUploadDiagram_GatewayError(t *testing.T) {
	mockGateway := &mockOrchestratorGateway{
		submitDiagramFunc: func(ctx context.Context, diagram *entity.Diagram, requestID string) (entity.ProcessID, error) {
			return "", errors.New("gateway error")
		},
	}
	mockValidator := &mockMIMEValidator{
		validateContentFunc: func(content []byte) (entity.MIMEType, error) {
			return entity.MIMETypePNG, nil
		},
		isAllowedFunc: func(mimeType entity.MIMEType) bool {
			return true
		},
	}

	uc := upload_diagram.New(mockGateway, mockValidator, 10*1024*1024)

	pngContent := []byte{0x89, 0x50, 0x4E, 0x47}

	output, err := uc.Execute(context.Background(), upload_diagram.Input{
		Content:  pngContent,
		Filename: "test.png",
	}, "request-123")

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if output != nil {
		t.Errorf("Expected nil output, got %v", output)
	}
}
