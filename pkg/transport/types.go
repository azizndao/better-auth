package transport

import (
	"net/http"

	"github.com/azizndao/grouter"
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
	DecodeJSON(req *grouter.Ctx, v any) error

	// Token operations
	ExtractToken(req *http.Request) string
}

// Default implementation of the Transport interface
type Default struct {
	validator   *validator.Validate
	translators map[string]ut.Translator
	uni         *ut.UniversalTranslator
}
