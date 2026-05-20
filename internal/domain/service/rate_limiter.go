package service

// RateLimiter define a interface para limitar a taxa de requisições
// Implementações desta interface devem ser fornecidas via dependency injection
// para desacoplar o domain layer de bibliotecas específicas de rate limiting
type RateLimiter interface {
	// Allow retorna true se a requisição é permitida dentro do limite de taxa
	// userID identifica o cliente/usuário para aplicar o limite
	Allow(userID string) bool
}
