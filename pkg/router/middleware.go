package router

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// Common middleware implementations for the router

// CORS middleware for handling Cross-Origin Resource Sharing
func CORS(options CORSOptions) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			
			// Set CORS headers
			if len(options.AllowedOrigins) > 0 {
				for _, allowedOrigin := range options.AllowedOrigins {
					if allowedOrigin == "*" || allowedOrigin == origin {
						w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
						break
					}
				}
			}
			
			if len(options.AllowedMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(options.AllowedMethods, ", "))
			}
			
			if len(options.AllowedHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(options.AllowedHeaders, ", "))
			}
			
			if options.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			
			if options.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", int(options.MaxAge.Seconds())))
			}
			
			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// CORSOptions contains configuration for CORS middleware
type CORSOptions struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// DefaultCORSOptions returns sensible default CORS options
func DefaultCORSOptions() CORSOptions {
	return CORSOptions{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type", "Accept", "Origin", "User-Agent", "DNT", "Cache-Control", "X-Mx-ReqToken", "Keep-Alive", "X-Requested-With", "If-Modified-Since"},
		MaxAge:         24 * time.Hour,
	}
}

// Logger middleware for request logging with custom format
func Logger(format LogFormat) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Create a response writer wrapper to capture status code
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			
			next.ServeHTTP(rw, r)
			
			duration := time.Since(start)
			
			switch format {
			case LogFormatCombined:
				fmt.Printf("%s - - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\" %v\n",
					r.RemoteAddr,
					start.Format("02/Jan/2006:15:04:05 -0700"),
					r.Method,
					r.URL.Path,
					r.Proto,
					rw.statusCode,
					rw.size,
					r.Referer(),
					r.UserAgent(),
					duration,
				)
			case LogFormatCommon:
				fmt.Printf("%s - - [%s] \"%s %s %s\" %d %d\n",
					r.RemoteAddr,
					start.Format("02/Jan/2006:15:04:05 -0700"),
					r.Method,
					r.URL.Path,
					r.Proto,
					rw.statusCode,
					rw.size,
				)
			case LogFormatShort:
				fmt.Printf("%s %s %d %v\n", r.Method, r.URL.Path, rw.statusCode, duration)
			}
		})
	}
}

// LogFormat defines the format for request logging
type LogFormat int

const (
	LogFormatShort LogFormat = iota
	LogFormatCommon
	LogFormatCombined
)

// responseWriter wraps http.ResponseWriter to capture status code and size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// Gzip middleware for response compression
func Gzip(level int) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if client accepts gzip
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}
			
			// Create gzip writer
			gz, err := gzip.NewWriterLevel(w, level)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			defer gz.Close()
			
			// Set response headers
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Vary", "Accept-Encoding")
			
			// Wrap response writer
			gzw := &gzipResponseWriter{ResponseWriter: w, gzipWriter: gz}
			next.ServeHTTP(gzw, r)
		})
	}
}

// gzipResponseWriter wraps http.ResponseWriter with gzip compression
type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter io.Writer
}

func (grw *gzipResponseWriter) Write(b []byte) (int, error) {
	return grw.gzipWriter.Write(b)
}

// Timeout middleware for request timeout handling
func Timeout(timeout time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, timeout, "Request Timeout")
	}
}

// RequestID middleware adds a unique request ID to each request
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate simple request ID (in production, use a more sophisticated approach)
			requestID := fmt.Sprintf("%d", time.Now().UnixNano())
			
			// Add to response header
			w.Header().Set("X-Request-Id", requestID)
			
			// Add to request context
			ctx := context.WithValue(r.Context(), "request-id", requestID)
			r = r.WithContext(ctx)
			
			next.ServeHTTP(w, r)
		})
	}
}

// Recovery middleware with better error handling and optional callback
func Recovery(callback func(err any, stack []byte)) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					stack := debug.Stack()
					
					// Call callback if provided
					if callback != nil {
						callback(err, stack)
					}
					
					// Log the error
					fmt.Printf("PANIC: %v\n%s\n", err, stack)
					
					// Return 500 error
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			
			next.ServeHTTP(w, r)
		})
	}
}

// BasicAuth middleware for HTTP Basic Authentication
func BasicAuth(realm string, validator func(username, password string) bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			
			if !ok || !validator(username, password) {
				w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimiter middleware for basic rate limiting (simple in-memory implementation)
func RateLimiter(requests int, window time.Duration) Middleware {
	clients := make(map[string]*rateLimitInfo)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)
			now := time.Now()
			
			// Clean up old entries
			for ip, info := range clients {
				if now.Sub(info.windowStart) > window {
					delete(clients, ip)
				}
			}
			
			// Get or create client info
			info, exists := clients[clientIP]
			if !exists {
				info = &rateLimitInfo{
					requests:    0,
					windowStart: now,
				}
				clients[clientIP] = info
			}
			
			// Reset window if needed
			if now.Sub(info.windowStart) > window {
				info.requests = 0
				info.windowStart = now
			}
			
			// Check rate limit
			if info.requests >= requests {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			
			info.requests++
			next.ServeHTTP(w, r)
		})
	}
}

type rateLimitInfo struct {
	requests    int
	windowStart time.Time
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP if multiple are present
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	
	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// Chain combines multiple middleware into a single middleware
func Chain(middleware ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middleware) - 1; i >= 0; i-- {
			next = middleware[i](next)
		}
		return next
	}
}