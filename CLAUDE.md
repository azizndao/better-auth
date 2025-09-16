# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Better Auth Go is a framework-agnostic authentication and authorization library for Go applications. It uses standard `net/http` interfaces with zero external framework dependencies, making it pluggable into any Go web application.

## Development Commands

### Building and Testing
```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -cover ./...

# Run specific test
go test -v -run TestBetterAuth_SignUp

# Build and run examples
go run examples/plugin-demo/main.go
go run examples/custom-database/main.go

# Module management
go mod tidy
go mod download
```

### No Build Tools
This project doesn't use Make, npm scripts, or other build tools. Use standard Go commands.

## Architecture Overview

### Key Components

1. **BetterAuth Core** (`internal/auth/core.go`) - Main authentication engine that orchestrates all authentication operations
2. **Plugin System** (`pkg/plugins/core/`) - Extensible architecture with conditional loading, service registry, and route management
3. **Transport Layer** (`pkg/transport/`) - HTTP request/response abstraction with JSON handling, validation, and context management
4. **Database Layer** (`internal/database/`) - GORM-based persistence with custom dialector support and conditional migrations
5. **Middleware** (`pkg/middleware/`) - Authentication and authorization middleware with transport integration

### Package Structure

- `internal/` - Core implementation (auth, config, database, models)
- `pkg/` - Public packages (middleware, plugins, transport, router)
- `examples/` - Usage demonstrations (plugin-demo, custom-database)
- Root files - Main package entry point (`betterauth.go`) and comprehensive tests (`auth_test.go`)

### Database Configuration Approach

**Custom Dialector Support**: Users can now pass any GORM dialector and configuration instead of relying on built-in drivers:
```go
// PostgreSQL example
dialector := postgres.Open("connection-string")
dbConfig := betterauth.NewDatabaseConfig(dialector, &gorm.Config{...})
auth, err := betterauth.NewWithDatabase(cfg, dbConfig)
```

This eliminates driver dependencies and provides maximum flexibility for database configuration.

### Plugin Architecture

**Conditional Loading System**: Plugins are registered but only enabled with configuration validation. This ensures:
- Database tables are created only when plugins are enabled
- Services are registered dynamically
- Routes are applied without conflicts (duplicate route protection)
- Lifecycle hooks (OnEnabled/OnDisabled) are called appropriately

**Plugin Capabilities**:
- Add database tables (via GORM models with conditional migrations)
- Register HTTP routes with automatic conflict resolution
- Provide middleware and services
- Validate configuration before enabling

**Available Plugins**:
- **JWT Plugin** (`pkg/plugins/jwt/`) - Token management with refresh tokens and blacklisting
- **Organizations Plugin** (`pkg/plugins/organizations/`) - Multi-tenant support with RBAC
- **Admin Plugin** (`pkg/plugins/admin/`) - User management with audit logging and system metrics

**Plugin Interface Implementation**:
```go
type Plugin interface {
    Name() string
    Initialize(ctx context.Context, db *gorm.DB) error
    DatabaseModels() []DatabaseModel
    Services() []PluginService
    Routes() map[string]http.Handler
    OnEnabled() error
    OnDisabled() error
    RequiredConfig() map[string]interface{}
    ValidateConfig(config map[string]interface{}) error
}
```

### Database Support

- **Primary**: PostgreSQL, MySQL, SQLite
- **Testing**: In-memory SQLite
- **ORM**: GORM with automatic migrations
- **Models**: User, Session, Organization, OAuthAccount + plugin-specific tables

### Framework Integration

Uses standard `net/http.Handler` interface, allowing integration with:
- net/http (direct)
- Chi, Gorilla Mux (mount/prefix)
- Gin, Echo, Fiber (via adapters)

### Security Features

- bcrypt password hashing
- JWT with configurable expiry
- Session management with device tracking
- CORS support
- Rate limiting (plugin-based)
- Input validation throughout

## Testing Approach

- **Single comprehensive test file**: `auth_test.go`
- **Framework**: testify for assertions  
- **Database**: Unique SQLite files per test (using `createTestAuth()` helper)
- **Isolation**: Each test uses a separate database to prevent conflicts
- **Coverage**: All core flows, middleware, database operations, plugins, transport layer
- **Test Helper**: `createTestAuth(t *testing.T)` creates isolated test instances with cleanup

## Configuration Patterns

Configuration is centralized in `internal/config/config.go` with:
- Database URL (supports multiple dialects)
- JWT/Session settings
- OAuth providers
- Email configuration
- CORS policies

## Important Implementation Notes

### Framework Agnostic Design
- All HTTP handling uses standard `net/http` interfaces
- Transport layer abstracts request/response operations (JSON handling, validation, context management)
- Middleware follows standard `func(http.Handler) http.Handler` pattern
- **Transport Integration**: Always use `transport.Transport` interface methods instead of manual JSON/HTTP handling

### Plugin Conditional Loading
- Tables only created when plugins are enabled (conditional GORM migrations)
- Prevents unnecessary database overhead
- Services registered dynamically via ServiceRegistry
- Route conflicts prevented by tracking applied routes
- Plugin lifecycle properly managed (Initialize → Enable → Disable)

### Transport Layer Usage
- **Use `transport.ExtractToken(r)`** instead of manual Authorization header parsing
- **Use `transport.SetUserContext(r, userCtx)`** instead of `context.WithValue()` 
- **Use `transport.RespondJSON(w, status, data)`** for consistent JSON responses
- **Use `transport.RespondError(w, status, message)`** for error responses

### Security Considerations
- Always use parameterized queries (GORM handles this)
- Password validation and hashing with bcrypt
- Token extraction supports multiple methods (header, cookie, query) via transport
- Secure cookie settings for production
- User context properly managed through transport layer

## Common Development Tasks

When working with this codebase:

1. **Adding new endpoints**: Add to appropriate handler file in `internal/auth/` and use transport methods for JSON/error handling
2. **Creating plugins**: 
   - Implement the `Plugin` interface in `pkg/plugins/[name]/`
   - Use `BasePlugin` from `pkg/plugins/core/` for common functionality
   - Add models, services, handlers, and plugin.go files
   - Test plugin registration, enabling, and route conflicts
3. **Database changes**: 
   - Add GORM models with proper tags and table names
   - Use conditional migrations via plugin system
   - Test with custom dialectors (PostgreSQL, MySQL, SQLite)
4. **Transport customization**: 
   - Always use existing `transport.Transport` interface methods
   - Extend via composition rather than modification
   - Maintain consistency with JSON response formats
5. **Middleware**: 
   - Add to `pkg/middleware/` using transport for token extraction and context
   - Follow standard `func(http.Handler) http.Handler` pattern
   - Use `transport.SetUserContext()` instead of manual context management
6. **Testing**: 
   - Use `createTestAuth(t)` helper for isolated test instances
   - Test database operations with unique SQLite files
   - Verify plugin functionality and route registration

## Dependencies

Key external dependencies:
- `github.com/golang-jwt/jwt/v5` - JWT handling
- `gorm.io/gorm` + drivers - Database ORM
- `golang.org/x/crypto` - Cryptographic functions
- `github.com/stretchr/testify` - Testing
- `github.com/google/uuid` - UUID generation

Go version: 1.24.5+ (toolchain: go1.24.5)

## Go Development Guidelines

### Always Use Latest Go Version
- This project uses Go 1.24.5+ to leverage the latest language features and performance improvements
- When adding new code, prefer modern Go idioms and features available in Go 1.24+
- Use the latest standard library capabilities where appropriate

### Go 1.24 Key Features to Utilize
- **Generic Type Aliases**: Use parameterized type aliases for better type abstraction
- **Tool Dependencies**: Use `go get -tool` for managing executable dependencies in go.mod
- **Swiss Tables Maps**: Benefit from ~30% faster map operations for large maps (>1024 entries)
- **Enhanced Testing**: Use `testing.B.Loop()` for more efficient benchmark iterations
- **FIPS 140-3 Compliance**: Leverage new cryptographic mechanisms for approved algorithms
- **Improved Performance**: Expect 2-3% CPU overhead reduction from runtime optimizations

### Modern Go Features to Continue Using
- **Generics**: Use for type-safe collections and utility functions
- **Context**: Always pass context.Context for cancellation and timeouts
- **Error Wrapping**: Use `fmt.Errorf` with `%w` verb for error chains
- **Structured Logging**: Prefer `slog` from Go 1.21+ for structured logging
- **HTTP Routing**: Leverage Go 1.22+ enhanced ServeMux pattern matching
- **Range Functions**: Use Go 1.23 range-over-function iterators where beneficial

## Current Architecture State

### Transport Layer Integration Status
The codebase has been partially updated to use the `pkg/transport/` package consistently:

**✅ Completed**:
- Handlers in `internal/auth/handlers.go` properly use transport methods
- Database configuration supports custom GORM dialectors
- Plugin system has conditional loading and route conflict prevention

**🚧 In Progress**:
- Middleware in `internal/auth/middleware.go` is being updated to use transport methods
- Some manual context management being replaced with transport methods
- Token extraction being unified through transport interface

**Key Integration Patterns**:
```go
// Use transport methods instead of manual handling
token := transport.ExtractToken(r)                    // ✅ Not: manual header parsing
userCtx := &models.UserContext{User: user}
r = transport.SetUserContext(r, userCtx)              // ✅ Not: context.WithValue()
transport.RespondJSON(w, http.StatusOK, data)         // ✅ Not: manual JSON encoding
transport.RespondError(w, http.StatusBadRequest, msg) // ✅ Not: manual error responses
```

When working on middleware or handlers, always check if transport methods are being used consistently throughout the request lifecycle.