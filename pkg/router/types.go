package router

import (
	"context"
	"net/http"
)

// Router interface defines the main routing functionality
type Router interface {
	// HTTP method routing with modern net/http patterns
	GET(pattern string, handler http.HandlerFunc, middleware ...Middleware)
	POST(pattern string, handler http.HandlerFunc, middleware ...Middleware)
	PUT(pattern string, handler http.HandlerFunc, middleware ...Middleware)
	PATCH(pattern string, handler http.HandlerFunc, middleware ...Middleware)
	DELETE(pattern string, handler http.HandlerFunc, middleware ...Middleware)
	OPTIONS(pattern string, handler http.HandlerFunc, middleware ...Middleware)
	HEAD(pattern string, handler http.HandlerFunc, middleware ...Middleware)
	
	// Advanced routing methods
	Handle(method, pattern string, handler http.HandlerFunc, middleware ...Middleware)
	HandleFunc(pattern string, handler http.HandlerFunc, middleware ...Middleware)
	
	// Route groups for organizing related routes
	Group(prefix string, middleware ...Middleware) RouteGroup
	
	// Middleware management
	Use(middleware ...Middleware)
	
	// Static file serving with modern features
	Static(pattern, dir string)
	StaticFile(pattern, file string)
	
	// Server integration
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Handler() http.Handler
}

// RouteGroup interface for grouped routes
type RouteGroup interface {
	// HTTP method routing within the group
	GET(pattern string, handler http.HandlerFunc, middleware ...Middleware)
	POST(pattern string, handler http.HandlerFunc, middleware ...Middleware)
	PUT(pattern string, handler http.HandlerFunc, middleware ...Middleware)
	PATCH(pattern string, handler http.HandlerFunc, middleware ...Middleware)
	DELETE(pattern string, handler http.HandlerFunc, middleware ...Middleware)
	OPTIONS(pattern string, handler http.HandlerFunc, middleware ...Middleware)
	HEAD(pattern string, handler http.HandlerFunc, middleware ...Middleware)
	
	// Advanced routing within the group
	Handle(method, pattern string, handler http.HandlerFunc, middleware ...Middleware)
	HandleFunc(pattern string, handler http.HandlerFunc, middleware ...Middleware)
	
	// Nested groups
	Group(prefix string, middleware ...Middleware) RouteGroup
	
	// Group middleware
	Use(middleware ...Middleware)
}

// Middleware defines the middleware function signature
type Middleware func(http.Handler) http.Handler

// HandlerContext provides additional context for handlers
type HandlerContext struct {
	// Route parameters extracted from URL path
	Params map[string]string
	
	// Original request and response writer
	Request  *http.Request
	Response http.ResponseWriter
	
	// User-defined values
	Values map[string]any
}

// ContextKey is used for context values
type ContextKey string

const (
	// ParamsContextKey is the key for storing route parameters in request context
	ParamsContextKey ContextKey = "router.params"
	
	// ValuesContextKey is the key for storing custom values in request context
	ValuesContextKey ContextKey = "router.values"
)

// RouteInfo contains information about a registered route
type RouteInfo struct {
	Method      string
	Pattern     string
	Handler     http.HandlerFunc
	Middleware  []Middleware
	Group       string
	Description string
}

// RouterOptions contains configuration options for the router
type RouterOptions struct {
	// Enable automatic OPTIONS responses for CORS
	AutoOPTIONS bool
	
	// Enable automatic HEAD responses from GET handlers
	AutoHEAD bool
	
	// Enable trailing slash redirect
	TrailingSlashRedirect bool
	
	// Enable method not allowed responses
	MethodNotAllowed bool
	
	// Custom not found handler
	NotFoundHandler http.Handler
	
	// Custom method not allowed handler
	MethodNotAllowedHandler http.Handler
	
	// Enable request logging
	EnableLogging bool
	
	// Enable panic recovery
	EnableRecovery bool
}

// DefaultRouterOptions returns sensible default options
func DefaultRouterOptions() RouterOptions {
	return RouterOptions{
		AutoOPTIONS:           true,
		AutoHEAD:             true,
		TrailingSlashRedirect: true,
		MethodNotAllowed:      true,
		EnableLogging:         false,
		EnableRecovery:        true,
	}
}

// PathParams extracts path parameters from request context
func PathParams(r *http.Request) map[string]string {
	if params, ok := r.Context().Value(ParamsContextKey).(map[string]string); ok {
		return params
	}
	return make(map[string]string)
}

// PathParam extracts a single path parameter from request context
func PathParam(r *http.Request, key string) string {
	params := PathParams(r)
	return params[key]
}

// SetPathParam sets a path parameter in the request context
func SetPathParam(r *http.Request, key, value string) *http.Request {
	params := PathParams(r)
	if params == nil {
		params = make(map[string]string)
	}
	params[key] = value
	
	ctx := context.WithValue(r.Context(), ParamsContextKey, params)
	return r.WithContext(ctx)
}

// Values gets custom values from request context
func Values(r *http.Request) map[string]any {
	if values, ok := r.Context().Value(ValuesContextKey).(map[string]any); ok {
		return values
	}
	return make(map[string]any)
}

// Value gets a single custom value from request context
func Value(r *http.Request, key string) any {
	values := Values(r)
	return values[key]
}

// SetValue sets a custom value in the request context
func SetValue(r *http.Request, key string, value any) *http.Request {
	values := Values(r)
	if values == nil {
		values = make(map[string]any)
	}
	values[key] = value
	
	ctx := context.WithValue(r.Context(), ValuesContextKey, values)
	return r.WithContext(ctx)
}