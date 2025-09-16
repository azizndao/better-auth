package core

import (
	"better-auth/pkg/router"
	"context"
	"net/http"

	"gorm.io/gorm"
)

// Plugin interface defines a better-auth plugin with enhanced capabilities
type Plugin interface {
	Name() string
	Initialize(ctx context.Context, db *gorm.DB) error
	GetModels() []any
	RegisterRoutes(router router.RouteGroup) error
}

// BasePlugin provides common plugin functionality
type BasePlugin struct {
	name    string
	handler http.Handler
}

// NewBasePlugin creates a new base plugin
func NewBasePlugin(name string) *BasePlugin {
	return &BasePlugin{
		name: name,
	}
}

// Name returns the plugin name
func (p *BasePlugin) Name() string {
	return p.name
}

// TransportProvider interface to avoid import cycle
type TransportProvider interface {
	GetTransportInterface() any // Using any to avoid import cycle
}

// PluginRegistry manages plugins with conditional loading
type PluginRegistry struct {
	plugins           map[string]Plugin
	db                *gorm.DB
	transportProvider TransportProvider
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry(router router.Router, db *gorm.DB) *PluginRegistry {
	return &PluginRegistry{
		plugins: make(map[string]Plugin),
		db:      db,
	}
}

// SetTransportProvider sets the transport provider
func (pr *PluginRegistry) SetTransportProvider(provider TransportProvider) {
	pr.transportProvider = provider
}

// Register registers a plugin (but doesn't enable it)
func (pr *PluginRegistry) Register(plugin Plugin) error {
	pr.plugins[plugin.Name()] = plugin
	return nil
}

// Get returns a plugin by name
func (pr *PluginRegistry) Get(name string) (Plugin, bool) {
	plugin, exists := pr.plugins[name]
	return plugin, exists
}
