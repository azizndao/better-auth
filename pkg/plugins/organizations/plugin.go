package organizations

import (
	"context"
	"fmt"
	"net/http"

	"better-auth/pkg/plugins/core"
	"better-auth/pkg/transport"

	"gorm.io/gorm"
)

// OrganizationPlugin provides organization management functionality
type OrganizationPlugin struct {
	*core.BasePlugin
	service   *OrganizationService
	handlers  *OrganizationHandlers
	transport transport.Transport
}

// NewOrganizationPlugin creates a new organization plugin
func NewOrganizationPlugin() *OrganizationPlugin {
	plugin := &OrganizationPlugin{
		BasePlugin: core.NewBasePlugin("organizations"),
	}

	// Add database models
	plugin.AddModel(&Organization{})
	plugin.AddModel(&OrganizationMember{})
	plugin.AddModel(&Team{})
	plugin.AddModel(&TeamMember{})
	plugin.AddModel(&OrganizationInvitation{})

	// Set required configuration
	plugin.SetRequiredConfig(map[string]any{
		"enabled": true,
	})

	return plugin
}

// SetTransport sets the transport for the plugin
func (p *OrganizationPlugin) SetTransport(transportInterface any) {
	if t, ok := transportInterface.(transport.Transport); ok {
		p.transport = t
	}
}

// Initialize initializes the organization plugin
func (p *OrganizationPlugin) Initialize(ctx context.Context, db *gorm.DB) error {
	// Initialize service
	p.service = NewOrganizationService(db)
	
	// Initialize handlers with transport (use default if not set)
	if p.transport == nil {
		p.transport = transport.NewDefault()
	}
	p.handlers = NewOrganizationHandlers(p.service, p.transport)

	// Add service to plugin
	p.AddService(p.service)

	// Setup routes
	p.setupRoutes()

	// Call base initialization
	return p.BasePlugin.Initialize(ctx, db)
}

// setupRoutes configures the plugin routes
func (p *OrganizationPlugin) setupRoutes() {
	// Organization CRUD routes
	p.AddRoute("POST /organizations", http.HandlerFunc(p.handlers.CreateOrganization))
	p.AddRoute("GET /organizations/{id}", http.HandlerFunc(p.handlers.GetOrganization))
	p.AddRoute("PUT /organizations/{id}", http.HandlerFunc(p.handlers.UpdateOrganization))
	p.AddRoute("DELETE /organizations/{id}", http.HandlerFunc(p.handlers.DeleteOrganization))

	// Member management routes
	p.AddRoute(
		"GET /organizations/{id}/members",
		http.HandlerFunc(p.handlers.GetOrganizationMembers),
	)
	p.AddRoute("POST /organizations/{id}/invite", http.HandlerFunc(p.handlers.InviteToOrganization))
	p.AddRoute(
		"POST /organizations/accept-invitation",
		http.HandlerFunc(p.handlers.AcceptInvitation),
	)
	p.AddRoute(
		"PUT /organizations/{id}/members/{memberId}/role",
		http.HandlerFunc(p.handlers.UpdateMemberRole),
	)
	p.AddRoute(
		"DELETE /organizations/{id}/members/{memberId}",
		http.HandlerFunc(p.handlers.RemoveMember),
	)

	// Team management routes
	p.AddRoute("POST /organizations/{id}/teams", http.HandlerFunc(p.handlers.CreateTeam))
	p.AddRoute("GET /organizations/{id}/teams", http.HandlerFunc(p.handlers.GetOrganizationTeams))
}

// ValidateConfig validates the plugin configuration
func (p *OrganizationPlugin) ValidateConfig(config map[string]any) error {
	enabled, ok := config["enabled"]
	if !ok {
		return fmt.Errorf("organizations plugin requires 'enabled' configuration")
	}

	if enabledBool, ok := enabled.(bool); !ok || !enabledBool {
		return fmt.Errorf("organizations plugin must be enabled")
	}

	return nil
}

// OnEnabled is called when the plugin is enabled
func (p *OrganizationPlugin) OnEnabled() error {
	return p.BasePlugin.OnEnabled()
}

// OnDisabled is called when the plugin is disabled
func (p *OrganizationPlugin) OnDisabled() error {
	return p.BasePlugin.OnDisabled()
}

// GetService returns the organization service
func (p *OrganizationPlugin) GetService() *OrganizationService {
	return p.service
}

// GetHandlers returns the organization handlers
func (p *OrganizationPlugin) GetHandlers() *OrganizationHandlers {
	return p.handlers
}

// Middleware provides organization-specific middleware
func (p *OrganizationPlugin) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add organization context if needed
		// For example, check if user has access to organization
		next.ServeHTTP(w, r)
	})
}

