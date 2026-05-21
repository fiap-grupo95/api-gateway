package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/entity"
)

// Config representa a configuração da aplicação
type Config struct {
	Port                  string
	UploadOrchestratorURL string
	ReportServiceURL      string
	MaxUploadSizeMB       int64
	AllowedMIMETypes      []entity.MIMEType
	JWTSecret             []byte
	AuthUsername          string
	AuthPasswordHash      string
}

// Load carrega a configuração das variáveis de ambiente
func Load() (*Config, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	authUsername := os.Getenv("AUTH_USERNAME")
	if authUsername == "" {
		return nil, fmt.Errorf("AUTH_USERNAME is required")
	}

	authPasswordHash := os.Getenv("AUTH_PASSWORD_HASH")
	if authPasswordHash == "" {
		return nil, fmt.Errorf("AUTH_PASSWORD_HASH is required")
	}

	maxMB, err := strconv.ParseInt(getEnvOrDefault("MAX_UPLOAD_SIZE_MB", "10"), 10, 64)
	if err != nil || maxMB <= 0 || maxMB > 50 {
		return nil, fmt.Errorf("MAX_UPLOAD_SIZE_MB must be between 1 and 50")
	}

	mimeTypesStr := strings.Split(getEnvOrDefault("ALLOWED_MIME_TYPES", "image/png,image/jpeg,application/pdf"), ",")
	var mimeTypes []entity.MIMEType
	for _, mt := range mimeTypesStr {
		mimeType := entity.MIMEType(strings.TrimSpace(mt))
		if err := mimeType.Validate(); err != nil {
			return nil, fmt.Errorf("invalid MIME type %q in ALLOWED_MIME_TYPES: %w", mt, err)
		}
		mimeTypes = append(mimeTypes, mimeType)
	}

	return &Config{
		Port:                  getEnvOrDefault("PORT", "8080"),
		UploadOrchestratorURL: requireEnv("UPLOAD_ORCHESTRATOR_URL"),
		ReportServiceURL:      requireEnv("REPORT_SERVICE_URL"),
		MaxUploadSizeMB:       maxMB,
		AllowedMIMETypes:      mimeTypes,
		JWTSecret:             []byte(jwtSecret),
		AuthUsername:          authUsername,
		AuthPasswordHash:      authPasswordHash,
	}, nil
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return v
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
