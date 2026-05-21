package authenticate_test

import (
	"context"
	"testing"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/errors"
	"github.com/fiap/secure-systems/api-gateway/internal/usecase/authenticate"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthenticate_Success(t *testing.T) {
	// Arrange
	username := "admin"
	password := "password123"
	
	// Gerar hash bcrypt para teste
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to generate hash: %v", err)
	}

	jwtSecret := []byte("test-secret-with-minimum-32-chars-required")
	uc := authenticate.New(username, string(hash), jwtSecret)

	// Act
	output, err := uc.Execute(context.Background(), authenticate.Input{
		Username: username,
		Password: password,
	})

	// Assert
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if output == nil {
		t.Fatal("Expected output, got nil")
	}
	if output.Token == "" {
		t.Error("Expected token, got empty string")
	}
	if output.ExpiresIn != 3600 {
		t.Errorf("ExpiresIn = %d, want 3600", output.ExpiresIn)
	}
}

func TestAuthenticate_InvalidUsername(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	jwtSecret := []byte("test-secret-with-minimum-32-chars-required")
	uc := authenticate.New("admin", string(hash), jwtSecret)

	output, err := uc.Execute(context.Background(), authenticate.Input{
		Username: "wrong-user",
		Password: "password123",
	})

	if err != errors.ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
	if output != nil {
		t.Errorf("Expected nil output, got %v", output)
	}
}

func TestAuthenticate_InvalidPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	jwtSecret := []byte("test-secret-with-minimum-32-chars-required")
	uc := authenticate.New("admin", string(hash), jwtSecret)

	output, err := uc.Execute(context.Background(), authenticate.Input{
		Username: "admin",
		Password: "wrong-password",
	})

	if err != errors.ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got %v", err)
	}
	if output != nil {
		t.Errorf("Expected nil output, got %v", output)
	}
}

func TestAuthenticate_EmptyUsername(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	jwtSecret := []byte("test-secret-with-minimum-32-chars-required")
	uc := authenticate.New("admin", string(hash), jwtSecret)

	output, err := uc.Execute(context.Background(), authenticate.Input{
		Username: "",
		Password: "password123",
	})

	if err == nil {
		t.Error("Expected error for empty username, got nil")
	}
	if output != nil {
		t.Errorf("Expected nil output, got %v", output)
	}
}

func TestAuthenticate_EmptyPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	jwtSecret := []byte("test-secret-with-minimum-32-chars-required")
	uc := authenticate.New("admin", string(hash), jwtSecret)

	output, err := uc.Execute(context.Background(), authenticate.Input{
		Username: "admin",
		Password: "",
	})

	if err == nil {
		t.Error("Expected error for empty password, got nil")
	}
	if output != nil {
		t.Errorf("Expected nil output, got %v", output)
	}
}

func TestAuthenticate_UsernameTooLong(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	jwtSecret := []byte("test-secret-with-minimum-32-chars-required")
	uc := authenticate.New("admin", string(hash), jwtSecret)

	// Username com 65 caracteres (máximo é 64)
	longUsername := "a123456789012345678901234567890123456789012345678901234567890123"

	output, err := uc.Execute(context.Background(), authenticate.Input{
		Username: longUsername,
		Password: "password123",
	})

	if err == nil {
		t.Error("Expected error for username too long, got nil")
	}
	if output != nil {
		t.Errorf("Expected nil output, got %v", output)
	}
}

func TestAuthenticate_PasswordTooShort(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	jwtSecret := []byte("test-secret-with-minimum-32-chars-required")
	uc := authenticate.New("admin", string(hash), jwtSecret)

	output, err := uc.Execute(context.Background(), authenticate.Input{
		Username: "admin",
		Password: "123", // Menos de 4 caracteres
	})

	if err == nil {
		t.Error("Expected error for password too short, got nil")
	}
	if output != nil {
		t.Errorf("Expected nil output, got %v", output)
	}
}

func TestAuthenticateInput_Validate(t *testing.T) {
	tests := []struct {
		name        string
		input       authenticate.Input
		expectError bool
	}{
		{
			name:        "valid input",
			input:       authenticate.Input{Username: "admin", Password: "pass1234"},
			expectError: false,
		},
		{
			name:        "empty username",
			input:       authenticate.Input{Username: "", Password: "pass1234"},
			expectError: true,
		},
		{
			name:        "empty password",
			input:       authenticate.Input{Username: "admin", Password: ""},
			expectError: true,
		},
		{
			name:        "username too long",
			input:       authenticate.Input{Username: "a1234567890123456789012345678901234567890123456789012345678901234", Password: "pass1234"},
			expectError: true,
		},
		{
			name:        "password too short",
			input:       authenticate.Input{Username: "admin", Password: "abc"},
			expectError: true,
		},
		{
			name:        "password too long",
			input:       authenticate.Input{Username: "admin", Password: "a12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678"},
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
