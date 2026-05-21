package service_test

import (
	"testing"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/entity"
	"github.com/fiap/secure-systems/api-gateway/internal/infrastructure/service"
)

func TestMIMEValidator_ValidateContent_PNG(t *testing.T) {
	validator := service.NewMIMEValidator([]entity.MIMEType{
		entity.MIMETypePNG,
		entity.MIMETypeJPEG,
		entity.MIMETypePDF,
	})

	// PNG magic bytes
	pngContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	mimeType, err := validator.ValidateContent(pngContent)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if mimeType != entity.MIMETypePNG {
		t.Errorf("MIMEType = %v, want %v", mimeType, entity.MIMETypePNG)
	}
}

func TestMIMEValidator_ValidateContent_JPEG(t *testing.T) {
	validator := service.NewMIMEValidator([]entity.MIMEType{
		entity.MIMETypePNG,
		entity.MIMETypeJPEG,
		entity.MIMETypePDF,
	})

	// JPEG magic bytes (JFIF format)
	jpegContent := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}

	mimeType, err := validator.ValidateContent(jpegContent)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if mimeType != entity.MIMETypeJPEG {
		t.Errorf("MIMEType = %v, want %v", mimeType, entity.MIMETypeJPEG)
	}
}

func TestMIMEValidator_ValidateContent_PDF(t *testing.T) {
	validator := service.NewMIMEValidator([]entity.MIMEType{
		entity.MIMETypePNG,
		entity.MIMETypeJPEG,
		entity.MIMETypePDF,
	})

	// PDF magic bytes
	pdfContent := []byte("%PDF-1.4\n")

	mimeType, err := validator.ValidateContent(pdfContent)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if mimeType != entity.MIMETypePDF {
		t.Errorf("MIMEType = %v, want %v", mimeType, entity.MIMETypePDF)
	}
}

func TestMIMEValidator_ValidateContent_InvalidType(t *testing.T) {
	validator := service.NewMIMEValidator([]entity.MIMEType{
		entity.MIMETypePNG,
		entity.MIMETypeJPEG,
		entity.MIMETypePDF,
	})

	// Plain text content
	textContent := []byte("This is plain text")

	mimeType, err := validator.ValidateContent(textContent)

	// http.DetectContentType retorna "text/plain; charset=utf-8"
	// que não é um MIMEType válido do domínio
	if err == nil {
		t.Error("Expected error for invalid MIME type, got nil")
	}
	if mimeType != "" {
		t.Errorf("Expected empty MIMEType, got %v", mimeType)
	}
}

func TestMIMEValidator_IsAllowed_PNG(t *testing.T) {
	validator := service.NewMIMEValidator([]entity.MIMEType{
		entity.MIMETypePNG,
		entity.MIMETypeJPEG,
	})

	if !validator.IsAllowed(entity.MIMETypePNG) {
		t.Error("PNG should be allowed")
	}
}

func TestMIMEValidator_IsAllowed_NotAllowed(t *testing.T) {
	validator := service.NewMIMEValidator([]entity.MIMEType{
		entity.MIMETypePNG,
		entity.MIMETypeJPEG,
	})

	if validator.IsAllowed(entity.MIMETypePDF) {
		t.Error("PDF should not be allowed")
	}
}

func TestMIMEValidator_IsAllowed_EmptyList(t *testing.T) {
	validator := service.NewMIMEValidator([]entity.MIMEType{})

	if validator.IsAllowed(entity.MIMETypePNG) {
		t.Error("No MIME types should be allowed with empty list")
	}
}

func TestMIMEValidator_ValidateContent_SmallContent(t *testing.T) {
	validator := service.NewMIMEValidator([]entity.MIMEType{
		entity.MIMETypePNG,
	})

	// Conteúdo menor que 512 bytes mas com PNG magic bytes completos
	smallPNG := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	mimeType, err := validator.ValidateContent(smallPNG)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if mimeType != entity.MIMETypePNG {
		t.Errorf("MIMEType = %v, want %v", mimeType, entity.MIMETypePNG)
	}
}

func TestMIMEValidator_ValidateContent_LargeContent(t *testing.T) {
	validator := service.NewMIMEValidator([]entity.MIMEType{
		entity.MIMETypePNG,
	})

	// Conteúdo maior que 512 bytes
	largePNG := make([]byte, 1024)
	// PNG magic bytes no início
	copy(largePNG, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})

	mimeType, err := validator.ValidateContent(largePNG)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if mimeType != entity.MIMETypePNG {
		t.Errorf("MIMEType = %v, want %v", mimeType, entity.MIMETypePNG)
	}
}

func TestMIMEValidator_MultipleTypes(t *testing.T) {
	tests := []struct {
		name      string
		allowed   []entity.MIMEType
		mimeType  entity.MIMEType
		shouldAllow bool
	}{
		{
			name:        "PNG allowed in list",
			allowed:     []entity.MIMEType{entity.MIMETypePNG, entity.MIMETypeJPEG},
			mimeType:    entity.MIMETypePNG,
			shouldAllow: true,
		},
		{
			name:        "JPEG allowed in list",
			allowed:     []entity.MIMEType{entity.MIMETypePNG, entity.MIMETypeJPEG},
			mimeType:    entity.MIMETypeJPEG,
			shouldAllow: true,
		},
		{
			name:        "PDF not in list",
			allowed:     []entity.MIMEType{entity.MIMETypePNG, entity.MIMETypeJPEG},
			mimeType:    entity.MIMETypePDF,
			shouldAllow: false,
		},
		{
			name:        "Only PDF allowed",
			allowed:     []entity.MIMEType{entity.MIMETypePDF},
			mimeType:    entity.MIMETypePNG,
			shouldAllow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := service.NewMIMEValidator(tt.allowed)
			result := validator.IsAllowed(tt.mimeType)
			if result != tt.shouldAllow {
				t.Errorf("IsAllowed(%v) = %v, want %v", tt.mimeType, result, tt.shouldAllow)
			}
		})
	}
}
