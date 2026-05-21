package authenticate

// Output representa a saída do use case de autenticação
type Output struct {
	Token     string
	ExpiresIn int
}
