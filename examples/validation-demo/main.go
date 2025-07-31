package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http/httptest"

	"better-auth/internal/models"
	"better-auth/pkg/transport"
)

func main() {
	// Create a transport instance with validation
	t := transport.NewDefault()

	// Example 1: Valid request
	fmt.Println("=== Example 1: Valid Sign-up Request ===")
	validJSON := `{
		"email": "user@example.com",
		"password": "securepassword123",
		"name": "John Doe"
	}`
	
	req := httptest.NewRequest("POST", "/signup", bytes.NewBufferString(validJSON))
	var validSignup models.SignUpRequest
	
	if err := t.DecodeJSON(req, &validSignup); err != nil {
		log.Printf("Validation error: %v", err)
	} else {
		fmt.Printf("✅ Request validated successfully: %+v\n", validSignup)
	}

	// Example 2: Invalid email
	fmt.Println("\n=== Example 2: Invalid Email Address ===")
	invalidEmailJSON := `{
		"email": "not-an-email",
		"password": "securepassword123",
		"name": "John Doe"
	}`
	
	req = httptest.NewRequest("POST", "/signup", bytes.NewBufferString(invalidEmailJSON))
	var invalidEmailSignup models.SignUpRequest
	
	if err := t.DecodeJSON(req, &invalidEmailSignup); err != nil {
		fmt.Printf("❌ Validation error: %v\n", err)
	} else {
		fmt.Printf("✅ Request validated successfully: %+v\n", invalidEmailSignup)
	}

	// Example 3: Missing required fields
	fmt.Println("\n=== Example 3: Missing Required Fields ===")
	missingFieldsJSON := `{
		"name": "John Doe"
	}`
	
	req = httptest.NewRequest("POST", "/signup", bytes.NewBufferString(missingFieldsJSON))
	var missingFieldsSignup models.SignUpRequest
	
	if err := t.DecodeJSON(req, &missingFieldsSignup); err != nil {
		fmt.Printf("❌ Validation error: %v\n", err)
	} else {
		fmt.Printf("✅ Request validated successfully: %+v\n", missingFieldsSignup)
	}

	// Example 4: Password too short
	fmt.Println("\n=== Example 4: Password Too Short ===")
	shortPasswordJSON := `{
		"email": "user@example.com",
		"password": "123",
		"name": "John Doe"
	}`
	
	req = httptest.NewRequest("POST", "/signup", bytes.NewBufferString(shortPasswordJSON))
	var shortPasswordSignup models.SignUpRequest
	
	if err := t.DecodeJSON(req, &shortPasswordSignup); err != nil {
		fmt.Printf("❌ Validation error: %v\n", err)
	} else {
		fmt.Printf("✅ Request validated successfully: %+v\n", shortPasswordSignup)
	}

	// Example 5: 2FA code validation
	fmt.Println("\n=== Example 5: 2FA Code Validation ===")
	invalidCodeJSON := `{
		"code": "12345"
	}`
	
	req = httptest.NewRequest("POST", "/2fa/verify", bytes.NewBufferString(invalidCodeJSON))
	var twoFARequest models.TwoFactorVerifyRequest
	
	if err := t.DecodeJSON(req, &twoFARequest); err != nil {
		fmt.Printf("❌ Validation error: %v\n", err)
	} else {
		fmt.Printf("✅ Request validated successfully: %+v\n", twoFARequest)
	}

	// Example 6: Valid 2FA code
	fmt.Println("\n=== Example 6: Valid 2FA Code ===")
	validCodeJSON := `{
		"code": "123456"
	}`
	
	req = httptest.NewRequest("POST", "/2fa/verify", bytes.NewBufferString(validCodeJSON))
	var validTwoFARequest models.TwoFactorVerifyRequest
	
	if err := t.DecodeJSON(req, &validTwoFARequest); err != nil {
		fmt.Printf("❌ Validation error: %v\n", err)
	} else {
		fmt.Printf("✅ Request validated successfully: %+v\n", validTwoFARequest)
	}

	fmt.Println("\n=== Validation Demo Complete ===")
	fmt.Println("The transport layer now automatically validates all requests using the 'validate' tags in your models!")
}