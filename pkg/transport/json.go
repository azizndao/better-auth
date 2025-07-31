package transport

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// DecodeJSON decodes JSON request body with automatic language detection and validation
func (t *Default) DecodeJSON(req *http.Request, v any) error {
	locale := t.detectLocale(req)
	return t.DecodeJSONWithLocale(req, v, locale)
}

// DecodeJSONWithLocale decodes JSON request body with explicit locale and validation
func (t *Default) DecodeJSONWithLocale(req *http.Request, v any, locale string) error {
	// Decode JSON first
	if err := json.NewDecoder(req.Body).Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	
	// Validate the decoded struct
	if err := t.validator.Struct(v); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return t.buildValidationError(validationErrors, locale)
		}
		return fmt.Errorf("validation failed: %w", err)
	}
	
	return nil
}

// RespondJSON sends a JSON response with the specified status code
func (t *Default) RespondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set(HeaderContentType, ContentTypeJSON)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}