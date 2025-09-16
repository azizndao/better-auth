package router

import (
	"fmt"
	"net/http"
	"path"
	"strings"
)

// DefaultRouter implements the Router interface using Go's enhanced net/http features
type DefaultRouter struct {
	mux         *http.ServeMux
	options     RouterOptions
	middleware  []Middleware
	routes      []RouteInfo
	prefix      string
	groupMW     []Middleware
}

// NewRouter creates a new router with default options
func NewRouter() Router {
	return NewRouterWithOptions(DefaultRouterOptions())
}

// NewRouterWithOptions creates a new router with custom options
func NewRouterWithOptions(options RouterOptions) Router {
	r := &DefaultRouter{
		mux:     http.NewServeMux(),
		options: options,
		routes:  make([]RouteInfo, 0),
	}
	
	// Set up default handlers if needed
	r.setupDefaultHandlers()
	
	return r
}

// setupDefaultHandlers configures default handlers based on options
func (r *DefaultRouter) setupDefaultHandlers() {
	if r.options.NotFoundHandler == nil {
		r.options.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			http.Error(w, "Not Found", http.StatusNotFound)
		})
	}
	
	if r.options.MethodNotAllowedHandler == nil {
		r.options.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		})
	}
}

// GET registers a GET route
func (r *DefaultRouter) GET(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	r.Handle(http.MethodGet, pattern, handler, middleware...)
}

// POST registers a POST route
func (r *DefaultRouter) POST(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	r.Handle(http.MethodPost, pattern, handler, middleware...)
}

// PUT registers a PUT route
func (r *DefaultRouter) PUT(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	r.Handle(http.MethodPut, pattern, handler, middleware...)
}

// PATCH registers a PATCH route
func (r *DefaultRouter) PATCH(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	r.Handle(http.MethodPatch, pattern, handler, middleware...)
}

// DELETE registers a DELETE route
func (r *DefaultRouter) DELETE(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	r.Handle(http.MethodDelete, pattern, handler, middleware...)
}

// OPTIONS registers an OPTIONS route
func (r *DefaultRouter) OPTIONS(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	r.Handle(http.MethodOptions, pattern, handler, middleware...)
}

// HEAD registers a HEAD route
func (r *DefaultRouter) HEAD(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	r.Handle(http.MethodHead, pattern, handler, middleware...)
}

// Handle registers a route with a specific HTTP method
func (r *DefaultRouter) Handle(method, pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	// Build full pattern with prefix
	fullPattern := r.buildPattern(method, pattern)
	
	// Combine all middleware (global + group + route-specific)
	allMiddleware := make([]Middleware, 0, len(r.middleware)+len(r.groupMW)+len(middleware))
	allMiddleware = append(allMiddleware, r.middleware...)
	allMiddleware = append(allMiddleware, r.groupMW...)
	allMiddleware = append(allMiddleware, middleware...)
	
	// Wrap handler with middleware chain
	finalHandler := r.applyMiddleware(http.Handler(handler), allMiddleware)
	
	// Register with the mux
	r.mux.Handle(fullPattern, finalHandler)
	
	// Store route info for introspection
	r.routes = append(r.routes, RouteInfo{
		Method:     method,
		Pattern:    pattern,
		Handler:    handler,
		Middleware: allMiddleware,
		Group:      r.prefix,
	})
	
	// Auto-generate HEAD handler from GET if enabled
	if r.options.AutoHEAD && method == http.MethodGet {
		headPattern := r.buildPattern(http.MethodHead, pattern)
		r.mux.Handle(headPattern, finalHandler)
	}
}

// HandleFunc registers a route that matches any HTTP method
func (r *DefaultRouter) HandleFunc(pattern string, handler http.HandlerFunc, middleware ...Middleware) {
	fullPattern := r.buildPattern("", pattern)
	allMiddleware := make([]Middleware, 0, len(r.middleware)+len(r.groupMW)+len(middleware))
	allMiddleware = append(allMiddleware, r.middleware...)
	allMiddleware = append(allMiddleware, r.groupMW...)
	allMiddleware = append(allMiddleware, middleware...)
	
	finalHandler := r.applyMiddleware(http.Handler(handler), allMiddleware)
	r.mux.HandleFunc(fullPattern, finalHandler.ServeHTTP)
}

// Group creates a new route group with a prefix
func (r *DefaultRouter) Group(prefix string, middleware ...Middleware) RouteGroup {
	// Clean and combine prefixes
	fullPrefix := path.Join(r.prefix, prefix)
	if !strings.HasSuffix(fullPrefix, "/") && strings.HasSuffix(prefix, "/") {
		fullPrefix += "/"
	}
	
	// Combine middleware
	groupMW := make([]Middleware, 0, len(r.groupMW)+len(middleware))
	groupMW = append(groupMW, r.groupMW...)
	groupMW = append(groupMW, middleware...)
	
	return &DefaultRouter{
		mux:        r.mux,
		options:    r.options,
		middleware: r.middleware,
		routes:     r.routes,
		prefix:     fullPrefix,
		groupMW:    groupMW,
	}
}

// Use adds middleware to the router
func (r *DefaultRouter) Use(middleware ...Middleware) {
	r.middleware = append(r.middleware, middleware...)
}

// Static serves static files from a directory
func (r *DefaultRouter) Static(pattern, dir string) {
	// Ensure pattern ends with /* for proper file serving
	if !strings.HasSuffix(pattern, "/") {
		pattern += "/"
	}
	pattern += "{file...}"
	
	fullPattern := r.buildPattern(http.MethodGet, pattern)
	
	fileServer := http.FileServer(http.Dir(dir))
	handler := http.StripPrefix(strings.TrimSuffix(r.prefix+pattern, "{file...}"), fileServer)
	
	r.mux.Handle(fullPattern, handler)
}

// StaticFile serves a single static file
func (r *DefaultRouter) StaticFile(pattern, file string) {
	fullPattern := r.buildPattern(http.MethodGet, pattern)
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, file)
	})
	
	r.mux.Handle(fullPattern, handler)
}

// ServeHTTP implements http.Handler
func (r *DefaultRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// Handler returns the underlying http.Handler
func (r *DefaultRouter) Handler() http.Handler {
	return r
}

// Routes returns information about all registered routes
func (r *DefaultRouter) Routes() []RouteInfo {
	return r.routes
}

// buildPattern constructs the full pattern for registration
func (r *DefaultRouter) buildPattern(method, pattern string) string {
	// Clean the pattern
	if pattern == "" {
		pattern = "/"
	}
	
	// Combine prefix and pattern
	fullPath := path.Join(r.prefix, pattern)
	
	// Preserve trailing slash if original pattern had it
	if strings.HasSuffix(pattern, "/") && !strings.HasSuffix(fullPath, "/") && fullPath != "/" {
		fullPath += "/"
	}
	
	// Add method prefix for Go 1.22+ enhanced routing
	if method != "" {
		return fmt.Sprintf("%s %s", method, fullPath)
	}
	
	return fullPath
}

// applyMiddleware applies a chain of middleware to a handler
func (r *DefaultRouter) applyMiddleware(handler http.Handler, middleware []Middleware) http.Handler {
	// Apply middleware in reverse order so they execute in the correct order
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	
	// Apply built-in middleware based on options
	if r.options.EnableRecovery {
		handler = recoveryMiddleware(handler)
	}
	
	if r.options.EnableLogging {
		handler = loggingMiddleware(handler)
	}
	
	return handler
}

// Built-in middleware implementations

// recoveryMiddleware handles panics and returns a 500 error
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs incoming requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s %s\n", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}