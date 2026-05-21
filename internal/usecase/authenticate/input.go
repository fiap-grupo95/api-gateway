package authenticate

import (
	"fmt"
)

// Input representa a entrada do use case de autenticação
type Input struct {
	Username string
	Password string
}

// Validate valida a entrada
func (i Input) Validate() error {
	if i.Username == "" {
		return fmt.Errorf("username is required")
	}
	if i.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(i.Username) < 1 || len(i.Username) > 64 {
		return fmt.Errorf("username must be between 1 and 64 characters")
	}
	if len(i.Password) < 4 || len(i.Password) > 128 {
		return fmt.Errorf("password must be between 4 and 128 characters")
	}
	return nil
}
