package transport

import (
	"fmt"
	"net/http"

	"better-auth/internal/models"

	ut "github.com/go-playground/universal-translator"
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
	BearerPrefix    = "Bearer "
	BearerPrefixLen = 7
)

// Transport interface defines how requests and responses are handled
type Transport interface {
	// JSON operations
	DecodeJSON(req *http.Request, v any) error
	RespondJSON(w http.ResponseWriter, status int, data any)

	// Error responses
	RespondError(w http.ResponseWriter, data error)

	// User context operations
	GetUserContext(req *http.Request) (*models.AuthData, bool)
	SetUserContext(req *http.Request, ctx *models.AuthData) *http.Request

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

type APIError struct {
	internal error `json:"-"`
	Code     int   `json:"code"`
	Data     any   `json:"data"`
}

func NewAPIError(code int, data any, internal error) *APIError {
	return &APIError{
		Code:     code,
		Data:     data,
		internal: internal,
	}
}

func (e *APIError) Error() string {
	if e.internal != nil {
		return e.internal.Error()
	}

	return fmt.Sprintf("API error: %d", e.Code)
}

func (e *APIError) Unwrap() error {
	return e.internal
}

func NewBadRequestError(data any, err error) *APIError {
	return &APIError{
		Code:     http.StatusBadRequest,
		Data:     data,
		internal: err,
	}
}

func NewUnauthorizedError(data any, err error) *APIError {
	return &APIError{
		Code:     http.StatusUnauthorized,
		Data:     data,
		internal: err,
	}
}

func NewForbiddenError(data any, err error) *APIError {
	return &APIError{
		Code:     http.StatusForbidden,
		Data:     data,
		internal: err,
	}
}

func NewNotFoundError(data any, err error) *APIError {
	return &APIError{
		Code:     http.StatusNotFound,
		Data:     data,
		internal: err,
	}
}

func NewInternalServerError(data any, err error) *APIError {
	return &APIError{
		Code:     http.StatusInternalServerError,
		Data:     data,
		internal: err,
	}
}

func NewConflictError(data any, err error) *APIError {
	return &APIError{
		Code:     http.StatusConflict,
		Data:     data,
		internal: err,
	}
}

func NewValidationError(data any, err error) *APIError {
	return &APIError{
		Code:     http.StatusUnprocessableEntity,
		Data:     data,
		internal: err,
	}
}
