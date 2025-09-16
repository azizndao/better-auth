package transport

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"better-auth/internal/models"
)

func TestDecodeJSON_WithValidation(t *testing.T) {
	transport := NewDefault()

	tests := []struct {
		name      string
		json      string
		model     any
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid signup request",
			json:      `{"email":"test@example.com","password":"password123","name":"Test User"}`,
			model:     &models.SignUpPayload{},
			wantError: false,
		},
		{
			name:      "invalid email",
			json:      `{"email":"invalid-email","password":"password123","name":"Test User"}`,
			model:     &models.SignUpPayload{},
			wantError: true,
			errorMsg:  "Email must be a valid email address",
		},
		{
			name:      "missing required email",
			json:      `{"password":"password123","name":"Test User"}`,
			model:     &models.SignUpPayload{},
			wantError: true,
			errorMsg:  "Email is a required field",
		},
		{
			name:      "password too short",
			json:      `{"email":"test@example.com","password":"123","name":"Test User"}`,
			model:     &models.SignUpPayload{},
			wantError: true,
			errorMsg:  "Password must be at least 8 characters in length",
		},
		{
			name:      "valid signin request",
			json:      `{"email":"test@example.com","password":"password123"}`,
			model:     &models.SignInPayload{},
			wantError: false,
		},
		{
			name:      "invalid JSON",
			json:      `{"email":"test@example.com","password":}`,
			model:     &models.SignInPayload{},
			wantError: true,
			errorMsg:  "invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.json))
			req.Header.Set("Content-Type", "application/json")

			err := transport.DecodeJSON(req, tt.model)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}

				// Check if it's a validation error
				var validationErr *ValidationError
				if errors.As(err, &validationErr) {
					if len(validationErr.Errors) == 0 {
						t.Errorf("expected validation errors but got none")
					}
					// Check if the expected error message is in one of the validation errors
					if tt.errorMsg != "" {
						found := false
						for _, vErr := range validationErr.Errors {
							if strings.Contains(vErr.Message, tt.errorMsg) {
								found = true
								break
							}
						}
						if !found {
							t.Errorf("expected validation error to contain %q, got validation errors: %+v", tt.errorMsg, validationErr.Errors)
						}
					}
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestDecodeJSON_ValidationErrorMessages(t *testing.T) {
	transport := NewDefault()

	// Test multiple validation errors
	jsonStr := `{"password":"123"}` // missing email (required), password too short (min=8)
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	var model models.SignUpPayload
	err := transport.DecodeJSON(req, &model)

	if err == nil {
		t.Fatal("expected validation error but got none")
	}

	// Check if it's a structured validation error
	var validationErr *ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got: %T", err)
	}

	if len(validationErr.Errors) != 2 {
		t.Errorf("expected 2 validation errors, got %d", len(validationErr.Errors))
	}

	// Check for specific validation errors
	errorFields := make(map[string]string)
	for _, vErr := range validationErr.Errors {
		errorFields[vErr.Field] = vErr.Message
	}

	if msg, exists := errorFields["Email"]; !exists || !strings.Contains(msg, "required field") {
		t.Errorf("expected Email required error, got: %s", msg)
	}

	if msg, exists := errorFields["Password"]; !exists || !strings.Contains(msg, "at least 8 characters in length") {
		t.Errorf("expected Password length error, got: %s", msg)
	}
}

func TestDecodeJSONWithLocale(t *testing.T) {
	transport := NewDefault()

	tests := []struct {
		name         string
		json         string
		locale       string
		expectedLang string // Expected language in error message
		model        any
	}{
		{
			name:         "English validation error",
			json:         `{"password":"123"}`,
			locale:       "en",
			expectedLang: "required field", // English translation
			model:        &models.SignUpPayload{},
		},
		{
			name:         "Spanish validation error",
			json:         `{"password":"123"}`,
			locale:       "es",
			expectedLang: "requerido", // Spanish translation
			model:        &models.SignUpPayload{},
		},
		{
			name:         "French validation error",
			json:         `{"password":"123"}`,
			locale:       "fr",
			expectedLang: "obligatoire", // French translation
			model:        &models.SignUpPayload{},
		},
		{
			name:         "Unsupported locale defaults to English",
			json:         `{"password":"123"}`,
			locale:       "de", // German not supported, should default to English
			expectedLang: "required field",
			model:        &models.SignUpPayload{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.json))
			req.Header.Set("Content-Type", "application/json")

			err := transport.DecodeJSONWithLocale(req, tt.model, tt.locale)

			if err == nil {
				t.Errorf("expected validation error but got none")
				return
			}

			var validationErr *ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("expected ValidationError, got: %T", err)
			}

			if len(validationErr.Errors) == 0 {
				t.Errorf("expected validation errors but got none")
				return
			}

			// Check if the error message is in the expected language
			found := false
			for _, vErr := range validationErr.Errors {
				if strings.Contains(strings.ToLower(vErr.Message), tt.expectedLang) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("expected error message to contain '%s', got messages: %v",
					tt.expectedLang, validationErr.Errors)
			}
		})
	}
}

func TestAcceptLanguageDetection(t *testing.T) {
	transport := NewDefault()

	tests := []struct {
		name           string
		acceptLanguage string
		expectedLocale string
	}{
		{
			name:           "English (US)",
			acceptLanguage: "en-US,en;q=0.9",
			expectedLocale: "en-US",
		},
		{
			name:           "Spanish (Spain)",
			acceptLanguage: "es-ES,es;q=0.9,en;q=0.8",
			expectedLocale: "es-ES",
		},
		{
			name:           "French (France)",
			acceptLanguage: "fr-FR,fr;q=0.9,en;q=0.8",
			expectedLocale: "fr-FR",
		},
		{
			name:           "Spanish variant supported directly",
			acceptLanguage: "es-AR,en;q=0.9", // Spanish (Argentina) -> supported directly
			expectedLocale: "es-AR",
		},
		{
			name:           "Unsupported language",
			acceptLanguage: "de-DE,de;q=0.9", // German not supported -> fallback to English
			expectedLocale: "en",
		},
		{
			name:           "No Accept-Language header",
			acceptLanguage: "",
			expectedLocale: "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(`{}`))
			if tt.acceptLanguage != "" {
				req.Header.Set("Accept-Language", tt.acceptLanguage)
			}

			locale := transport.(*Default).detectLocale(req)
			if locale != tt.expectedLocale {
				t.Errorf("expected locale %s, got %s", tt.expectedLocale, locale)
			}
		})
	}
}

func TestRespondValidationError(t *testing.T) {
	transport := NewDefault()

	// Create a validation error
	validationErr := &ValidationError{
		Errors: []models.ValidationError{
			{
				Field:   "Email",
				Message: "Email is a required field",
				Tag:     "required",
				Value:   "",
			},
			{
				Field:   "Password",
				Message: "Password must be at least 8 characters in length",
				Tag:     "min",
				Value:   "[hidden]",
			},
		},
	}

	// Create a response recorder
	w := httptest.NewRecorder()

	// Test the RespondValidationError method
	transport.RespondValidationError(w, validationErr)

	// Check the response
	if w.Code != 400 {
		t.Errorf("expected status code 400, got %d", w.Code)
	}

	var response models.ValidationErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Error != "Validation Failed" {
		t.Errorf("expected error 'Validation Failed', got '%s'", response.Error)
	}

	if response.Code != 400 {
		t.Errorf("expected code 400, got %d", response.Code)
	}

	if response.Type != "validation_error" {
		t.Errorf("expected type 'validation_error', got '%s'", response.Type)
	}

	if len(response.Errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(response.Errors))
	}

	// Check individual errors
	emailError := response.Errors[0]
	if emailError.Field != "Email" || emailError.Tag != "required" {
		t.Errorf("unexpected email error: %+v", emailError)
	}

	passwordError := response.Errors[1]
	if passwordError.Field != "Password" || passwordError.Tag != "min" || passwordError.Value != "[hidden]" {
		t.Errorf("unexpected password error: %+v", passwordError)
	}
}

