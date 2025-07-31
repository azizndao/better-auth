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

1. **AuthCore** (`internal/auth/core.go`) - Main authentication engine that orchestrates all authentication operations
2. **Plugin System** (`pkg/plugins/core/`) - Extensible architecture allowing conditional feature loading
3. **Transport Layer** (`pkg/transport/`) - HTTP request/response abstraction for framework independence
4. **Database Layer** (`internal/database/`) - GORM-based persistence with multi-database support
5. **Middleware** (`pkg/middleware/`) - Authentication and authorization middleware

### Package Structure

- `internal/` - Core implementation (auth, config, database, models)
- `pkg/` - Public packages (middleware, plugins, transport)
- `examples/` - Usage demonstrations
- Root files - Main package entry point and comprehensive tests

### Plugin Architecture

Plugins are conditionally loaded and can:
- Add database tables (via GORM models)
- Register HTTP routes
- Provide middleware
- Add services to the registry

Available plugins: JWT, Organizations, Admin

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
- **Database**: In-memory SQLite
- **Coverage**: All core flows, middleware, database operations, plugins

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
- Transport layer abstracts request/response operations
- Middleware follows standard `func(http.Handler) http.Handler` pattern

### Plugin Conditional Loading
- Tables only created when plugins are enabled
- Prevents unnecessary database overhead
- Services registered dynamically based on enabled plugins

### Security Considerations
- Always use parameterized queries
- Password validation and hashing
- Token extraction supports multiple methods (header, cookie, query)
- Secure cookie settings for production

## Common Development Tasks

When working with this codebase:

1. **Adding new endpoints**: Add to appropriate handler file in `internal/auth/`
2. **Creating plugins**: Implement the `Plugin` interface in `pkg/plugins/core/`
3. **Database changes**: Add migrations via GORM models
4. **Transport customization**: Extend or override methods in `pkg/transport/`
5. **Middleware**: Add to `pkg/middleware/` following standard patterns

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