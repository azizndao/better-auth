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
	// Create a transport instance with validation
	t := transport.NewDefault()

	fmt.Println("🚀 Better Auth - Enhanced Validation Response Demo")
	fmt.Println("=======================================================")

	// Example 1: Valid request
	fmt.Println("\n✅ Example 1: Valid Sign-up Request")
	fmt.Println("-----------------------------------")
	validJSON := `{
		"email": "user@example.com",
		"password": "securepassword123",
		"name": "John Doe"
	}`
	
	demonstrateValidation(t, "POST", "/signup", validJSON, &models.SignUpRequest{})

	// Example 2: Multiple validation errors
	fmt.Println("\n❌ Example 2: Multiple Validation Errors")
	fmt.Println("----------------------------------------")
	invalidJSON := `{
		"email": "not-an-email",
		"password": "123"
	}`
	
	demonstrateValidation(t, "POST", "/signup", invalidJSON, &models.SignUpRequest{})

	// Example 3: Missing required fields
	fmt.Println("\n❌ Example 3: Missing Required Fields")
	fmt.Println("------------------------------------")
	missingJSON := `{
		"name": "John Doe"
	}`
	
	demonstrateValidation(t, "POST", "/signup", missingJSON, &models.SignUpRequest{})

	// Example 4: 2FA Code validation
	fmt.Println("\n❌ Example 4: Invalid 2FA Code Length")
	fmt.Println("------------------------------------")
	invalidCodeJSON := `{
		"code": "12345"
	}`
	
	demonstrateValidation(t, "POST", "/2fa/verify", invalidCodeJSON, &models.TwoFactorVerifyRequest{})

	// Example 5: JSON syntax error
	fmt.Println("\n❌ Example 5: Invalid JSON Syntax")
	fmt.Println("----------------------------------")
	malformedJSON := `{
		"email": "user@example.com",
		"password": "validpassword",
		"name": "John Doe",
	}` // Extra comma causes JSON syntax error
	
	demonstrateValidation(t, "POST", "/signup", malformedJSON, &models.SignUpRequest{})

	fmt.Println("\n🎉 Demo Complete!")
	fmt.Println("=================")
	fmt.Println("The transport layer now provides:")
	fmt.Println("• Structured validation error responses")
	fmt.Println("• Field-by-field error details")
	fmt.Println("• Security (password values are hidden)")
	fmt.Println("• Clear error messages for developers")
	fmt.Println("• Proper HTTP status codes")
}

func demonstrateValidation(t transport.Transport, method, path, jsonStr string, model any) {
	fmt.Printf("Request: %s %s\n", method, path)
	fmt.Printf("Body: %s\n", jsonStr)

	req := httptest.NewRequest(method, path, bytes.NewBufferString(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	err := t.DecodeJSON(req, model)

	if err == nil {
		fmt.Printf("✅ Success: Request validated successfully\n")
		fmt.Printf("Parsed data: %+v\n", model)
		return
	}

	// Check if it's a validation error
	var validationErr *transport.ValidationError
	if errors.As(err, &validationErr) {
		// Simulate the HTTP response
		w := httptest.NewRecorder()
		t.RespondValidationError(w, validationErr)

		fmt.Printf("❌ Validation Failed (HTTP %d)\n", w.Code)
		
		// Parse and display the structured response
		var response models.ValidationErrorResponse
		if jsonErr := json.Unmarshal(w.Body.Bytes(), &response); jsonErr == nil {
			fmt.Printf("Response Type: %s\n", response.Type)
			fmt.Printf("Message: %s\n", response.Message)
			fmt.Printf("Field Errors:\n")
			for i, fieldErr := range response.Errors {
				fmt.Printf("  %d. Field: %s\n", i+1, fieldErr.Field)
				fmt.Printf("     Tag: %s\n", fieldErr.Tag)
				fmt.Printf("     Message: %s\n", fieldErr.Message)
				fmt.Printf("     Value: %v\n", fieldErr.Value)
			}
		} else {
			fmt.Printf("Raw Response: %s\n", w.Body.String())
		}
	} else {
		// Handle non-validation errors (like JSON syntax errors)
		fmt.Printf("❌ Error: %v\n", err)
		fmt.Printf("This would result in a generic 400 Bad Request response\n")
	}
}