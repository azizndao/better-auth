package transport

import (
	"net/http"
	"strings"

	"better-auth/internal/models"

	"github.com/go-playground/validator/v10"
)

// buildValidationError converts validator errors to structured ValidationError
func (t *Default) buildValidationError(validationErrors validator.ValidationErrors, locale string) *APIError {
	translator := t.getTranslator(locale)
	errors := make([]models.ValidationError, 0, len(validationErrors))

	for _, fieldError := range validationErrors {
		errorValue := fieldError.Value()

		// Hide sensitive field values for security
		if t.isSensitiveField(fieldError.Field()) {
			errorValue = "[hidden]"
		}

		errors = append(errors, models.ValidationError{
			Field:   fieldError.Field(),
			Message: fieldError.Translate(translator),
			Tag:     fieldError.Tag(),
			Value:   errorValue,
		})
	}

	return &APIError{
		Code:     http.StatusUnprocessableEntity,
		Data:     errors,
		internal: nil,
	}
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

