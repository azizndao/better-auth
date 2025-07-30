# Better Auth Go - Framework Agnostic Authentication

A comprehensive, **framework-agnostic** authentication and authorization library for Go, inspired by Better Auth. This implementation provides all essential authentication features using the latest **net/http** capabilities with zero external framework dependencies, making it **pluggable into any Go web application**.

## 🚀 Key Features

### ✅ **Framework Agnostic Design**
- Uses standard `net/http` interfaces
- No external framework dependencies
- Plugs into any Go web application
- Works with Chi, Gorilla Mux, Gin, Echo, Fiber, or plain net/http

### 🔐 **Complete Authentication System**
- **Email/Password Authentication** - Secure user registration and login
- **Session Management** - Robust session handling with customizable expiry
- **JWT Support** - Full JWT token generation and validation
- **Two-Factor Authentication** - TOTP-based 2FA with QR codes and backup codes
- **Password Reset** - Secure password reset via email tokens
- **Email Verification** - Email verification workflow

### 🌐 **OAuth Integration**
- **Multiple Providers** - Google, GitHub, Discord, Microsoft, Facebook, Twitter
- **Automatic Account Linking** - Link OAuth accounts to existing users
- **Provider-Specific Normalization** - Consistent user data across providers

### 🏢 **Organization Management** 
- **Multi-Tenant Support** - Organization and team management
- **Role-Based Access Control** - Flexible role and permission system
- **Member Invitations** - Email-based organization invitations
- **Access Control** - Fine-grained permission management

### 🔧 **Advanced Features**
- **Plugin Architecture** - Extensible plugin system
- **Rate Limiting** - Configurable rate limiting with IP-based tracking
- **Database Agnostic** - Support for PostgreSQL, MySQL, SQLite, and in-memory
- **Audit Logging** - Comprehensive request and action logging
- **IP Whitelisting** - Restrict access by IP addresses
- **Admin Panel** - Built-in admin routes for user management

### 🛡️ **Security First**
- **Password Hashing** - bcrypt for secure password storage
- **CORS Support** - Built-in CORS middleware
- **Security Headers** - Automatic security headers
- **Input Validation** - Comprehensive request validation
- **SQL Injection Prevention** - Parameterized queries throughout

## 🚀 Quick Start

### Installation

```bash
go get github.com/your-username/better-auth-go
```

### Basic Usage

```go
package main

import (
    "log"
    "net/http"
    "time"
    
    betterauth "better-auth"
)

func main() {
    // Configure Better Auth
    config := &betterauth.Config{
        DatabaseURL:      "sqlite:./auth.db",
        SecretKey:        "your-secret-key-here-make-it-long-and-secure",
        SessionExpiry:    24 * time.Hour,
        JWTExpiry:        1 * time.Hour,
        BaseURL:          "http://localhost:8080",
        PathPrefix:       "/auth",
        CORSConfig:       betterauth.DefaultCORSConfig(),
    }

    // Initialize Better Auth
    auth, err := betterauth.New(config)
    if err != nil {
        log.Fatal("Failed to initialize Better Auth:", err)
    }

    // Add plugins
    auth.Use(betterauth.NewAdminPlugin())
    auth.Use(betterauth.NewRateLimitPlugin(time.Minute, 100))

    // Standard net/http usage
    mux := http.NewServeMux()
    mux.Handle("/auth/", auth)  // Mount auth routes
    
    // Protect your routes
    mux.HandleFunc("/api/profile", auth.AuthMiddleware(http.HandlerFunc(profileHandler)).ServeHTTP)
    
    log.Fatal(http.ListenAndServe(":8080", mux))
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
    // Access user context
    userCtx, _ := auth.GetTransport().GetUserContext(r)
    // ... your handler logic
}
```

## 🔌 Framework Integration Examples

### Standalone net/http

```go
mux := http.NewServeMux()
mux.Handle("/auth/", auth)
mux.Handle("/api/", auth.AuthMiddleware(apiHandler))
http.ListenAndServe(":8080", mux)
```

### Chi Router

```go
import "github.com/go-chi/chi/v5"

r := chi.NewRouter()
r.Mount("/auth", auth)
r.Route("/api", func(r chi.Router) {
    r.Use(auth.AuthMiddleware)
    r.Get("/profile", profileHandler)
})
http.ListenAndServe(":8080", r)
```

### Gorilla Mux

```go
import "github.com/gorilla/mux"

r := mux.NewRouter()
r.PathPrefix("/auth").Handler(auth)
api := r.PathPrefix("/api").Subrouter()
api.Use(auth.AuthMiddleware)
api.HandleFunc("/profile", profileHandler)
http.ListenAndServe(":8080", r)
```

### Gin (Wrapped)

```go
import "github.com/gin-gonic/gin"

r := gin.Default()
r.Any("/auth/*path", func(c *gin.Context) {
    auth.ServeHTTP(c.Writer, c.Request)
})

// Custom middleware adapter
r.Use(func(c *gin.Context) {
    auth.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        c.Request = r
        c.Next()
    })).ServeHTTP(c.Writer, c.Request)
})

r.GET("/api/profile", profileHandler)
r.Run(":8080")
```

### Echo Framework

```go
import "github.com/labstack/echo/v4"

e := echo.New()
e.Any("/auth/*", func(c echo.Context) error {
    auth.ServeHTTP(c.Response(), c.Request())
    return nil
})

// Middleware adapter
e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        auth.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            c.SetRequest(r)
            next(c)
        })).ServeHTTP(c.Response(), c.Request())
        return nil
    }
})

e.GET("/api/profile", profileHandler)
e.Start(":8080")
```

### Fiber (v2)

```go
import "github.com/gofiber/fiber/v2"
import "github.com/gofiber/fiber/v2/middleware/adaptor"

app := fiber.New()
app.All("/auth/*", adaptor.HTTPHandler(auth))

// Middleware adapter
app.Use(func(c *fiber.Ctx) error {
    return adaptor.HTTPMiddleware(auth.AuthMiddleware)(c)
})

app.Get("/api/profile", adaptor.HTTPHandlerFunc(profileHandler))
app.Listen(":8080")
```

## 📖 API Endpoints

### Authentication
- `POST /auth/sign-up` - User registration
- `POST /auth/sign-in` - User login  
- `POST /auth/sign-out` - User logout
- `GET /auth/session` - Get current session
- `POST /auth/reset-password` - Request password reset
- `POST /auth/verify-email` - Verify email address

### Two-Factor Authentication
- `POST /auth/two-factor/setup` - Setup 2FA (requires auth)
- `POST /auth/two-factor/verify` - Verify 2FA code (requires auth)

### OAuth
- `GET /auth/oauth/{provider}` - Initiate OAuth flow
- `GET /auth/oauth/{provider}/callback` - OAuth callback

### Organizations  
- `POST /auth/organization/create` - Create organization (requires auth)
- `GET /auth/organization/{id}` - Get organization details (requires auth)
- `POST /auth/organization/{id}/invite` - Invite user to organization (requires auth)
- `POST /auth/organization/{id}/members/{userId}/role` - Update member role (requires auth)

### Admin (Requires admin role)
- `GET /auth/admin/users` - List all users
- `GET /auth/admin/users/{id}` - Get user details
- `PUT /auth/admin/users/{id}/block` - Block user
- `PUT /auth/admin/users/{id}/unblock` - Unblock user
- `DELETE /auth/admin/users/{id}` - Delete user
- `GET /auth/admin/sessions` - List active sessions
- `DELETE /auth/admin/sessions/{token}` - Delete session
- `GET /auth/admin/stats` - System statistics

## 🔧 Configuration

### Database Support

```go
// PostgreSQL
config.DatabaseURL = "postgres://user:password@localhost/dbname?sslmode=disable"

// MySQL  
config.DatabaseURL = "mysql://user:password@localhost/dbname"

// SQLite
config.DatabaseURL = "sqlite:./auth.db"

// In-Memory (for testing)
config.DatabaseURL = "memory"
```

### OAuth Providers

```go
config.Providers = []betterauth.OAuthProvider{
    {
        Name:         "google",
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret", 
        RedirectURL:  "http://localhost:8080/auth/oauth/google/callback",
        Scopes:       []string{"openid", "email", "profile"},
    },
    {
        Name:         "github",
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        RedirectURL:  "http://localhost:8080/auth/oauth/github/callback", 
        Scopes:       []string{"user:email"},
    },
}
```

### Email Configuration

```go
config.EmailConfig = &betterauth.EmailConfig{
    Provider: "smtp",
    From:     "noreply@yourapp.com",
    SMTP: &betterauth.SMTPConfig{
        Host:     "smtp.gmail.com",
        Port:     587,
        Username: "your-email@gmail.com",
        Password: "your-app-password",
        TLS:      true,
    },
}
```

### CORS Configuration

```go
config.CORSConfig = &betterauth.CORSConfig{
    AllowedOrigins:   []string{"http://localhost:3000", "https://yourdomain.com"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With"},
    AllowCredentials: true,
}
```

## 🔌 Plugin System

### Built-in Plugins

#### Admin Plugin
Provides administrative endpoints for user management.

```go
adminPlugin := betterauth.NewAdminPlugin()
auth.Use(adminPlugin)
```

#### Rate Limit Plugin  
Configurable rate limiting by IP address.

```go
rateLimitPlugin := betterauth.NewRateLimitPlugin(time.Minute, 100) // 100 requests per minute
auth.Use(rateLimitPlugin)
```

#### Audit Log Plugin
Logs all requests and authentication events.

```go
auditLogPlugin := betterauth.NewAuditLogPlugin()
auth.Use(auditLogPlugin)
```

#### IP Whitelist Plugin
Restrict access to specific IP addresses.

```go
ipWhitelistPlugin := betterauth.NewIPWhitelistPlugin([]string{"127.0.0.1", "192.168.1.0/24"})
auth.Use(ipWhitelistPlugin)
```

### Custom Plugins

Create custom plugins by implementing the `Plugin` interface:

```go
type MyPlugin struct {
    auth *betterauth.BetterAuth
}

func (p *MyPlugin) Name() string {
    return "my-plugin"
}

func (p *MyPlugin) Initialize(auth *betterauth.BetterAuth) error {
    p.auth = auth
    return nil
}

func (p *MyPlugin) RegisterRoutes(mux *http.ServeMux, pathPrefix string) error {
    mux.HandleFunc(fmt.Sprintf("GET %s/my-endpoint", pathPrefix), p.myHandler)
    return nil
}

func (p *MyPlugin) myHandler(w http.ResponseWriter, r *http.Request) {
    p.auth.GetTransport().RespondJSON(w, http.StatusOK, map[string]string{
        "message": "Hello from my plugin",
    })
}

// Optionally implement middleware
func (p *MyPlugin) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Custom middleware logic
        next.ServeHTTP(w, r)
    })
}
```

## 🔒 Middleware Usage

### Authentication Middleware

```go
// Protect routes that require authentication
protectedMux := http.NewServeMux()
protectedMux.HandleFunc("/profile", getProfile)
protectedMux.HandleFunc("/settings", getSettings)

// Apply auth middleware
protectedHandler := auth.AuthMiddleware(protectedMux)
mux.Handle("/api/", protectedHandler)
```

### JWT Middleware

```go
// Use JWT-based authentication instead of sessions
jwtProtectedHandler := auth.JWTMiddleware(protectedMux)
mux.Handle("/jwt-api/", jwtProtectedHandler)
```

### Role-based Middleware

```go
// Require specific roles
adminMux := http.NewServeMux()
adminMux.HandleFunc("/dashboard", adminDashboard)

adminHandler := auth.AuthMiddleware(auth.RequireRole("admin")(adminMux))
mux.Handle("/admin/", adminHandler)

// Require specific permissions  
protectedHandler := auth.AuthMiddleware(auth.RequirePermission("read:users")(protectedMux))
mux.Handle("/protected/", protectedHandler)
```

### Middleware Chaining

```go
chain := betterauth.NewMiddlewareChain()
chain.Use(auth.CORSMiddleware)
chain.Use(auth.LoggingMiddleware)
chain.Use(auth.RecoveryMiddleware)
chain.Use(auth.SecurityHeadersMiddleware)
chain.Use(auth.AuthMiddleware)

handler := chain.Handler(http.HandlerFunc(protectedEndpoint))
```

## 🔍 Custom Transport

Customize how Better Auth handles requests and responses:

```go
type MyTransport struct {
    *betterauth.DefaultTransport
}

func (t *MyTransport) RespondJSON(w http.ResponseWriter, status int, data interface{}) {
    // Custom JSON response logic
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Custom-Header", "MyValue")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": status < 400,
        "data":    data,
    })
}

auth.SetTransport(&MyTransport{DefaultTransport: &betterauth.DefaultTransport{}})
```

## 🧪 Testing

Run the comprehensive test suite:

```bash
go test -v ./...
```

The test suite includes:
- Unit tests for all core functionality
- Integration tests for API endpoints  
- Database tests for all supported databases
- JWT token generation and validation tests
- Rate limiting tests
- Session management tests
- Plugin system tests
- Middleware tests

## 📊 Example Usage

```bash
# Sign up a new user
curl -X POST http://localhost:8080/auth/sign-up \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"SecurePass123!","name":"Test User"}'

# Sign in
curl -X POST http://localhost:8080/auth/sign-in \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"SecurePass123!"}'

# Get session (use token from sign-in response)
curl -X GET http://localhost:8080/auth/session \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# Access protected endpoint
curl -X GET http://localhost:8080/api/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## 🚦 Production Considerations

1. **Secret Key**: Use a strong, randomly generated secret key for JWT signing
2. **Database Security**: Use connection pooling and proper database credentials  
3. **HTTPS**: Always use HTTPS in production
4. **Rate Limiting**: Enable rate limiting to prevent brute force attacks
5. **Password Policy**: Implement strong password requirements
6. **Session Security**: Configure secure session cookies
7. **Input Validation**: All inputs are validated and sanitized
8. **Environment Variables**: Store sensitive config in environment variables

## 🆚 Why Better Auth Go?

### vs. Other Go Auth Libraries

- **Framework Agnostic**: Works with ANY Go web framework
- **Zero Dependencies**: Uses only standard library `net/http`
- **Plugin Architecture**: Extensible and customizable
- **Complete Feature Set**: Everything you need out of the box
- **Production Ready**: Battle-tested security practices
- **Easy Integration**: Drop-in authentication for existing applications

### Framework Compatibility Matrix

| Framework | Integration Method | Complexity |
|-----------|-------------------|------------|
| net/http  | Direct            | ⭐ Simple  |
| Chi       | Mount/Use         | ⭐ Simple  |  
| Gorilla   | PathPrefix        | ⭐ Simple  |
| Gin       | Wrapper           | ⭐⭐ Easy   |
| Echo      | Adapter           | ⭐⭐ Easy   |
| Fiber     | Adaptor           | ⭐⭐ Easy   |

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch  
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## 📄 License

MIT License - see LICENSE file for details.

## 🙏 Support

- 🐛 Create an issue for bug reports
- 💬 Join our Discord for community support  
- 📚 Check the documentation for detailed guides
- ⭐ Star the repository if you find it useful

---

**Built with ❤️ for the Go community - Framework agnostic, production ready, developer friendly authentication.**

## 🔗 Links

- [Documentation](https://github.com/your-username/better-auth-go/wiki)
- [Examples](https://github.com/your-username/better-auth-go/tree/main/examples)
- [Plugin Gallery](https://github.com/your-username/better-auth-go/tree/main/plugins)
- [Migration Guide](https://github.com/your-username/better-auth-go/blob/main/MIGRATION.md)