package core

import (
	"context"
	"errors"
	"net/http"

	"gorm.io/gorm"
)

// Plugin errors
var (
	ErrPluginNotFound       = errors.New("plugin not found")
	ErrPluginAlreadyEnabled = errors.New("plugin already enabled")
	ErrPluginNotEnabled     = errors.New("plugin not enabled")
	ErrInvalidConfig        = errors.New("invalid plugin configuration")
)

// DatabaseModel represents a GORM model that can be migrated
type DatabaseModel any

// PluginService represents a plugin service
type PluginService interface {
	Initialize(db *gorm.DB) error
	Name() string
}

// Plugin interface defines a better-auth plugin with enhanced capabilities
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
	RequiredConfig() map[string]any
	ValidateConfig(config map[string]any) error
}

// TransportAwarePlugin interface for plugins that need access to transport
type TransportAwarePlugin interface {
	Plugin
	SetTransport(transport any) // Using any to avoid import cycle
}

// BasePlugin provides common plugin functionality
type BasePlugin struct {
	name           string
	routes         map[string]http.Handler
	handler        http.Handler
	models         []DatabaseModel
	services       []PluginService
	requiredConfig map[string]any
	enabled        bool
}

// NewBasePlugin creates a new base plugin
func NewBasePlugin(name string) *BasePlugin {
	return &BasePlugin{
		name:           name,
		routes:         make(map[string]http.Handler),
		models:         make([]DatabaseModel, 0),
		services:       make([]PluginService, 0),
		requiredConfig: make(map[string]any),
		enabled:        false,
	}
}

// Name returns the plugin name
func (p *BasePlugin) Name() string {
	return p.name
}

// Initialize initializes the plugin
func (p *BasePlugin) Initialize(ctx context.Context, db *gorm.DB) error {
	p.enabled = true
	return nil
}

// Handler returns the plugin's main handler
func (p *BasePlugin) Handler() http.Handler {
	return p.handler
}

// Routes returns the plugin's routes
func (p *BasePlugin) Routes() map[string]http.Handler {
	return p.routes
}

// DatabaseModels returns the plugin's database models
func (p *BasePlugin) DatabaseModels() []DatabaseModel {
	return p.models
}

// Services returns the plugin's services
func (p *BasePlugin) Services() []PluginService {
	return p.services
}

// OnEnabled is called when the plugin is enabled
func (p *BasePlugin) OnEnabled() error {
	p.enabled = true
	return nil
}

// OnDisabled is called when the plugin is disabled
func (p *BasePlugin) OnDisabled() error {
	p.enabled = false
	return nil
}

// RequiredConfig returns the plugin's required configuration
func (p *BasePlugin) RequiredConfig() map[string]any {
	return p.requiredConfig
}

// ValidateConfig validates the plugin's configuration
func (p *BasePlugin) ValidateConfig(config map[string]any) error {
	return nil
}

// SetHandler sets the plugin's main handler
func (p *BasePlugin) SetHandler(handler http.Handler) {
	p.handler = handler
}

// AddRoute adds a route to the plugin
func (p *BasePlugin) AddRoute(path string, handler http.Handler) {
	p.routes[path] = handler
}

// AddModel adds a database model to the plugin
func (p *BasePlugin) AddModel(model DatabaseModel) {
	p.models = append(p.models, model)
}

// AddService adds a service to the plugin
func (p *BasePlugin) AddService(service PluginService) {
	p.services = append(p.services, service)
}

// SetRequiredConfig sets the required configuration for the plugin
func (p *BasePlugin) SetRequiredConfig(config map[string]any) {
	p.requiredConfig = config
}

// IsEnabled returns whether the plugin is enabled
func (p *BasePlugin) IsEnabled() bool {
	return p.enabled
}

// AuthRouter interface to avoid import cycle
type AuthRouter interface {
	AddRoute(path string, handler http.Handler)
}

// TransportProvider interface to avoid import cycle
type TransportProvider interface {
	GetTransportInterface() any // Using any to avoid import cycle
}

// PluginRegistry manages plugins with conditional loading
type PluginRegistry struct {
	plugins           map[string]Plugin
	enabledPlugins    map[string]Plugin
	appliedRoutes     map[string]bool // Track applied routes to avoid duplicates
	router            AuthRouter
	db                *gorm.DB
	serviceRegistry   *ServiceRegistry
	transportProvider TransportProvider
}

// ServiceRegistry manages plugin services
type ServiceRegistry struct {
	services map[string]PluginService
}

// NewServiceRegistry creates a new service registry
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]PluginService),
	}
}

// Register registers a service
func (sr *ServiceRegistry) Register(service PluginService) error {
	if err := service.Initialize(nil); err != nil {
		return err
	}
	sr.services[service.Name()] = service
	return nil
}

// Get returns a service by name
func (sr *ServiceRegistry) Get(name string) (PluginService, bool) {
	service, exists := sr.services[name]
	return service, exists
}

// GetAll returns all registered services
func (sr *ServiceRegistry) GetAll() map[string]PluginService {
	return sr.services
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry(router AuthRouter, db *gorm.DB) *PluginRegistry {
	return &PluginRegistry{
		plugins:         make(map[string]Plugin),
		enabledPlugins:  make(map[string]Plugin),
		appliedRoutes:   make(map[string]bool),
		router:          router,
		db:              db,
		serviceRegistry: NewServiceRegistry(),
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

// Enable enables a plugin and its dependencies
func (pr *PluginRegistry) Enable(
	ctx context.Context,
	pluginName string,
	config map[string]any,
) error {
	plugin, exists := pr.plugins[pluginName]
	if !exists {
		return ErrPluginNotFound
	}

	if _, enabled := pr.enabledPlugins[pluginName]; enabled {
		return ErrPluginAlreadyEnabled
	}

	if err := plugin.ValidateConfig(config); err != nil {
		return err
	}

	// Set transport for transport-aware plugins
	if transportAware, ok := plugin.(TransportAwarePlugin); ok && pr.transportProvider != nil {
		transportAware.SetTransport(pr.transportProvider.GetTransportInterface())
	}

	if err := plugin.Initialize(ctx, pr.db); err != nil {
		return err
	}

	if err := pr.migrateTables(plugin); err != nil {
		return err
	}

	if err := pr.initializeServices(plugin); err != nil {
		return err
	}

	if err := plugin.OnEnabled(); err != nil {
		return err
	}

	pr.enabledPlugins[pluginName] = plugin
	return nil
}

// Disable disables a plugin
func (pr *PluginRegistry) Disable(pluginName string) error {
	plugin, exists := pr.enabledPlugins[pluginName]
	if !exists {
		return ErrPluginNotEnabled
	}

	if err := plugin.OnDisabled(); err != nil {
		return err
	}

	delete(pr.enabledPlugins, pluginName)

	// Clear applied routes for the disabled plugin
	for path := range plugin.Routes() {
		delete(pr.appliedRoutes, path)
	}

	return nil
}

// migrateTables migrates database tables for a plugin
func (pr *PluginRegistry) migrateTables(plugin Plugin) error {
	models := plugin.DatabaseModels()
	if len(models) == 0 {
		return nil
	}

	// Convert []DatabaseModel to []interface{}
	interfaceModels := make([]any, len(models))
	for i, model := range models {
		interfaceModels[i] = model
	}

	return pr.db.AutoMigrate(interfaceModels...)
}

// initializeServices initializes services for a plugin
func (pr *PluginRegistry) initializeServices(plugin Plugin) error {
	services := plugin.Services()
	for _, service := range services {
		if err := service.Initialize(pr.db); err != nil {
			return err
		}
		pr.serviceRegistry.Register(service)
	}
	return nil
}

// Get returns a plugin by name
func (pr *PluginRegistry) Get(name string) (Plugin, bool) {
	plugin, exists := pr.plugins[name]
	return plugin, exists
}

// GetEnabled returns an enabled plugin by name
func (pr *PluginRegistry) GetEnabled(name string) (Plugin, bool) {
	plugin, exists := pr.enabledPlugins[name]
	return plugin, exists
}

// GetAllEnabled returns all enabled plugins
func (pr *PluginRegistry) GetAllEnabled() map[string]Plugin {
	return pr.enabledPlugins
}

// GetServiceRegistry returns the service registry
func (pr *PluginRegistry) GetServiceRegistry() *ServiceRegistry {
	return pr.serviceRegistry
}

// ApplyRoutes applies all enabled plugin routes to the auth system
func (pr *PluginRegistry) ApplyRoutes() {
	for _, plugin := range pr.enabledPlugins {
		for path, handler := range plugin.Routes() {
			// Skip routes that have already been applied
			if !pr.appliedRoutes[path] {
				pr.router.AddRoute(path, handler)
				pr.appliedRoutes[path] = true
			}
		}
	}
}

// IsPluginEnabled checks if a plugin is enabled
func (pr *PluginRegistry) IsPluginEnabled(name string) bool {
	_, exists := pr.enabledPlugins[name]
	return exists
}
