package admin

import (
	"context"
	"fmt"
	"net/http"

	"better-auth/pkg/plugins/core"
	"better-auth/pkg/transport"

	"gorm.io/gorm"
)

// AdminPlugin provides admin management functionality
type AdminPlugin struct {
	*core.BasePlugin
	service   *AdminService
	handlers  *AdminHandlers
	transport transport.Transport
}

// NewAdminPlugin creates a new admin plugin
func NewAdminPlugin() *AdminPlugin {
	plugin := &AdminPlugin{
		BasePlugin: core.NewBasePlugin("admin"),
	}

	// Add database models
	plugin.AddModel(&AdminUser{})
	plugin.AddModel(&AdminSession{})
	plugin.AddModel(&AdminAuditLog{})
	plugin.AddModel(&SystemMetrics{})

	// Set required configuration
	plugin.SetRequiredConfig(map[string]any{
		"enabled": true,
	})

	return plugin
}

// SetTransport sets the transport for the plugin
func (p *AdminPlugin) SetTransport(transportInterface any) {
	if t, ok := transportInterface.(transport.Transport); ok {
		p.transport = t
	}
}

// Initialize initializes the admin plugin
func (p *AdminPlugin) Initialize(ctx context.Context, db *gorm.DB) error {
	// Initialize service
	p.service = NewAdminService(db)
	
	// Initialize handlers with transport (use default if not set)
	if p.transport == nil {
		p.transport = transport.NewDefault()
	}
	p.handlers = NewAdminHandlers(p.service, p.transport)

	// Add service to plugin
	p.AddService(p.service)

	// Setup routes
	p.setupRoutes()

	// Call base initialization
	return p.BasePlugin.Initialize(ctx, db)
}

// setupRoutes configures the plugin routes
func (p *AdminPlugin) setupRoutes() {
	// Admin management routes
	p.AddRoute("POST /admin/admins", http.HandlerFunc(p.handlers.CreateAdmin))
	p.AddRoute("GET /admin/admins", http.HandlerFunc(p.handlers.ListAdmins))
	p.AddRoute("GET /admin/admins/{id}", http.HandlerFunc(p.handlers.GetAdmin))
	p.AddRoute("PUT /admin/admins/{id}", http.HandlerFunc(p.handlers.UpdateAdmin))
	p.AddRoute("DELETE /admin/admins/{id}", http.HandlerFunc(p.handlers.DeleteAdmin))

	// Audit and monitoring routes
	p.AddRoute("GET /admin/audit-logs", http.HandlerFunc(p.handlers.GetAuditLogs))
	p.AddRoute("GET /admin/system/stats", http.HandlerFunc(p.handlers.GetSystemStats))
	p.AddRoute("GET /admin/system/metrics", http.HandlerFunc(p.handlers.GetMetrics))
}

// ValidateConfig validates the plugin configuration
func (p *AdminPlugin) ValidateConfig(config map[string]any) error {
	enabled, ok := config["enabled"]
	if !ok {
		return fmt.Errorf("admin plugin requires 'enabled' configuration")
	}

	if enabledBool, ok := enabled.(bool); !ok || !enabledBool {
		return fmt.Errorf("admin plugin must be enabled")
	}

	return nil
}

// OnEnabled is called when the plugin is enabled
func (p *AdminPlugin) OnEnabled() error {
	return p.BasePlugin.OnEnabled()
}

// OnDisabled is called when the plugin is disabled
func (p *AdminPlugin) OnDisabled() error {
	return p.BasePlugin.OnDisabled()
}

// GetService returns the admin service
func (p *AdminPlugin) GetService() *AdminService {
	return p.service
}

// GetHandlers returns the admin handlers
func (p *AdminPlugin) GetHandlers() *AdminHandlers {
	return p.handlers
}

// Middleware provides admin-specific middleware
func (p *AdminPlugin) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is an admin route
		if len(r.URL.Path) >= 6 && r.URL.Path[:6] == "/admin" {
			// Ensure user is authenticated
			userCtx, ok := p.transport.GetUserContext(r)
			if !ok {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Check if user is an admin
			user := userCtx.User
			if user == nil {
				http.Error(w, "Invalid user context", http.StatusUnauthorized)
				return
			}

			// Verify admin permissions would be done here
			// For now, we'll just pass through since individual handlers check permissions
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAdminMiddleware creates middleware that requires admin privileges
func (p *AdminPlugin) RequireAdminMiddleware(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context
			userCtx, ok := p.transport.GetUserContext(r)
			if !ok {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Check admin permission
			canAccess, err := p.service.CheckPermission(r.Context(), userCtx.User.ID, permission)
			if err != nil || !canAccess {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

