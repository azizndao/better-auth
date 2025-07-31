# Better Auth Plugin System

This document describes the plugin system architecture and how to use the available plugins.

## Architecture Overview

The Better Auth plugin system provides:

- **Conditional Loading**: Tables and services are only created when plugins are enabled
- **Plugin Registry**: Centralized management of plugins with lifecycle hooks
- **Database Integration**: Automatic GORM model migration for plugin tables
- **Service Registry**: Plugin services are registered and available to other components
- **Middleware Support**: Plugins can provide HTTP middleware

## Core Components

### Plugin Interface

```go
type Plugin interface {
    Name() string
    Initialize(ctx context.Context, db *gorm.DB) error
    Handler() http.Handler
    Routes() map[string]http.Handler
    
    // Database-related methods
    DatabaseModels() []DatabaseModel
    Services() []PluginService
    
    // Plugin lifecycle
    OnEnabled() error
    OnDisabled() error
    
    // Configuration
    RequiredConfig() map[string]interface{}
    ValidateConfig(config map[string]interface{}) error
}
```

### Plugin Registry

The `PluginRegistry` manages plugins with conditional loading:

- **Register**: Add a plugin to the registry (doesn't enable it)
- **Enable**: Enable a plugin with configuration validation and table migration
- **Disable**: Disable a plugin and clean up resources
- **Conditional Migration**: Only migrate tables for enabled plugins

## Available Plugins

### 1. Organizations Plugin (`organizations`)

Provides multi-tenant organization management with teams and permissions.

**Models**:
- `Organization`: Company/organization entity
- `OrganizationMember`: User membership in organizations
- `Team`: Teams within organizations
- `TeamMember`: User membership in teams
- `OrganizationInvitation`: Invitations to join organizations

**Features**:
- Organization CRUD operations
- Member management with roles and permissions
- Team management
- Invitation system
- Role-based access control

```go
orgPlugin := organizations.NewOrganizationPlugin()
auth.RegisterPlugin(orgPlugin)
auth.EnablePlugin(ctx, "organizations", map[string]interface{}{
    "enabled": true,
})
```

### 2. JWT Plugin (`jwt`)

Advanced JWT token management with refresh tokens and blacklisting.

**Models**:
- `JWTRefreshToken`: Refresh tokens storage
- `JWTBlacklist`: Blacklisted access tokens

**Features**:
- Access and refresh token generation
- Token refresh mechanism  
- Token blacklisting
- Automatic cleanup of expired tokens
- Configurable token lifetimes

```go
jwtPlugin := jwt.NewJWTPlugin(&jwt.JWTConfig{
    SecretKey:              []byte("your-secret-key"),
    AccessTokenExpiry:      15 * time.Minute,
    RefreshTokenExpiry:     7 * 24 * time.Hour,
    EnableRefreshTokens:    true,
    EnableTokenBlacklist:   true,
    MaxRefreshTokensPerUser: 5,
})

auth.RegisterPlugin(jwtPlugin)
auth.EnablePlugin(ctx, "jwt", map[string]interface{}{
    "enabled":              true,
    "secretKey":            "your-secret-key",
    "enableRefreshTokens":  true,
    "enableTokenBlacklist": true,
})
```

### 3. Admin Plugin (`admin`)

Administrative interface with system monitoring and audit logging.

**Models**:
- `AdminUser`: Admin user privileges and permissions
- `AdminSession`: Admin login sessions
- `AdminAuditLog`: Audit trail of admin actions
- `SystemMetrics`: System performance metrics

**Features**:
- Admin user management with roles
- Comprehensive audit logging
- System metrics collection
- Permission-based access control
- Activity monitoring

```go
adminPlugin := admin.NewAdminPlugin()
auth.RegisterPlugin(adminPlugin)
auth.EnablePlugin(ctx, "admin", map[string]interface{}{
    "enabled": true,
})
```

## Usage Example

```go
package main

import (
    "context"
    "log"
    "time"

    "better-auth/internal/auth"
    "better-auth/internal/config"
    "better-auth/internal/database"
    "better-auth/pkg/plugins/admin"
    "better-auth/pkg/plugins/jwt"
    "better-auth/pkg/plugins/organizations"
)

func main() {
    ctx := context.Background()

    // Database configuration
    dbConfig := &database.DatabaseConfig{
        Type:     "sqlite",
        FilePath: "./better-auth.db",
    }

    // Auth configuration
    cfg := &config.Config{
        SecretKey:     "your-secret-key-at-least-32-characters-long",
        SessionExpiry: 24 * time.Hour,
        PathPrefix:    "/auth",
    }

    // Create auth system
    betterAuth, err := auth.NewFromDatabaseConfig(dbConfig, cfg)
    if err != nil {
        log.Fatal("Failed to create auth system:", err)
    }

    // Create and register plugins
    jwtPlugin := jwt.NewJWTPlugin(&jwt.JWTConfig{
        SecretKey:              []byte(cfg.SecretKey),
        AccessTokenExpiry:      15 * time.Minute,
        RefreshTokenExpiry:     7 * 24 * time.Hour,
        EnableRefreshTokens:    true,
        EnableTokenBlacklist:   true,
    })

    orgPlugin := organizations.NewOrganizationPlugin()
    adminPlugin := admin.NewAdminPlugin()

    // Register plugins
    betterAuth.RegisterPlugin(jwtPlugin)
    betterAuth.RegisterPlugin(orgPlugin)
    betterAuth.RegisterPlugin(adminPlugin)

    // Enable plugins with configuration
    betterAuth.EnablePlugin(ctx, "jwt", map[string]interface{}{
        "enabled":              true,
        "secretKey":            cfg.SecretKey,
        "enableRefreshTokens":  true,
        "enableTokenBlacklist": true,
    })

    betterAuth.EnablePlugin(ctx, "organizations", map[string]interface{}{
        "enabled": true,
    })

    betterAuth.EnablePlugin(ctx, "admin", map[string]interface{}{
        "enabled": true,
    })

    // Start server
    log.Println("Server starting with plugins enabled")
    http.ListenAndServe(":8080", betterAuth)
}
```

## Plugin Routes

When plugins are enabled, they automatically register their routes:

### JWT Plugin Routes
- `POST /auth/jwt/refresh` - Refresh access token
- `POST /auth/jwt/revoke` - Revoke tokens
- `POST /auth/jwt/validate` - Validate token
- `GET /auth/jwt/info` - Get token information

### Organizations Plugin Routes
- `POST /auth/organizations` - Create organization
- `GET /auth/organizations/{id}` - Get organization
- `PUT /auth/organizations/{id}` - Update organization
- `DELETE /auth/organizations/{id}` - Delete organization
- `GET /auth/organizations/{id}/members` - List members
- `POST /auth/organizations/{id}/invite` - Invite user
- `POST /auth/organizations/{id}/teams` - Create team

### Admin Plugin Routes
- `POST /auth/admin/admins` - Create admin user
- `GET /auth/admin/admins` - List admin users
- `GET /auth/admin/audit-logs` - View audit logs
- `GET /auth/admin/system/stats` - System statistics
- `GET /auth/admin/system/metrics` - Performance metrics

## Database Schema

The plugin system uses GORM for database operations. Tables are only created when plugins are enabled:

### Core Tables (Always Present)
- `users` - User accounts
- `sessions` - User sessions  
- `oauth_accounts` - OAuth provider links

### Plugin Tables (Conditional)
- **Organizations Plugin**: `organizations`, `organization_members`, `teams`, `team_members`, `organization_invitations`
- **JWT Plugin**: `jwt_refresh_tokens`, `jwt_blacklist`
- **Admin Plugin**: `admin_users`, `admin_sessions`, `admin_audit_logs`, `system_metrics`

## Creating Custom Plugins

To create a custom plugin:

1. **Implement the Plugin interface**:
```go
type MyPlugin struct {
    *core.BasePlugin
    service *MyService
}

func NewMyPlugin() *MyPlugin {
    plugin := &MyPlugin{
        BasePlugin: core.NewBasePlugin("my-plugin"),
    }
    
    // Add models
    plugin.AddModel(&MyModel{})
    
    // Add services
    plugin.AddService(NewMyService())
    
    return plugin
}
```

2. **Define your models with GORM tags**:
```go
type MyModel struct {
    ID        string         `gorm:"primaryKey" json:"id"`
    Name      string         `gorm:"not null" json:"name"`
    CreatedAt time.Time      `json:"createdAt"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

3. **Implement plugin methods**:
```go
func (p *MyPlugin) Initialize(ctx context.Context, db *gorm.DB) error {
    p.service = NewMyService(db)
    p.setupRoutes()
    return p.BasePlugin.Initialize(ctx, db)
}

func (p *MyPlugin) setupRoutes() {
    p.AddRoute("GET /my-plugin/data", http.HandlerFunc(p.handleGetData))
}
```

4. **Register and enable the plugin**:
```go
myPlugin := NewMyPlugin()
auth.RegisterPlugin(myPlugin)
auth.EnablePlugin(ctx, "my-plugin", map[string]interface{}{
    "enabled": true,
})
```

## Benefits

1. **Conditional Loading**: Only load what you need
2. **Database Efficiency**: Tables only created for enabled plugins
3. **Modular Architecture**: Clean separation of concerns
4. **Extensibility**: Easy to add new functionality
5. **Configuration**: Flexible plugin configuration
6. **Lifecycle Management**: Proper plugin initialization and cleanup