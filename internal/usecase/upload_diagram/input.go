package upload_diagram

import (
	"fmt"
)

// Input representa a entrada do use case
type Input struct {
	Content  []byte
	Filename string
}

// Validate valida a entrada (validações básicas de DTO)
func (i Input) Validate() error {
	if len(i.Content) == 0 {
		return fmt.Errorf("content is required")
	}
	if i.Filename == "" {
		return fmt.Errorf("filename is required")
	}
	return nil
}
