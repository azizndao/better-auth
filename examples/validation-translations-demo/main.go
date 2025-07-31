package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"

	"better-auth/internal/models"
	"better-auth/pkg/transport"
)

func main() {
	// Create a transport instance with multi-language validation
	t := transport.NewDefault()

	fmt.Println("🌍 Better Auth - Multi-Language Validation Demo")
	fmt.Println("================================================")

	// Example 1: English validation (default)
	fmt.Println("\n🇺🇸 Example 1: English Validation")
	fmt.Println("----------------------------------")
	demonstrateValidation(t, "en", `{
		"email": "not-an-email",
		"password": "123"
	}`)

	// Example 2: Spanish validation
	fmt.Println("\n🇪🇸 Example 2: Spanish Validation")
	fmt.Println("----------------------------------")
	demonstrateValidation(t, "es", `{
		"email": "no-es-email",
		"password": "123"
	}`)

	// Example 3: French validation
	fmt.Println("\n🇫🇷 Example 3: French Validation")
	fmt.Println("---------------------------------")
	demonstrateValidation(t, "fr", `{
		"email": "pas-un-email",
		"password": "123"
	}`)

	// Example 4: Auto-detection from Accept-Language header
	fmt.Println("\n🌐 Example 4: Auto-Detection from Accept-Language Header")
	fmt.Println("-------------------------------------------------------")
	demonstrateAutoDetection(t, "es-ES,es;q=0.9,en;q=0.8", `{
		"password": "123"
	}`)

	// Example 5: Fallback for unsupported language
	fmt.Println("\n🌐 Example 5: Fallback for Unsupported Language")
	fmt.Println("------------------------------------------------")
	demonstrateAutoDetection(t, "de-DE,de;q=0.9", `{
		"email": "invalid-email"
	}`)

	// Example 6: Multiple validation errors in different languages
	fmt.Println("\n📋 Example 6: Multiple Validation Errors")
	fmt.Println("----------------------------------------")
	demonstrateMultipleErrors(t)

	fmt.Println("\n🎉 Demo Complete!")
	fmt.Println("=================")
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("• Professional validation messages using go-playground/validator translations")
	fmt.Println("• Multi-language support (English, Spanish, French)")
	fmt.Println("• Automatic language detection from Accept-Language header")
	fmt.Println("• Structured error responses with field-level details")
	fmt.Println("• Security (password values are hidden)")
	fmt.Println("• Fallback to English for unsupported languages")
}

func demonstrateValidation(t transport.Transport, locale, jsonStr string) {
	fmt.Printf("Locale: %s\n", locale)
	fmt.Printf("Request Body: %s\n", jsonStr)

	req := httptest.NewRequest("POST", "/signup", bytes.NewBufferString(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	var model models.SignUpRequest
	err := t.DecodeJSONWithLocale(req, &model, locale)

	if err == nil {
		fmt.Printf("✅ Success: Request validated successfully\n")
		return
	}

	var validationErr *transport.ValidationError
	if errors.As(err, &validationErr) {
		fmt.Printf("❌ Validation Failed:\n")
		for i, fieldErr := range validationErr.Errors {
			fmt.Printf("  %d. %s: %s\n", i+1, fieldErr.Field, fieldErr.Message)
		}
	} else {
		fmt.Printf("❌ Error: %v\n", err)
	}
}

func demonstrateAutoDetection(t transport.Transport, acceptLang, jsonStr string) {
	fmt.Printf("Accept-Language: %s\n", acceptLang)
	fmt.Printf("Request Body: %s\n", jsonStr)

	req := httptest.NewRequest("POST", "/signup", bytes.NewBufferString(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", acceptLang)

	var model models.SignUpRequest
	err := t.DecodeJSON(req, &model) // Auto-detects language from Accept-Language header

	if err == nil {
		fmt.Printf("✅ Success: Request validated successfully\n")
		return
	}

	var validationErr *transport.ValidationError
	if errors.As(err, &validationErr) {
		fmt.Printf("❌ Validation Failed (Auto-detected language):\n")
		for i, fieldErr := range validationErr.Errors {
			fmt.Printf("  %d. %s: %s\n", i+1, fieldErr.Field, fieldErr.Message)
		}
	} else {
		fmt.Printf("❌ Error: %v\n", err)
	}
}

func demonstrateMultipleErrors(t transport.Transport) {
	languages := []struct {
		locale string
		flag   string
		name   string
	}{
		{"en", "🇺🇸", "English"},
		{"es", "🇪🇸", "Spanish"},
		{"fr", "🇫🇷", "French"},
	}

	// Invalid request with multiple validation errors
	invalidJSON := `{}`

	for _, lang := range languages {
		fmt.Printf("\n%s %s:\n", lang.flag, lang.name)
		
		req := httptest.NewRequest("POST", "/signup", bytes.NewBufferString(invalidJSON))
		req.Header.Set("Content-Type", "application/json")

		var model models.SignUpRequest
		err := t.DecodeJSONWithLocale(req, &model, lang.locale)

		if err != nil {
			var validationErr *transport.ValidationError
			if errors.As(err, &validationErr) {
				// Simulate HTTP response
				w := httptest.NewRecorder()
				t.RespondValidationError(w, validationErr)

				var response models.ValidationErrorResponse
				if jsonErr := json.Unmarshal(w.Body.Bytes(), &response); jsonErr == nil {
					fmt.Printf("  HTTP %d - %s\n", response.Code, response.Message)
					for _, fieldErr := range response.Errors {
						fmt.Printf("  • %s: %s\n", fieldErr.Field, fieldErr.Message)
					}
				}
			}
		}
	}
}