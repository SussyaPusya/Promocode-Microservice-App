package domainerrors

import "fmt"

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed on field '%s': %s", e.Field, e.Message)
}
