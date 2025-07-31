package transport

import (
	"fmt"
	"net/http"

	"better-auth/internal/models"
)

// RespondError sends a simple error response
func (t *Default) RespondError(w http.ResponseWriter, status int, message string) {
	t.RespondJSON(w, status, map[string]string{"error": message})
}

// RespondValidationError sends a structured validation error response
func (t *Default) RespondValidationError(w http.ResponseWriter, validationErr *ValidationError) {
	response := models.ValidationErrorResponse{
		Error:   "Validation Failed",
		Code:    http.StatusBadRequest,
		Type:    "validation_error",
		Errors:  validationErr.Errors,
		Message: fmt.Sprintf("Request validation failed with %d error(s)", len(validationErr.Errors)),
	}
	t.RespondJSON(w, http.StatusBadRequest, response)
}

// WriteError is an alias for RespondError for backward compatibility
func (t *Default) WriteError(w http.ResponseWriter, status int, message string) {
	t.RespondError(w, status, message)
}