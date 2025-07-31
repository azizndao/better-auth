package transport

import (
	"fmt"
	"net/http"

	"better-auth/internal/models"
	"github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// Constants for transport configuration
const (
	// Default locale when no language is specified or supported
	DefaultLocale = "en"
	
	// Context key for user data
	UserContextKey = "betterauth.user"
	
	// Cookie name for session tokens
	SessionCookieName = "better-auth.session_token"
	
	// HTTP headers
	HeaderAuthorization  = "Authorization"
	HeaderAcceptLanguage = "Accept-Language"
	HeaderContentType    = "Content-Type"
	
	// Content types
	ContentTypeJSON = "application/json"
	
	// Token prefix
	BearerPrefix = "Bearer "
	BearerPrefixLen = 7
)

// ValidationError represents a custom error type for validation failures
type ValidationError struct {
	Errors []models.ValidationError `json:"errors"`
}

// Error implements the error interface
func (ve *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: %d errors", len(ve.Errors))
}

// Transport interface defines how requests and responses are handled
type Transport interface {
	// JSON operations
	DecodeJSON(req *http.Request, v any) error
	DecodeJSONWithLocale(req *http.Request, v any, locale string) error
	RespondJSON(w http.ResponseWriter, status int, data any)
	
	// Error responses
	RespondError(w http.ResponseWriter, status int, message string)
	RespondValidationError(w http.ResponseWriter, validationErr *ValidationError)
	WriteError(w http.ResponseWriter, status int, message string)
	
	// User context operations
	GetUserContext(req *http.Request) (*models.UserContext, bool)
	SetUserContext(req *http.Request, ctx *models.UserContext) *http.Request
	
	// Token operations
	ExtractToken(req *http.Request) string
	
	// Cookie operations
	SetCookie(w http.ResponseWriter, name, value string, maxAge int)
	GetCookie(req *http.Request, name string) (string, error)
}

// Default implementation of the Transport interface
type Default struct {
	validator   *validator.Validate
	translators map[string]ut.Translator
	uni         *ut.UniversalTranslator
}

type contextKey string