package transport

import (
	"strings"

	"github.com/azizndao/grouter"
	"github.com/go-playground/validator/v10"
)

// ValidationError represents a single field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag"`
	Value   any    `json:"value,omitempty"`
}

// buildValidationError converts validator errors to structured ValidationError
func (t *Default) buildValidationError(validationErrors validator.ValidationErrors, locale string) *grouter.Error {
	translator := t.getTranslator(locale)
	errors := make([]ValidationError, 0, len(validationErrors))

	for _, fieldError := range validationErrors {
		errorValue := fieldError.Value()

		// Hide sensitive field values for security
		if t.isSensitiveField(fieldError.Field()) {
			errorValue = "[hidden]"
		}

		errors = append(errors, ValidationError{
			Field:   fieldError.Field(),
			Message: fieldError.Translate(translator),
			Tag:     fieldError.Tag(),
			Value:   errorValue,
		})
	}

	return grouter.ErrorUnprocessableEntity(errors, nil)
}

// isSensitiveField checks if a field contains sensitive information
func (t *Default) isSensitiveField(fieldName string) bool {
	sensitiveFields := []string{"password", "secret", "token", "key"}
	fieldLower := strings.ToLower(fieldName)

	for _, sensitive := range sensitiveFields {
		if strings.Contains(fieldLower, sensitive) {
			return true
		}
	}

	return false
}
