package http

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cjp2600/stepwise/internal/logger"
)

// Client represents an HTTP client for making requests
type Client struct {
	httpClient *http.Client
	logger     *logger.Logger
}

// Request represents an HTTP request
type Request struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    interface{}
	Query   map[string]string
	Timeout time.Duration
	Auth    *Auth
}

// Auth represents authentication configuration
type Auth struct {
	Type     string            `yaml:"type" json:"type"` // basic, bearer, oauth, api_key, custom
	Username string            `yaml:"username" json:"username"`
	Password string            `yaml:"password" json:"password"`
	Token    string            `yaml:"token" json:"token"`
	APIKey   string            `yaml:"api_key" json:"api_key"`
	APIKeyIn string            `yaml:"api_key_in" json:"api_key_in"` // header, query
	OAuth    *OAuthConfig      `yaml:"oauth" json:"oauth"`
	Custom   map[string]string `yaml:"custom" json:"custom"`
}

// OAuthConfig represents OAuth 2.0 configuration
type OAuthConfig struct {
	ClientID     string `yaml:"client_id" json:"client_id"`
	ClientSecret string `yaml:"client_secret" json:"client_secret"`
	TokenURL     string `yaml:"token_url" json:"token_url"`
	Scope        string `yaml:"scope" json:"scope"`
	GrantType    string `yaml:"grant_type" json:"grant_type"` // client_credentials, password
	Username     string `yaml:"username" json:"username"`
	Password     string `yaml:"password" json:"password"`
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
	Duration   time.Duration
	Error      error
}

// NewClient creates a new HTTP client
func NewClient(timeout time.Duration, log *logger.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: false, // Set to true for self-signed certificates
				},
			},
		},
		logger: log,
	}
}

// Execute performs an HTTP request
func (c *Client) Execute(req *Request) (*Response, error) {
	start := time.Now()

	// Build URL with query parameters
	finalURL := req.URL
	if len(req.Query) > 0 {
		parsedURL, err := url.Parse(req.URL)
		if err != nil {
			return nil, fmt.Errorf("invalid URL: %w", err)
		}

		query := parsedURL.Query()
		for key, value := range req.Query {
			query.Set(key, value)
		}
		parsedURL.RawQuery = query.Encode()
		finalURL = parsedURL.String()
	}

	// Create HTTP request
	httpReq, err := http.NewRequest(req.Method, finalURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Apply authentication
	if req.Auth != nil {
		if err := c.applyAuthentication(httpReq, req.Auth); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
	}

	// Set body if provided
	if req.Body != nil {
		bodyBytes, err := c.serializeBody(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize body: %w", err)
		}
		httpReq.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		httpReq.ContentLength = int64(len(bodyBytes))
	}

	// Log request
	authType := "none"
	if req.Auth != nil {
		authType = req.Auth.Type
	}
	c.logger.Debug("Making HTTP request",
		"method", req.Method,
		"url", finalURL,
		"headers", req.Headers,
		"auth_type", authType)

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// DEBUG: выводим тело ответа сразу после чтения
	fmt.Println("======= HTTP CLIENT RAW RESPONSE =======")
	fmt.Printf("[DEBUG] Body type: %T\n", body)
	fmt.Printf("[DEBUG] Body len: %d\n", len(body))
	fmt.Println(string(body))
	fmt.Printf("[DEBUG] Body hex: %x\n", body)
	fmt.Println("======= END HTTP CLIENT RAW RESPONSE =======")

	duration := time.Since(start)

	// Log response
	c.logger.Debug("Received HTTP response",
		"status", resp.StatusCode,
		"duration", duration,
		"body_size", len(body))

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
		Duration:   duration,
	}, nil
}

// applyAuthentication applies authentication to the request
func (c *Client) applyAuthentication(req *http.Request, auth *Auth) error {
	switch auth.Type {
	case "basic":
		return c.applyBasicAuth(req, auth)
	case "bearer":
		return c.applyBearerAuth(req, auth)
	case "oauth":
		return c.applyOAuthAuth(req, auth)
	case "api_key":
		return c.applyAPIKeyAuth(req, auth)
	case "custom":
		return c.applyCustomAuth(req, auth)
	default:
		return fmt.Errorf("unsupported authentication type: %s", auth.Type)
	}
}

// applyBasicAuth applies Basic Authentication
func (c *Client) applyBasicAuth(req *http.Request, auth *Auth) error {
	if auth.Username == "" || auth.Password == "" {
		return fmt.Errorf("username and password required for basic authentication")
	}

	credentials := auth.Username + ":" + auth.Password
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	req.Header.Set("Authorization", "Basic "+encoded)

	c.logger.Debug("Applied Basic Authentication", "username", auth.Username)
	return nil
}

// applyBearerAuth applies Bearer Token Authentication
func (c *Client) applyBearerAuth(req *http.Request, auth *Auth) error {
	if auth.Token == "" {
		return fmt.Errorf("token required for bearer authentication")
	}

	req.Header.Set("Authorization", "Bearer "+auth.Token)
	c.logger.Debug("Applied Bearer Authentication")
	return nil
}

// applyOAuthAuth applies OAuth 2.0 Authentication
func (c *Client) applyOAuthAuth(req *http.Request, auth *Auth) error {
	if auth.OAuth == nil {
		return fmt.Errorf("OAuth configuration required")
	}

	// For now, we'll support client credentials grant
	// In a full implementation, you'd want to handle token caching and refresh
	token, err := c.getOAuthToken(auth.OAuth)
	if err != nil {
		return fmt.Errorf("failed to get OAuth token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	c.logger.Debug("Applied OAuth Authentication")
	return nil
}

// getOAuthToken retrieves an OAuth token
func (c *Client) getOAuthToken(config *OAuthConfig) (string, error) {
	// This is a simplified OAuth implementation
	// In production, you'd want proper token management with caching and refresh

	data := url.Values{}
	data.Set("grant_type", config.GrantType)
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)

	if config.Scope != "" {
		data.Set("scope", config.Scope)
	}

	if config.GrantType == "password" {
		data.Set("username", config.Username)
		data.Set("password", config.Password)
	}

	req, err := http.NewRequest("POST", config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("OAuth token request failed: %s", string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse OAuth response: %w", err)
	}

	return tokenResp.AccessToken, nil
}

// applyAPIKeyAuth applies API Key Authentication
func (c *Client) applyAPIKeyAuth(req *http.Request, auth *Auth) error {
	if auth.APIKey == "" {
		return fmt.Errorf("API key required for API key authentication")
	}

	switch auth.APIKeyIn {
	case "header":
		req.Header.Set("X-API-Key", auth.APIKey)
	case "query":
		// API key is already handled in query parameters
		// This is just for documentation
	default:
		// Default to header
		req.Header.Set("X-API-Key", auth.APIKey)
	}

	c.logger.Debug("Applied API Key Authentication", "location", auth.APIKeyIn)
	return nil
}

// applyCustomAuth applies Custom Authentication
func (c *Client) applyCustomAuth(req *http.Request, auth *Auth) error {
	if auth.Custom == nil {
		return fmt.Errorf("custom authentication headers required")
	}

	for key, value := range auth.Custom {
		req.Header.Set(key, value)
	}

	c.logger.Debug("Applied Custom Authentication", "headers", len(auth.Custom))
	return nil
}

// serializeBody serializes the request body
func (c *Client) serializeBody(body interface{}) ([]byte, error) {
	switch v := body.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return json.Marshal(v)
	}
}

// GetJSONBody returns the response body as JSON
func (r *Response) GetJSONBody() (interface{}, error) {
	var result interface{}
	err := json.Unmarshal(r.Body, &result)
	return result, err
}

// GetTextBody returns the response body as text
func (r *Response) GetTextBody() string {
	return string(r.Body)
}

// IsSuccess returns true if the response status code indicates success
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// GetHeader returns a specific header value
func (r *Response) GetHeader(name string) string {
	if values, exists := r.Headers[name]; exists && len(values) > 0 {
		return values[0]
	}
	return ""
}
