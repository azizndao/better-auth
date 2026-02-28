package transport

import (
	"github.com/azizndao/grouter"
	"github.com/go-playground/validator/v10"
)

// DecodeJSON decodes JSON request body with automatic language detection and validation
func (t *Default) DecodeJSON(c *grouter.Ctx, v any) error {
	locale := t.detectLocale(c)
	// Decode JSON first
	if err := c.BodyParser(v); err != nil {
		return grouter.ErrorBadRequest("Invalid request body", err)
	}

	// Validate the decoded struct
	if err := t.validator.Struct(v); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return t.buildValidationError(validationErrors, locale)
		}
		return grouter.ErrorBadRequest(err, err)
	}

	return nil
}
