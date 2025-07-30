package config

import "errors"

var (
	// ErrMissingSecretKey is returned when no secret key is provided
	ErrMissingSecretKey = errors.New("secret key is required")
	
	// ErrInvalidDatabaseURL is returned when database URL is invalid
	ErrInvalidDatabaseURL = errors.New("invalid database URL")
	
	// ErrInvalidBaseURL is returned when base URL is invalid
	ErrInvalidBaseURL = errors.New("invalid base URL")
)