package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// Handler utilities and common handler patterns

// JSONHandler wraps a handler function that returns JSON responses
func JSONHandler(handler func(http.ResponseWriter, *http.Request) (any, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := handler(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		if data != nil {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(data); err != nil {
				http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
			}
		}
	}
}

// StatusHandler creates a handler that returns a specific status code with optional message
func StatusHandler(code int, message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if message != "" {
			http.Error(w, message, code)
		} else {
			w.WriteHeader(code)
		}
	}
}

// RedirectHandler creates a handler that redirects to a specific URL
func RedirectHandler(url string, code int) http.HandlerFunc {
	if code == 0 {
		code = http.StatusFound
	}
	
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, url, code)
	}
}

// HealthCheckHandler creates a simple health check endpoint
func HealthCheckHandler(checks ...HealthCheck) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := HealthResult{
			Status: "ok",
			Checks: make(map[string]CheckResult),
		}
		
		overallHealthy := true
		
		for _, check := range checks {
			checkResult := CheckResult{Name: check.Name}
			
			if err := check.Check(); err != nil {
				checkResult.Status = "failed"
				checkResult.Error = err.Error()
				overallHealthy = false
			} else {
				checkResult.Status = "ok"
			}
			
			result.Checks[check.Name] = checkResult
		}
		
		if !overallHealthy {
			result.Status = "failed"
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

// HealthCheck defines a health check function
type HealthCheck struct {
	Name  string
	Check func() error
}

// HealthResult represents the response from a health check
type HealthResult struct {
	Status string                 `json:"status"`
	Checks map[string]CheckResult `json:"checks,omitempty"`
}

// CheckResult represents the result of an individual health check
type CheckResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// NotFoundHandler creates a custom 404 handler
func NotFoundHandler(message string) http.HandlerFunc {
	if message == "" {
		message = "Page not found"
	}
	
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		
		response := map[string]any{
			"error":   message,
			"code":    http.StatusNotFound,
			"path":    r.URL.Path,
			"method":  r.Method,
		}
		
		json.NewEncoder(w).Encode(response)
	}
}

// MethodNotAllowedHandler creates a custom 405 handler
func MethodNotAllowedHandler(allowedMethods []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(allowedMethods) > 0 {
			w.Header().Set("Allow", fmt.Sprintf("%v", allowedMethods))
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		
		response := map[string]any{
			"error":          "Method not allowed",
			"code":           http.StatusMethodNotAllowed,
			"method":         r.Method,
			"path":           r.URL.Path,
			"allowed_methods": allowedMethods,
		}
		
		json.NewEncoder(w).Encode(response)
	}
}

// Handler utility functions for common request parsing

// GetPathParam extracts a path parameter from the request
func GetPathParam(r *http.Request, key string) string {
	return r.PathValue(key) // Go 1.22+ feature
}

// GetQueryParam extracts a query parameter with optional default value
func GetQueryParam(r *http.Request, key, defaultValue string) string {
	if value := r.URL.Query().Get(key); value != "" {
		return value
	}
	return defaultValue
}

// GetQueryParamInt extracts an integer query parameter with optional default value
func GetQueryParamInt(r *http.Request, key string, defaultValue int) int {
	if value := r.URL.Query().Get(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// GetQueryParamBool extracts a boolean query parameter with optional default value
func GetQueryParamBool(r *http.Request, key string, defaultValue bool) bool {
	if value := r.URL.Query().Get(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

// GetHeader extracts a header value with optional default value
func GetHeader(r *http.Request, key, defaultValue string) string {
	if value := r.Header.Get(key); value != "" {
		return value
	}
	return defaultValue
}

// ParseJSON parses JSON from request body into the provided struct
func ParseJSON(r *http.Request, v any) error {
	if r.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Strict JSON parsing
	
	return decoder.Decode(v)
}

// RespondJSON sends a JSON response with the specified status code
func RespondJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// RespondError sends a JSON error response
func RespondError(w http.ResponseWriter, status int, message string) error {
	return RespondJSON(w, status, map[string]any{
		"error": message,
		"code":  status,
	})
}

// RespondSuccess sends a JSON success response
func RespondSuccess(w http.ResponseWriter, data any) error {
	return RespondJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    data,
	})
}

// RespondCreated sends a JSON response for created resources
func RespondCreated(w http.ResponseWriter, data any) error {
	return RespondJSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"data":    data,
	})
}

// RespondNoContent sends a 204 No Content response
func RespondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Wrapper types for typed handlers

// TypedHandler is a handler that works with typed request/response
type TypedHandler[Req, Res any] func(r *http.Request, req Req) (Res, error)

// WrapTypedHandler converts a typed handler to a standard http.HandlerFunc
func WrapTypedHandler[Req, Res any](handler TypedHandler[Req, Res]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Req
		
		// Parse JSON request if method has body
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			if err := ParseJSON(r, &req); err != nil {
				RespondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON: %v", err))
				return
			}
		}
		
		// Call the typed handler
		res, err := handler(r, req)
		if err != nil {
			RespondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		
		// Send response
		RespondSuccess(w, res)
	}
}

// WebSocketUpgrader provides a simple WebSocket upgrade utility
// Note: This would require the gorilla/websocket package in a real implementation
type WebSocketOptions struct {
	CheckOrigin     func(r *http.Request) bool
	Subprotocols    []string
	ReadBufferSize  int
	WriteBufferSize int
}

// FileUploadHandler creates a handler for file uploads
func FileUploadHandler(maxSize int64, allowedTypes []string, saveDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse multipart form
		err := r.ParseMultipartForm(maxSize)
		if err != nil {
			RespondError(w, http.StatusBadRequest, "Failed to parse multipart form")
			return
		}
		
		// Get file from form
		file, header, err := r.FormFile("file")
		if err != nil {
			RespondError(w, http.StatusBadRequest, "No file provided")
			return
		}
		defer file.Close()
		
		// Check file type if specified
		if len(allowedTypes) > 0 {
			contentType := header.Header.Get("Content-Type")
			allowed := false
			for _, t := range allowedTypes {
				if contentType == t {
					allowed = true
					break
				}
			}
			if !allowed {
				RespondError(w, http.StatusBadRequest, "File type not allowed")
				return
			}
		}
		
		// In a real implementation, you would save the file here
		
		RespondSuccess(w, map[string]any{
			"filename": header.Filename,
			"size":     header.Size,
			"type":     header.Header.Get("Content-Type"),
		})
	}
}