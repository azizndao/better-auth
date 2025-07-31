package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"better-auth/internal/config"
)

// OAuthService handles OAuth provider integrations
type OAuthService struct {
	providers map[string]*OAuthProvider
	client    *http.Client
}

// OAuthProvider represents an OAuth provider configuration
type OAuthProvider struct {
	Name         string
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Scopes       []string
	RedirectURL  string
}

// OAuthUserInfo represents user information from OAuth provider
type OAuthUserInfo struct {
	ID       string         `json:"id"`
	Email    string         `json:"email"`
	Name     string         `json:"name"`
	Picture  string         `json:"picture"`
	Provider string         `json:"provider"`
	Raw      map[string]any `json:"raw"`
}

// TokenResponse represents OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(cfg *config.Config) *OAuthService {
	service := &OAuthService{
		providers: make(map[string]*OAuthProvider),
		client:    &http.Client{Timeout: 30 * time.Second},
	}

	// Initialize OAuth providers from config
	for _, provider := range cfg.Providers {
		var authURL, tokenURL, userInfoURL string
		var defaultScopes []string

		switch provider.Name {
		case "google":
			authURL = "https://accounts.google.com/o/oauth2/v2/auth"
			tokenURL = "https://oauth2.googleapis.com/token"
			userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
			defaultScopes = []string{"openid", "email", "profile"}
		case "github":
			authURL = "https://github.com/login/oauth/authorize"
			tokenURL = "https://github.com/login/oauth/access_token"
			userInfoURL = "https://api.github.com/user"
			defaultScopes = []string{"user:email"}
		case "discord":
			authURL = "https://discord.com/api/oauth2/authorize"
			tokenURL = "https://discord.com/api/oauth2/token"
			userInfoURL = "https://discord.com/api/users/@me"
			defaultScopes = []string{"identify", "email"}
		default:
			continue
		}

		scopes := provider.Scopes
		if len(scopes) == 0 {
			scopes = defaultScopes
		}

		service.providers[provider.Name] = &OAuthProvider{
			Name:         provider.Name,
			ClientID:     provider.ClientID,
			ClientSecret: provider.ClientSecret,
			AuthURL:      authURL,
			TokenURL:     tokenURL,
			UserInfoURL:  userInfoURL,
			Scopes:       scopes,
			RedirectURL:  provider.RedirectURL,
		}
	}

	return service
}

// GetAuthURL generates the OAuth authorization URL
func (o *OAuthService) GetAuthURL(provider, state string) (string, error) {
	p, exists := o.providers[provider]
	if !exists {
		return "", fmt.Errorf("provider %s not configured", provider)
	}

	params := url.Values{
		"client_id":     {p.ClientID},
		"redirect_uri":  {p.RedirectURL},
		"response_type": {"code"},
		"scope":         {strings.Join(p.Scopes, " ")},
		"state":         {state},
	}

	authURL, err := url.Parse(p.AuthURL)
	if err != nil {
		return "", fmt.Errorf("invalid auth URL: %v", err)
	}

	authURL.RawQuery = params.Encode()
	return authURL.String(), nil
}

// ExchangeCode exchanges authorization code for access token
func (o *OAuthService) ExchangeCode(provider, code string) (*TokenResponse, error) {
	p, exists := o.providers[provider]
	if !exists {
		return nil, fmt.Errorf("provider %s not configured", provider)
	}

	data := url.Values{
		"client_id":     {p.ClientID},
		"client_secret": {p.ClientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {p.RedirectURL},
	}

	req, err := http.NewRequest("POST", p.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %v", err)
	}

	return &tokenResp, nil
}

// GetUserInfo retrieves user information using access token
func (o *OAuthService) GetUserInfo(provider, accessToken string) (*OAuthUserInfo, error) {
	p, exists := o.providers[provider]
	if !exists {
		return nil, fmt.Errorf("provider %s not configured", provider)
	}

	req, err := http.NewRequest("GET", p.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("user info request failed: %s", string(body))
	}

	var rawUserInfo map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&rawUserInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %v", err)
	}

	userInfo := &OAuthUserInfo{
		Provider: provider,
		Raw:      rawUserInfo,
	}

	// Parse provider-specific user info
	switch provider {
	case "google":
		userInfo.ID = getString(rawUserInfo, "id")
		userInfo.Email = getString(rawUserInfo, "email")
		userInfo.Name = getString(rawUserInfo, "name")
		userInfo.Picture = getString(rawUserInfo, "picture")

	case "github":
		userInfo.ID = fmt.Sprintf("%v", rawUserInfo["id"])
		userInfo.Name = getString(rawUserInfo, "name")
		userInfo.Picture = getString(rawUserInfo, "avatar_url")

		// GitHub requires separate API call for email
		if email, err := o.getGitHubEmail(accessToken); err == nil {
			userInfo.Email = email
		}

	case "discord":
		userInfo.ID = getString(rawUserInfo, "id")
		userInfo.Email = getString(rawUserInfo, "email")
		userInfo.Name = getString(rawUserInfo, "username")
		if avatar := getString(rawUserInfo, "avatar"); avatar != "" {
			userInfo.Picture = fmt.Sprintf(
				"https://cdn.discordapp.com/avatars/%s/%s.png",
				userInfo.ID,
				avatar,
			)
		}
	}

	return userInfo, nil
}

// GenerateState generates a random state parameter for OAuth
func (o *OAuthService) GenerateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// ValidateState validates the OAuth state parameter
func (o *OAuthService) ValidateState(expected, actual string) bool {
	return expected == actual
}

// getGitHubEmail retrieves primary email from GitHub API
func (o *OAuthService) getGitHubEmail(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, email := range emails {
		if primary, ok := email["primary"].(bool); ok && primary {
			if emailStr, ok := email["email"].(string); ok {
				return emailStr, nil
			}
		}
	}

	return "", fmt.Errorf("no primary email found")
}

// getString safely extracts string value from map
func getString(data map[string]any, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
