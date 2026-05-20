package service

// Logger define a interface para logging abstrata do domínio
// Implementações desta interface devem ser fornecidas via dependency injection
// para desacoplar o domain layer de frameworks específicos de logging
type Logger interface {
	// Debug registra mensagens de debug
	Debug(msg string, keysAndValues ...interface{})

	// Info registra mensagens informativas
	Info(msg string, keysAndValues ...interface{})

	// Warn registra mensagens de aviso
	Warn(msg string, keysAndValues ...interface{})

	// Error registra mensagens de erro
	Error(msg string, keysAndValues ...interface{})

	// Fatal registra mensagens críticas e pode encerrar a aplicação
	Fatal(msg string, keysAndValues ...interface{})
}
