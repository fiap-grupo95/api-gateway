package errors

import "errors"

// Process errors
var (
	ErrEmptyProcessID             = errors.New("process ID cannot be empty")
	ErrInvalidProcessIDFormat     = errors.New("invalid process ID format")
	ErrInvalidStatusTransition    = errors.New("invalid status transition")
	ErrCannotFailCompletedProcess = errors.New("cannot fail a completed process")
)

// Diagram errors
var (
	ErrEmptyDiagramID  = errors.New("diagram ID cannot be empty")
	ErrEmptyFilename   = errors.New("filename cannot be empty")
	ErrEmptyContent    = errors.New("content cannot be empty")
	ErrInvalidMIMEType = errors.New("invalid MIME type")
	ErrFileTooLarge    = errors.New("file size exceeds maximum allowed")
)

// Report errors
var (
	ErrEmptyReportID = errors.New("report ID cannot be empty")
)

// Authentication errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrExpiredToken       = errors.New("token expired")
)

// Gateway errors
var (
	ErrGatewayUnavailable = errors.New("gateway unavailable")
	ErrInvalidResponse    = errors.New("invalid response from gateway")
)
