package entity

import (
	"strings"
	"time"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/errors"
	"github.com/google/uuid"
)

// Process representa o processamento de um diagrama no sistema
type Process struct {
	id        ProcessID
	diagramID DiagramID
	status    ProcessStatus
	reportID  *ReportID
	createdAt time.Time
	updatedAt time.Time
}

// NewProcess cria uma nova instância de Process com validação
func NewProcess(diagramID DiagramID) (*Process, error) {
	if err := diagramID.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Process{
		id:        ProcessID(uuid.New().String()),
		diagramID: diagramID,
		status:    StatusPending,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// ReconstructProcess reconstrói um Process do repositório
func ReconstructProcess(id ProcessID, diagramID DiagramID, status ProcessStatus, reportID *ReportID, createdAt, updatedAt time.Time) *Process {
	return &Process{
		id:        id,
		diagramID: diagramID,
		status:    status,
		reportID:  reportID,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// ═══════════════════════════════════════════════════════════════
// REGRAS DE NEGÓCIO (métodos que alteram o estado)
// ═══════════════════════════════════════════════════════════════

// StartProcessing inicia o processamento
func (p *Process) StartProcessing() error {
	if p.status != StatusPending {
		return errors.ErrInvalidStatusTransition
	}
	p.status = StatusProcessing
	p.updatedAt = time.Now()
	return nil
}

// Complete marca o processo como completo
func (p *Process) Complete(reportID ReportID) error {
	if p.status != StatusProcessing {
		return errors.ErrInvalidStatusTransition
	}
	if err := reportID.Validate(); err != nil {
		return err
	}
	p.status = StatusCompleted
	p.reportID = &reportID
	p.updatedAt = time.Now()
	return nil
}

// Fail marca o processo como falho
func (p *Process) Fail() error {
	if p.status == StatusCompleted {
		return errors.ErrCannotFailCompletedProcess
	}
	p.status = StatusFailed
	p.updatedAt = time.Now()
	return nil
}

// CanTransitionTo verifica se uma transição de status é válida
func (p *Process) CanTransitionTo(newStatus ProcessStatus) bool {
	validTransitions := map[ProcessStatus][]ProcessStatus{
		StatusPending:    {StatusProcessing, StatusFailed},
		StatusProcessing: {StatusCompleted, StatusFailed},
		StatusCompleted:  {},
		StatusFailed:     {StatusPending},
	}

	allowed, exists := validTransitions[p.status]
	if !exists {
		return false
	}

	for _, s := range allowed {
		if s == newStatus {
			return true
		}
	}
	return false
}

// ═══════════════════════════════════════════════════════════════
// GETTERS (encapsulamento)
// ═══════════════════════════════════════════════════════════════

func (p *Process) ID() ProcessID         { return p.id }
func (p *Process) DiagramID() DiagramID  { return p.diagramID }
func (p *Process) Status() ProcessStatus { return p.status }
func (p *Process) ReportID() *ReportID   { return p.reportID }
func (p *Process) CreatedAt() time.Time  { return p.createdAt }
func (p *Process) UpdatedAt() time.Time  { return p.updatedAt }

// ═══════════════════════════════════════════════════════════════
// VALUE OBJECTS
// ═══════════════════════════════════════════════════════════════

// ProcessID é um identificador único de processo
type ProcessID string

func (p ProcessID) Validate() error {
	if p == "" {
		return errors.ErrEmptyProcessID
	}
	_, err := uuid.Parse(string(p))
	if err != nil {
		return errors.ErrInvalidProcessIDFormat
	}
	return nil
}

func (p ProcessID) String() string {
	return string(p)
}

// DiagramID é um identificador único de diagrama
type DiagramID string

func (d DiagramID) Validate() error {
	if d == "" {
		return errors.ErrEmptyDiagramID
	}
	if len(strings.TrimSpace(string(d))) == 0 {
		return errors.ErrEmptyDiagramID
	}
	return nil
}

func (d DiagramID) String() string {
	return string(d)
}

// ReportID é um identificador único de relatório
type ReportID string

func (r ReportID) Validate() error {
	if r == "" {
		return errors.ErrEmptyReportID
	}
	if len(strings.TrimSpace(string(r))) == 0 {
		return errors.ErrEmptyReportID
	}
	return nil
}

func (r ReportID) String() string {
	return string(r)
}

// ProcessStatus representa o status de um processo
type ProcessStatus string

const (
	StatusPending    ProcessStatus = "RECEBIDO"
	StatusProcessing ProcessStatus = "EM_PROCESSAMENTO"
	StatusCompleted  ProcessStatus = "ANALISADO"
	StatusFailed     ProcessStatus = "ERRO"
)

func (s ProcessStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusProcessing, StatusCompleted, StatusFailed:
		return true
	default:
		return false
	}
}

func (s ProcessStatus) String() string {
	return string(s)
}
