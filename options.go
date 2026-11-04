package kitsu

import (
	"net/url"
	"strings"
)

// ClientOption configures a Client.
type ClientOption func(*Client)

// WithBaseURL overrides the default JSON:API base URL.
// Useful for targeting test servers created with net/http/httptest.
func WithBaseURL(rawURL string) ClientOption {
	return func(c *Client) {
		trimmed := strings.TrimRight(rawURL, "/")
		if u, err := url.Parse(trimmed); err == nil && u.Scheme != "" && u.Host != "" {
			c.baseURL = trimmed
		} else {
			c.baseURL = rawURL
		}
	}
}

// WithTokenURL overrides the default OAuth2 token endpoint URL.
func WithTokenURL(tokenURL string) ClientOption {
	return func(c *Client) {
		c.tokenURL = tokenURL
	}
}

// WithToken sets a static OAuth2 Bearer token for authenticating requests.
func WithToken(token string) ClientOption {
	return func(c *Client) {
		c.token = strings.TrimSpace(token)
	}
}

// WithUserAgent sets a custom User-Agent header for API requests.
func WithUserAgent(ua string) ClientOption {
	return func(c *Client) {
		c.userAgent = strings.TrimSpace(ua)
	}
}
