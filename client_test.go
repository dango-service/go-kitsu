package kitsu

import (
	"net/http"
	"testing"
	"time"
)

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient(nil)
	if c == nil {
		t.Fatal("expected non-nil Client")
	}

	if c.baseURL != DefaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, DefaultBaseURL)
	}
	if c.tokenURL != DefaultTokenURL {
		t.Errorf("tokenURL = %q, want %q", c.tokenURL, DefaultTokenURL)
	}
	if c.userAgent != DefaultUserAgent {
		t.Errorf("userAgent = %q, want %q", c.userAgent, DefaultUserAgent)
	}
	if c.client.Timeout != DefaultTimeout {
		t.Errorf("client.Timeout = %v, want %v", c.client.Timeout, DefaultTimeout)
	}

	if c.Anime == nil {
		t.Error("AnimeService is nil")
	}
	if c.LibraryEntry == nil {
		t.Error("LibraryEntryService is nil")
	}
	if c.User == nil {
		t.Error("UserService is nil")
	}
	if c.Mapping == nil {
		t.Error("MappingService is nil")
	}
	if c.Auth == nil {
		t.Error("AuthService is nil")
	}
}

func TestNewClient_Options(t *testing.T) {
	customHTTP := &http.Client{Timeout: 10 * time.Second}
	c := NewClient(
		customHTTP,
		WithBaseURL("https://example.com/api/edge"),
		WithTokenURL("https://example.com/oauth/token"),
		WithToken(" secret-token "),
		WithUserAgent("custom-app/1.0.0"),
	)

	if c.baseURL != "https://example.com/api/edge" {
		t.Errorf("baseURL = %q, want https://example.com/api/edge", c.baseURL)
	}
	if c.tokenURL != "https://example.com/oauth/token" {
		t.Errorf("tokenURL = %q, want https://example.com/oauth/token", c.tokenURL)
	}
	if c.token != "secret-token" {
		t.Errorf("token = %q, want secret-token", c.token)
	}
	if c.userAgent != "custom-app/1.0.0" {
		t.Errorf("userAgent = %q, want custom-app/1.0.0", c.userAgent)
	}
	if c.client != customHTTP {
		t.Error("custom httpClient was not injected")
	}
}
