package transport

import (
	"github.com/go-playground/validator/v10"
)

// NewDefault creates a new default transport with multi-language validation support
func NewDefault() Transport {
	validate := validator.New()
	translators := initializeTranslators(validate)
	
	return &Default{
		validator:   validate,
		translators: translators,
		uni:         nil, // Not needed after initialization
	}
}