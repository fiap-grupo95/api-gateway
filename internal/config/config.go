package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                  string
	UploadOrchestratorURL string
	ReportServiceURL      string
	MaxUploadSizeMB       int64
	AllowedMIMETypes      []string
	JWTSecret             []byte
}

func Load() (*Config, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	maxMB, err := strconv.ParseInt(getEnvOrDefault("MAX_UPLOAD_SIZE_MB", "10"), 10, 64)
	if err != nil || maxMB <= 0 || maxMB > 50 {
		return nil, fmt.Errorf("MAX_UPLOAD_SIZE_MB must be between 1 and 50")
	}

	mimeTypes := strings.Split(getEnvOrDefault("ALLOWED_MIME_TYPES", "image/png,image/jpeg,application/pdf"), ",")

	return &Config{
		Port:                  getEnvOrDefault("PORT", "8080"),
		UploadOrchestratorURL: requireEnv("UPLOAD_ORCHESTRATOR_URL"),
		ReportServiceURL:      requireEnv("REPORT_SERVICE_URL"),
		MaxUploadSizeMB:       maxMB,
		AllowedMIMETypes:      mimeTypes,
		JWTSecret:             []byte(jwtSecret),
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
