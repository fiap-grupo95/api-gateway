package authenticate

import (
	"context"
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/fiap/secure-systems/api-gateway/internal/domain/errors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// UseCase representa o caso de uso de autenticação
type UseCase struct {
	username     string
	passwordHash []byte
	jwtSecret    []byte
}

// New cria uma nova instância do use case
func New(username, passwordHash string, jwtSecret []byte) *UseCase {
	return &UseCase{
		username:     username,
		passwordHash: []byte(passwordHash),
		jwtSecret:    jwtSecret,
	}
}

// Execute executa o caso de uso
func (uc *UseCase) Execute(ctx context.Context, input Input) (*Output, error) {
	// 1. Validar input
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// 2. Verificar credenciais (timing attack protection)
	usernameMatch := subtle.ConstantTimeCompare([]byte(input.Username), []byte(uc.username)) == 1
	passwordErr := bcrypt.CompareHashAndPassword(uc.passwordHash, []byte(input.Password))

	if !usernameMatch || passwordErr != nil {
		return nil, errors.ErrInvalidCredentials
	}

	// 3. Gerar JWT token
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": input.Username,
		"iat": now.Unix(),
		"exp": now.Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(uc.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign jwt: %w", err)
	}

	// 4. Retornar output
	return &Output{
		Token:     signed,
		ExpiresIn: 3600, // 1 hora
	}, nil
}
