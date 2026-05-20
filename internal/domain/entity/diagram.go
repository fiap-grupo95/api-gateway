package entity

import (
	"fmt"
	"strings"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/errors"
	"github.com/google/uuid"
)

// Diagram representa um diagrama a ser processado
type Diagram struct {
	id       DiagramID
	filename string
	content  []byte
	mimeType MIMEType
	size     int64
}

// NewDiagram cria uma nova instância de Diagram com validação
func NewDiagram(filename string, content []byte, mimeType MIMEType) (*Diagram, error) {
	if filename == "" {
		return nil, errors.ErrEmptyFilename
	}
	if len(content) == 0 {
		return nil, errors.ErrEmptyContent
	}
	if err := mimeType.Validate(); err != nil {
		return nil, err
	}

	sanitized := sanitizeFilename(filename)

	return &Diagram{
		id:       DiagramID(uuid.New().String()),
		filename: sanitized,
		content:  content,
		mimeType: mimeType,
		size:     int64(len(content)),
	}, nil
}

// ═══════════════════════════════════════════════════════════════
// GETTERS (encapsulamento)
// ═══════════════════════════════════════════════════════════════

func (d *Diagram) ID() DiagramID      { return d.id }
func (d *Diagram) Filename() string   { return d.filename }
func (d *Diagram) Content() []byte    { return d.content }
func (d *Diagram) MIMEType() MIMEType { return d.mimeType }
func (d *Diagram) Size() int64        { return d.size }

// ═══════════════════════════════════════════════════════════════
// VALUE OBJECTS
// ═══════════════════════════════════════════════════════════════

// MIMEType representa um tipo MIME válido
type MIMEType string

const (
	MIMETypePNG  MIMEType = "image/png"
	MIMETypeJPEG MIMEType = "image/jpeg"
	MIMETypePDF  MIMEType = "application/pdf"
)

func (m MIMEType) Validate() error {
	switch m {
	case MIMETypePNG, MIMETypeJPEG, MIMETypePDF:
		return nil
	default:
		return fmt.Errorf("%w: %s", errors.ErrInvalidMIMEType, m)
	}
}

func (m MIMEType) String() string {
	return string(m)
}

// ═══════════════════════════════════════════════════════════════
// FUNÇÕES AUXILIARES (regras de negócio de domain)
// ═══════════════════════════════════════════════════════════════

func sanitizeFilename(name string) string {
	// Remove path components
	name = strings.ReplaceAll(name, "../", "")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")

	// Limita tamanho
	if len(name) > 255 {
		name = name[:255]
	}

	if name == "" || name == "." || name == ".." {
		return "diagram"
	}

	return name
}
