package core

import (
	"context"
	"net/http"

	"github.com/azizndao/grouter"
	"gorm.io/gorm"
)

type Plugin interface {
	Name() string
	Init(ctx context.Context, db *gorm.DB) error
	GetModels() []any
	RegisterRoutes(router grouter.RouteGroup) error
}

type BasePlugin struct {
	name    string
	handler http.Handler
}

func NewBasePlugin(name string) *BasePlugin {
	return &BasePlugin{
		name: name,
	}
}

func (p *BasePlugin) Name() string {
	return p.name
}

type TransportProvider interface {
	GetTransportInterface() any
}

type PluginRegistry struct {
	plugins           map[string]Plugin
	db                *gorm.DB
	transportProvider TransportProvider
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry(router grouter.Router, db *gorm.DB) *PluginRegistry {
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
