package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"better-auth/internal/models"
)

// Transport interface defines how requests and responses are handled
type Transport interface {
	DecodeJSON(req *http.Request, v interface{}) error
	RespondJSON(w http.ResponseWriter, status int, data interface{})
	RespondError(w http.ResponseWriter, status int, message string)
	WriteError(w http.ResponseWriter, status int, message string)
	GetUserContext(req *http.Request) (*models.UserContext, bool)
	SetUserContext(req *http.Request, ctx *models.UserContext) *http.Request
	ExtractToken(req *http.Request) string
	SetCookie(w http.ResponseWriter, name, value string, maxAge int)
	GetCookie(req *http.Request, name string) (string, error)
}

// Default implementation
type Default struct{}

// NewDefault creates a new default transport
func NewDefault() Transport {
	return &Default{}
}

func (t *Default) DecodeJSON(req *http.Request, v interface{}) error {
	return json.NewDecoder(req.Body).Decode(v)
}

func (t *Default) RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (t *Default) RespondError(w http.ResponseWriter, status int, message string) {
	t.RespondJSON(w, status, map[string]string{"error": message})
}

func (t *Default) WriteError(w http.ResponseWriter, status int, message string) {
	t.RespondError(w, status, message)
}

func (t *Default) GetUserContext(req *http.Request) (*models.UserContext, bool) {
	ctx := req.Context().Value(contextKey("betterauth.user"))
	if ctx == nil {
		return nil, false
	}
	userCtx, ok := ctx.(*models.UserContext)
	return userCtx, ok
}

func (t *Default) SetUserContext(req *http.Request, ctx *models.UserContext) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), contextKey("betterauth.user"), ctx))
}

func (t *Default) ExtractToken(req *http.Request) string {
	if auth := req.Header.Get("Authorization"); auth != "" {
		if len(auth) > 7 && auth[:7] == "Bearer " {
			return auth[7:]
		}
	}

	if cookie, err := req.Cookie("better-auth.session_token"); err == nil {
		return cookie.Value
	}

	return req.URL.Query().Get("token")
}

func (t *Default) SetCookie(w http.ResponseWriter, name, value string, maxAge int) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

func (t *Default) GetCookie(req *http.Request, name string) (string, error) {
	cookie, err := req.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

type contextKey string