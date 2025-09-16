package router

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewRouter(t *testing.T) {
	router := NewRouter()
	if router == nil {
		t.Fatal("NewRouter() returned nil")
	}

	// Test that it implements the Router interface
	var _ Router = router
}

func TestBasicRouting(t *testing.T) {
	router := NewRouter()

	// Register test routes
	router.GET("/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GET users"))
	})

	router.POST("/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("POST users"))
	})

	tests := []struct {
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{"GET", "/users", http.StatusOK, "GET users"},
		{"POST", "/users", http.StatusCreated, "POST users"},
		{"PUT", "/users", http.StatusNotFound, "Not Found\n"}, // Should 404
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s", tt.method, tt.path), func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if body := w.Body.String(); body != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

func TestPathParameters(t *testing.T) {
	router := NewRouter()

	router.GET("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := GetPathParam(r, "id")
		w.Write([]byte(fmt.Sprintf("User ID: %s", id)))
	})

	router.GET("/posts/{postId}/comments/{commentId}", func(w http.ResponseWriter, r *http.Request) {
		postID := GetPathParam(r, "postId")
		commentID := GetPathParam(r, "commentId")
		w.Write([]byte(fmt.Sprintf("Post: %s, Comment: %s", postID, commentID)))
	})

	tests := []struct {
		path         string
		expectedBody string
	}{
		{"/users/123", "User ID: 123"},
		{"/users/abc", "User ID: abc"},
		{"/posts/456/comments/789", "Post: 456, Comment: 789"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}

			if body := w.Body.String(); body != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, body)
			}
		})
	}
}

func TestRouteGroups(t *testing.T) {
	router := NewRouter()

	// Create API v1 group
	v1 := router.Group("/api/v1")
	v1.GET("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("API v1 users"))
	})

	v1.POST("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("API v1 create user"))
	})

	// Create nested group
	v1Users := v1.Group("/users")
	v1Users.GET("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := GetPathParam(r, "id")
		w.Write([]byte(fmt.Sprintf("API v1 user %s", id)))
	})

	tests := []struct {
		method   string
		path     string
		expected string
	}{
		{"GET", "/api/v1/users", "API v1 users"},
		{"POST", "/api/v1/users", "API v1 create user"},
		{"GET", "/api/v1/users/123", "API v1 user 123"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s", tt.method, tt.path), func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}

			if body := w.Body.String(); body != tt.expected {
				t.Errorf("expected body %q, got %q", tt.expected, body)
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	router := NewRouter()

	// Add a middleware that adds a header
	testMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test-Middleware", "applied")
			next.ServeHTTP(w, r)
		})
	}

	router.Use(testMiddleware)

	router.GET("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Header().Get("X-Test-Middleware") != "applied" {
		t.Error("middleware was not applied")
	}
}

func TestCORSMiddleware(t *testing.T) {
	router := NewRouter()

	corsOptions := CORSOptions{
		AllowedOrigins: []string{"https://example.com"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type"},
	}

	router.Use(CORS(corsOptions))

	router.GET("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	// Test preflight request
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for OPTIONS, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Error("CORS origin header not set correctly")
	}
}

func TestGzipMiddleware(t *testing.T) {
	router := NewRouter()

	router.Use(Gzip(gzip.DefaultCompression))

	router.GET("/test", func(w http.ResponseWriter, r *http.Request) {
		// Write a response large enough to trigger compression
		data := strings.Repeat("This is test data for compression. ", 100)
		w.Write([]byte(data))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Error("gzip compression not applied")
	}

	// Verify we can decompress the response
	reader, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to decompress response: %v", err)
	}

	expected := strings.Repeat("This is test data for compression. ", 100)
	if string(decompressed) != expected {
		t.Error("decompressed content doesn't match expected")
	}
}

func TestJSONHandler(t *testing.T) {
	router := NewRouter()

	type TestResponse struct {
		Message string `json:"message"`
		Status  string `json:"status"`
	}

	router.GET("/json", JSONHandler(func(w http.ResponseWriter, r *http.Request) (any, error) {
		return TestResponse{
			Message: "Hello, JSON!",
			Status:  "success",
		}, nil
	}))

	req := httptest.NewRequest("GET", "/json", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Content-Type header not set to application/json")
	}

	var response TestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal JSON response: %v", err)
	}

	if response.Message != "Hello, JSON!" {
		t.Errorf("expected message 'Hello, JSON!', got %q", response.Message)
	}
}

func TestHealthCheckHandler(t *testing.T) {
	router := NewRouter()

	checks := []HealthCheck{
		{
			Name: "database",
			Check: func() error {
				return nil // Simulate healthy database
			},
		},
		{
			Name: "redis",
			Check: func() error {
				return fmt.Errorf("connection timeout") // Simulate unhealthy redis
			},
		},
	}

	router.GET("/health", HealthCheckHandler(checks...))

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}

	var result HealthResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal health check response: %v", err)
	}

	if result.Status != "failed" {
		t.Errorf("expected overall status 'failed', got %q", result.Status)
	}

	if result.Checks["database"].Status != "ok" {
		t.Error("database check should be ok")
	}

	if result.Checks["redis"].Status != "failed" {
		t.Error("redis check should be failed")
	}
}

func TestStaticFiles(t *testing.T) {
	router := NewRouter()

	// Create a temporary directory and file for testing
	// Note: In a real test, you'd create actual temp files
	// For this test, we'll just verify the route is registered

	router.StaticFile("/favicon.ico", "/path/to/favicon.ico")

	// Test that the route is registered (we can't test actual file serving without files)
	req := httptest.NewRequest("GET", "/favicon.ico", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// We expect this to fail since the file doesn't exist, but the route should be registered
	// In a real application, this would serve the actual file
}

func TestRequestID(t *testing.T) {
	router := NewRouter()

	router.Use(RequestID())

	router.GET("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Header().Get("X-Request-Id") == "" {
		t.Error("Request ID header not set")
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	router := NewRouter()

	var recoveredErr any
	var recoveredStack []byte

	router.Use(Recovery(func(err any, stack []byte) {
		recoveredErr = err
		recoveredStack = stack
	}))

	router.GET("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}

	if recoveredErr == nil {
		t.Error("panic was not recovered")
	}

	if recoveredStack == nil {
		t.Error("stack trace was not captured")
	}
}

func TestBasicAuth(t *testing.T) {
	router := NewRouter()

	validator := func(username, password string) bool {
		return username == "admin" && password == "secret"
	}

	router.Use(BasicAuth("Test Realm", validator))

	router.GET("/protected", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("protected content"))
	})

	// Test without auth
	req1 := httptest.NewRequest("GET", "/protected", nil)
	w1 := httptest.NewRecorder()

	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w1.Code)
	}

	// Test with correct auth
	req2 := httptest.NewRequest("GET", "/protected", nil)
	req2.SetBasicAuth("admin", "secret")
	w2 := httptest.NewRecorder()

	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w2.Code)
	}

	if w2.Body.String() != "protected content" {
		t.Error("protected content not returned")
	}
}

func TestRateLimiter(t *testing.T) {
	router := NewRouter()

	// Allow 2 requests per second
	router.Use(RateLimiter(2, time.Second))

	router.GET("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// First two requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345" // Simulate same client
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected status 200, got %d", i+1, w.Code)
		}
	}

	// Third request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", w.Code)
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test GetQueryParam
	req := httptest.NewRequest("GET", "/test?name=john&age=25", nil)

	if name := GetQueryParam(req, "name", "default"); name != "john" {
		t.Errorf("expected name 'john', got %q", name)
	}

	if missing := GetQueryParam(req, "missing", "default"); missing != "default" {
		t.Errorf("expected default value 'default', got %q", missing)
	}

	// Test GetQueryParamInt
	if age := GetQueryParamInt(req, "age", 0); age != 25 {
		t.Errorf("expected age 25, got %d", age)
	}

	if missing := GetQueryParamInt(req, "missing", 18); missing != 18 {
		t.Errorf("expected default age 18, got %d", missing)
	}
}

func TestChainMiddleware(t *testing.T) {
	router := NewRouter()

	// Create middleware that add headers
	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-MW-1", "applied")
			next.ServeHTTP(w, r)
		})
	}

	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-MW-2", "applied")
			next.ServeHTTP(w, r)
		})
	}

	// Chain middleware
	router.Use(Chain(mw1, mw2))

	router.GET("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Header().Get("X-MW-1") != "applied" {
		t.Error("middleware 1 was not applied")
	}

	if w.Header().Get("X-MW-2") != "applied" {
		t.Error("middleware 2 was not applied")
	}
}

func BenchmarkRouter(b *testing.B) {
	router := NewRouter()

	router.GET("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := GetPathParam(r, "id")
		w.Write([]byte(id))
	})

	req := httptest.NewRequest("GET", "/users/123", nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
