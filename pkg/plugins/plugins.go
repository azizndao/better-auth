package plugins

import (
	"net/http"
)

// Plugin interface defines a better-auth plugin
type Plugin interface {
	Name() string
	Initialize() error
	Handler() http.Handler
	Routes() map[string]http.Handler
}

// BasePlugin provides common plugin functionality
type BasePlugin struct {
	name    string
	routes  map[string]http.Handler
	handler http.Handler
}

// NewBasePlugin creates a new base plugin
func NewBasePlugin(name string) *BasePlugin {
	return &BasePlugin{
		name:   name,
		routes: make(map[string]http.Handler),
	}
}

// Name returns the plugin name
func (p *BasePlugin) Name() string {
	return p.name
}

// Initialize initializes the plugin
func (p *BasePlugin) Initialize() error {
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

// SetHandler sets the plugin's main handler
func (p *BasePlugin) SetHandler(handler http.Handler) {
	p.handler = handler
}

// AddRoute adds a route to the plugin
func (p *BasePlugin) AddRoute(path string, handler http.Handler) {
	p.routes[path] = handler
}

// AuthRouter interface to avoid import cycle
type AuthRouter interface {
	AddRoute(path string, handler http.Handler)
}

// PluginManager manages plugins
type PluginManager struct {
	plugins map[string]Plugin
	router  AuthRouter
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(router AuthRouter) *PluginManager {
	return &PluginManager{
		plugins: make(map[string]Plugin),
		router:  router,
	}
}

// Register registers a plugin
func (pm *PluginManager) Register(plugin Plugin) error {
	if err := plugin.Initialize(); err != nil {
		return err
	}

	pm.plugins[plugin.Name()] = plugin
	return nil
}

// Get returns a plugin by name
func (pm *PluginManager) Get(name string) (Plugin, bool) {
	plugin, exists := pm.plugins[name]
	return plugin, exists
}

// GetAll returns all registered plugins
func (pm *PluginManager) GetAll() map[string]Plugin {
	return pm.plugins
}

// ApplyRoutes applies all plugin routes to the auth system
func (pm *PluginManager) ApplyRoutes() {
	for _, plugin := range pm.plugins {
		for path, handler := range plugin.Routes() {
			pm.router.AddRoute(path, handler)
		}
	}
}

// Common plugin types

// TwoFactorPlugin provides 2FA functionality
type TwoFactorPlugin struct {
	*BasePlugin
}

// NewTwoFactorPlugin creates a new 2FA plugin
func NewTwoFactorPlugin() *TwoFactorPlugin {
	plugin := &TwoFactorPlugin{
		BasePlugin: NewBasePlugin("two-factor"),
	}

	plugin.AddRoute("/2fa/enable", http.HandlerFunc(plugin.handleEnable2FA))
	plugin.AddRoute("/2fa/disable", http.HandlerFunc(plugin.handleDisable2FA))
	plugin.AddRoute("/2fa/verify", http.HandlerFunc(plugin.handleVerify2FA))

	return plugin
}

func (p *TwoFactorPlugin) handleEnable2FA(w http.ResponseWriter, r *http.Request) {
	// Implementation for enabling 2FA
	w.WriteHeader(http.StatusNotImplemented)
}

func (p *TwoFactorPlugin) handleDisable2FA(w http.ResponseWriter, r *http.Request) {
	// Implementation for disabling 2FA
	w.WriteHeader(http.StatusNotImplemented)
}

func (p *TwoFactorPlugin) handleVerify2FA(w http.ResponseWriter, r *http.Request) {
	// Implementation for verifying 2FA
	w.WriteHeader(http.StatusNotImplemented)
}

// OAuthPlugin provides OAuth functionality
type OAuthPlugin struct {
	*BasePlugin
}

// NewOAuthPlugin creates a new OAuth plugin
func NewOAuthPlugin() *OAuthPlugin {
	plugin := &OAuthPlugin{
		BasePlugin: NewBasePlugin("oauth"),
	}

	plugin.AddRoute("/oauth/google", http.HandlerFunc(plugin.handleGoogleOAuth))
	plugin.AddRoute("/oauth/github", http.HandlerFunc(plugin.handleGitHubOAuth))
	plugin.AddRoute("/oauth/callback", http.HandlerFunc(plugin.handleOAuthCallback))

	return plugin
}

func (p *OAuthPlugin) handleGoogleOAuth(w http.ResponseWriter, r *http.Request) {
	// Implementation for Google OAuth
	w.WriteHeader(http.StatusNotImplemented)
}

func (p *OAuthPlugin) handleGitHubOAuth(w http.ResponseWriter, r *http.Request) {
	// Implementation for GitHub OAuth
	w.WriteHeader(http.StatusNotImplemented)
}

func (p *OAuthPlugin) handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	// Implementation for OAuth callback
	w.WriteHeader(http.StatusNotImplemented)
}

// WebhooksPlugin provides webhook functionality
type WebhooksPlugin struct {
	*BasePlugin
	handlers map[string][]func(data any)
}

// NewWebhooksPlugin creates a new webhooks plugin
func NewWebhooksPlugin() *WebhooksPlugin {
	plugin := &WebhooksPlugin{
		BasePlugin: NewBasePlugin("webhooks"),
		handlers:   make(map[string][]func(data any)),
	}

	return plugin
}

// OnUserCreated registers a handler for user creation events
func (p *WebhooksPlugin) OnUserCreated(handler func(data any)) {
	p.handlers["user.created"] = append(p.handlers["user.created"], handler)
}

// OnUserSignIn registers a handler for user sign-in events
func (p *WebhooksPlugin) OnUserSignIn(handler func(data any)) {
	p.handlers["user.signin"] = append(p.handlers["user.signin"], handler)
}

// Trigger triggers webhook handlers for an event
func (p *WebhooksPlugin) Trigger(event string, data any) {
	if handlers, exists := p.handlers[event]; exists {
		for _, handler := range handlers {
			go handler(data) // Run asynchronously
		}
	}
}

