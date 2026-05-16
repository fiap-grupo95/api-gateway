//go:build ignore

// Gerador de JWT para testes do API Gateway.
// Uso: go run api-gateway/docs/generate_token.go
// Requer: JWT_SECRET definido no ambiente (fonte: .env)
//
// Exemplo:
//
//	source .env && go run api-gateway/docs/generate_token.go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		fmt.Fprintln(os.Stderr, "erro: JWT_SECRET não definido ou menor que 32 caracteres")
		fmt.Fprintln(os.Stderr, "dica: execute 'source .env' antes de rodar este script")
		os.Exit(1)
	}

	claims := jwt.MapClaims{
		"sub":  "test-user",
		"name": "Tester",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro ao assinar token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(signed)
}
