// Package transport provides HTTP transport utilities
package transport

import (
	"net/http"
)

// ExtractToken extracts authentication token from various sources
func (t *Default) ExtractToken(req *http.Request) string {
	// Try Authorization header first
	if token := t.extractTokenFromHeader(req); token != "" {
		return token
	}

	// Try session cookie
	if token := t.extractTokenFromCookie(req); token != "" {
		return token
	}

	// Try query parameter as fallback
	return req.URL.Query().Get("token")
}

// extractTokenFromHeader extracts token from Authorization header
func (t *Default) extractTokenFromHeader(req *http.Request) string {
	auth := req.Header.Get(HeaderAuthorization)
	if auth == "" {
		return ""
	}

	if len(auth) > BearerPrefixLen && auth[:BearerPrefixLen] == BearerPrefix {
		return auth[BearerPrefixLen:]
	}

	return ""
}

// extractTokenFromCookie extracts token from session cookie
func (t *Default) extractTokenFromCookie(req *http.Request) string {
	cookie, err := req.Cookie(SessionCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// SetCookie sets an HTTP cookie with secure defaults
func (t *Default) SetCookie(w http.ResponseWriter, name, value string, maxAge int) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

// GetCookie retrieves a cookie value by name
func (t *Default) GetCookie(req *http.Request, name string) (string, error) {
	cookie, err := req.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
