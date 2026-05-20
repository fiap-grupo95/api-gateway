package entity_test

import (
	"testing"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/entity"
	"github.com/fiap/secure-systems/api-gateway/internal/domain/errors"
)

func TestProcess_StatusTransitions(t *testing.T) {
	// Criar processo
	diagramID := entity.DiagramID("diagram-123")
	process, err := entity.NewProcess(diagramID)
	if err != nil {
		t.Fatalf("NewProcess failed: %v", err)
	}

	// Verificar estado inicial
	if process.Status() != entity.StatusPending {
		t.Errorf("Expected status %v, got %v", entity.StatusPending, process.Status())
	}

	// Transição válida: pending → processing
	err = process.StartProcessing()
	if err != nil {
		t.Errorf("StartProcessing failed: %v", err)
	}
	if process.Status() != entity.StatusProcessing {
		t.Errorf("Expected status %v, got %v", entity.StatusProcessing, process.Status())
	}

	// Transição inválida: processing → processing
	err = process.StartProcessing()
	if err != errors.ErrInvalidStatusTransition {
		t.Errorf("Expected ErrInvalidStatusTransition, got %v", err)
	}

	// Transição válida: processing → completed
	reportID := entity.ReportID("report-456")
	err = process.Complete(reportID)
	if err != nil {
		t.Errorf("Complete failed: %v", err)
	}
	if process.Status() != entity.StatusCompleted {
		t.Errorf("Expected status %v, got %v", entity.StatusCompleted, process.Status())
	}
	if process.ReportID() == nil || *process.ReportID() != reportID {
		t.Errorf("Expected report ID %v, got %v", reportID, process.ReportID())
	}

	// Transição inválida: completed → failed
	err = process.Fail()
	if err != errors.ErrCannotFailCompletedProcess {
		t.Errorf("Expected ErrCannotFailCompletedProcess, got %v", err)
	}
}

func TestProcess_CanTransitionTo(t *testing.T) {
	diagramID := entity.DiagramID("diagram-123")
	process, _ := entity.NewProcess(diagramID)

	tests := []struct {
		name       string
		from       entity.ProcessStatus
		to         entity.ProcessStatus
		canTransit bool
	}{
		{"pending to processing", entity.StatusPending, entity.StatusProcessing, true},
		{"pending to failed", entity.StatusPending, entity.StatusFailed, true},
		{"pending to completed", entity.StatusPending, entity.StatusCompleted, false},
		{"processing to completed", entity.StatusProcessing, entity.StatusCompleted, true},
		{"processing to failed", entity.StatusProcessing, entity.StatusFailed, true},
		{"processing to pending", entity.StatusProcessing, entity.StatusPending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reconstruir processo com status de origem
			p := entity.ReconstructProcess(
				process.ID(),
				diagramID,
				tt.from,
				nil,
				process.CreatedAt(),
				process.UpdatedAt(),
			)

			result := p.CanTransitionTo(tt.to)
			if result != tt.canTransit {
				t.Errorf("CanTransitionTo(%v → %v) = %v, want %v", tt.from, tt.to, result, tt.canTransit)
			}
		})
	}
}

func TestDiagram_NewDiagram(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		content     []byte
		mimeType    entity.MIMEType
		expectError bool
	}{
		{
			name:        "valid PNG",
			filename:    "test.png",
			content:     []byte{0x89, 0x50, 0x4E, 0x47},
			mimeType:    entity.MIMETypePNG,
			expectError: false,
		},
		{
			name:        "empty filename",
			filename:    "",
			content:     []byte("content"),
			mimeType:    entity.MIMETypePNG,
			expectError: true,
		},
		{
			name:        "empty content",
			filename:    "test.png",
			content:     []byte{},
			mimeType:    entity.MIMETypePNG,
			expectError: true,
		},
		{
			name:        "invalid MIME type",
			filename:    "test.txt",
			content:     []byte("content"),
			mimeType:    entity.MIMEType("text/plain"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagram, err := entity.NewDiagram(tt.filename, tt.content, tt.mimeType)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if diagram == nil {
					t.Errorf("Expected diagram, got nil")
				}
			}
		})
	}
}

func TestDiagram_FilenameSanitization(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{"normal filename", "test.png", "test.png"},
		{"path traversal", "../../../etc/passwd", "etc_passwd"},
		{"windows path", "C:\\Users\\test.png", "C:_Users_test.png"},
		{"empty after sanitization", "..", "diagram"},
		{"with forward slashes", "path/to/file.png", "path_to_file.png"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagram, err := entity.NewDiagram(tt.filename, []byte("content"), entity.MIMETypePNG)
			if err != nil {
				t.Fatalf("NewDiagram failed: %v", err)
			}

			if diagram.Filename() != tt.expected {
				t.Errorf("Filename = %v, want %v", diagram.Filename(), tt.expected)
			}
		})
	}
}

func TestProcessID_String(t *testing.T) {
	processID := entity.ProcessID("550e8400-e29b-41d4-a716-446655440000")
	expected := "550e8400-e29b-41d4-a716-446655440000"

	if processID.String() != expected {
		t.Errorf("String() = %v, want %v", processID.String(), expected)
	}
}

func TestProcessID_Validate(t *testing.T) {
	tests := []struct {
		name        string
		processID   entity.ProcessID
		expectError bool
	}{
		{"valid UUID", entity.ProcessID("550e8400-e29b-41d4-a716-446655440000"), false},
		{"empty UUID", entity.ProcessID(""), true},
		{"invalid format", entity.ProcessID("invalid-uuid"), true},
		{"uppercase UUID", entity.ProcessID("550E8400-E29B-41D4-A716-446655440000"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.processID.Validate()
			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestDiagramID_String(t *testing.T) {
	diagramID := entity.DiagramID("diagram-123")
	expected := "diagram-123"

	if diagramID.String() != expected {
		t.Errorf("String() = %v, want %v", diagramID.String(), expected)
	}
}

func TestDiagramID_Validate(t *testing.T) {
	tests := []struct {
		name        string
		diagramID   entity.DiagramID
		expectError bool
	}{
		{"valid ID", entity.DiagramID("diagram-123"), false},
		{"empty ID", entity.DiagramID(""), true},
		{"whitespace only", entity.DiagramID("   "), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.diagramID.Validate()
			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestReportID_String(t *testing.T) {
	reportID := entity.ReportID("report-456")
	expected := "report-456"

	if reportID.String() != expected {
		t.Errorf("String() = %v, want %v", reportID.String(), expected)
	}
}

func TestReportID_Validate(t *testing.T) {
	tests := []struct {
		name        string
		reportID    entity.ReportID
		expectError bool
	}{
		{"valid ID", entity.ReportID("report-456"), false},
		{"empty ID", entity.ReportID(""), true},
		{"whitespace only", entity.ReportID("   "), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.reportID.Validate()
			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestMIMEType_String(t *testing.T) {
	tests := []struct {
		name     string
		mimeType entity.MIMEType
		expected string
	}{
		{"PNG", entity.MIMETypePNG, "image/png"},
		{"JPEG", entity.MIMETypeJPEG, "image/jpeg"},
		{"PDF", entity.MIMETypePDF, "application/pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mimeType.String()
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestProcessStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status entity.ProcessStatus
		expected string
	}{
		{"pending", entity.StatusPending, "pending"},
		{"processing", entity.StatusProcessing, "processing"},
		{"completed", entity.StatusCompleted, "completed"},
		{"failed", entity.StatusFailed, "failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.String()
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}
