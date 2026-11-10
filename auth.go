package kitsu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// DefaultAuthURL is the standard Kitsu OAuth2 authorization endpoint.
	DefaultAuthURL = "https://kitsu.io/api/oauth/authorize"
)

// Token represents OAuth2 credentials returned by Kitsu's authorization server.
type Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	RefreshToken string    `json:"refresh_token"`
	Scope        string    `json:"scope"`
	CreatedAt    int64     `json:"created_at"`
	Expiry       time.Time `json:"expiry"`
}

// String redacts secret tokens to prevent accidental credential leakage in application logs.
func (t Token) String() string {
	return fmt.Sprintf("Token{AccessToken: [REDACTED], TokenType: %q, Expiry: %v}", t.TokenType, t.Expiry)
}

// GoString redacts secret tokens to prevent accidental credential leakage in format specifiers.
func (t Token) GoString() string {
	return t.String()
}

// AuthorizeURL returns the URL to direct users to for initiating the OAuth2 code grant flow.
// Providing a non-empty, cryptographically random state parameter is strongly recommended
// to prevent Cross-Site Request Forgery (CSRF).
func (s *AuthService) AuthorizeURL(clientID, redirectURI, state string) string {
	baseURL := DefaultAuthURL

	v := url.Values{}
	v.Set("client_id", clientID)
	v.Set("response_type", "code")
	if redirectURI != "" {
		v.Set("redirect_uri", redirectURI)
	}
	if state != "" {
		v.Set("state", state)
	}

	return fmt.Sprintf("%s?%s", baseURL, v.Encode())
}

// LoginPassword authenticates directly against Kitsu via the OAuth2 Resource Owner Password Grant.
func (s *AuthService) LoginPassword(ctx context.Context, clientID, clientSecret, username, password string) (*Token, error) {
	if ctx == nil {
		return nil, errors.New("kitsu: context must not be nil")
	}
	if username == "" || password == "" {
		return nil, errors.New("kitsu: username and password must not be empty")
	}

	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("username", username)
	form.Set("password", password)
	if clientID != "" {
		form.Set("client_id", clientID)
	}
	if clientSecret != "" {
		form.Set("client_secret", clientSecret)
	}

	return s.doTokenRequest(ctx, form)
}

// ExchangeCode exchanges an authorization code for an OAuth2 Token.
func (s *AuthService) ExchangeCode(ctx context.Context, clientID, clientSecret, redirectURI, code string) (*Token, error) {
	if ctx == nil {
		return nil, errors.New("kitsu: context must not be nil")
	}
	if code == "" {
		return nil, errors.New("kitsu: authorization code must not be empty")
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("code", code)
	if redirectURI != "" {
		form.Set("redirect_uri", redirectURI)
	}

	return s.doTokenRequest(ctx, form)
}

// RefreshToken requests a new access token using an existing refresh token.
func (s *AuthService) RefreshToken(ctx context.Context, clientID, clientSecret, refreshToken string) (*Token, error) {
	if ctx == nil {
		return nil, errors.New("kitsu: context must not be nil")
	}
	if refreshToken == "" {
		return nil, errors.New("kitsu: refresh token must not be empty")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	if clientID != "" {
		form.Set("client_id", clientID)
	}
	if clientSecret != "" {
		form.Set("client_secret", clientSecret)
	}

	return s.doTokenRequest(ctx, form)
}

func (s *AuthService) doTokenRequest(ctx context.Context, form url.Values) (*Token, error) {
	tokenURL := s.client.tokenURL
	if tokenURL == "" {
		tokenURL = DefaultTokenURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("kitsu: create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	if s.client.userAgent != "" {
		req.Header.Set("User-Agent", s.client.userAgent)
	}

	resp, err := s.client.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kitsu: token request failed: %w", err)
	}
	defer resp.Body.Close()

	limitedBody := io.LimitReader(resp.Body, maxResponseBodyBytes)

	if resp.StatusCode != http.StatusOK {
		var errData struct {
			Error       string `json:"error"`
			Description string `json:"error_description"`
		}
		_ = json.NewDecoder(limitedBody).Decode(&errData)
		msg := errData.Description
		if msg == "" {
			msg = errData.Error
		}
		if msg == "" {
			msg = fmt.Sprintf("status code %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("kitsu: token request rejected: %s", msg)
	}

	var token Token
	if err := json.NewDecoder(limitedBody).Decode(&token); err != nil {
		return nil, fmt.Errorf("kitsu: decode token response: %w", err)
	}

	if token.ExpiresIn > 0 {
		token.Expiry = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	}

	return &token, nil
}
