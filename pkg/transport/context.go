package transport

import (
	"context"
	"net/http"

	"better-auth/internal/models"
)

// GetUserContext retrieves user context from the request
func (t *Default) GetUserContext(req *http.Request) (*models.AuthData, bool) {
	ctx := req.Context().Value(contextKey(UserContextKey))
	if ctx == nil {
		return nil, false
	}
	userCtx, ok := ctx.(*models.AuthData)
	return userCtx, ok
}

// SetUserContext sets user context in the request
func (t *Default) SetUserContext(req *http.Request, ctx *models.AuthData) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), contextKey(UserContextKey), ctx))
}

