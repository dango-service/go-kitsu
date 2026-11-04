package kitsu

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestAuthService_AuthorizeURL(t *testing.T) {
	client := NewClient(nil)

	urlStr := client.Auth.AuthorizeURL("client-123", "https://app.example.com/callback", "random-state")
	if !strings.HasPrefix(urlStr, DefaultAuthURL) {
		t.Errorf("AuthorizeURL does not start with DefaultAuthURL: %s", urlStr)
	}
	if !strings.Contains(urlStr, "client_id=client-123") {
		t.Errorf("AuthorizeURL missing client_id: %s", urlStr)
	}
	if !strings.Contains(urlStr, "response_type=code") {
		t.Errorf("AuthorizeURL missing response_type: %s", urlStr)
	}
	if !strings.Contains(urlStr, "redirect_uri=https%3A%2F%2Fapp.example.com%2Fcallback") {
		t.Errorf("AuthorizeURL missing redirect_uri: %s", urlStr)
	}
	if !strings.Contains(urlStr, "state=random-state") {
		t.Errorf("AuthorizeURL missing state: %s", urlStr)
	}
}

func TestAuthService_LoginPassword(t *testing.T) {
	fixture, err := os.ReadFile("testdata/token_success.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %q, want POST", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type = %q, want application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		}

		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}

		if r.Form.Get("grant_type") != "password" {
			t.Errorf("grant_type = %q, want password", r.Form.Get("grant_type"))
		}
		if r.Form.Get("username") != "testuser" {
			t.Errorf("username = %q, want testuser", r.Form.Get("username"))
		}
		if r.Form.Get("password") != "secretpass" {
			t.Errorf("password = %q, want secretpass", r.Form.Get("password"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewClient(nil, WithTokenURL(server.URL))

	token, err := client.Auth.LoginPassword(context.Background(), "", "", "testuser", "secretpass")
	if err != nil {
		t.Fatalf("LoginPassword() unexpected error: %v", err)
	}

	if token == nil {
		t.Fatal("expected non-nil Token")
	}
	if token.AccessToken != "mock-kitsu-access-token" {
		t.Errorf("token.AccessToken = %q, want mock-kitsu-access-token", token.AccessToken)
	}
	if token.RefreshToken != "mock-kitsu-refresh-token" {
		t.Errorf("token.RefreshToken = %q, want mock-kitsu-refresh-token", token.RefreshToken)
	}
	if token.Expiry.Before(time.Now()) {
		t.Errorf("token.Expiry %v is not in future", token.Expiry)
	}

	// Verify credential redaction
	tokenStr := token.String()
	if strings.Contains(tokenStr, "mock-kitsu-access-token") {
		t.Errorf("token.String() leaked access token: %s", tokenStr)
	}
	if !strings.Contains(tokenStr, "[REDACTED]") {
		t.Errorf("token.String() did not redact token: %s", tokenStr)
	}
}

func TestAuthService_ExchangeCode(t *testing.T) {
	fixture, err := os.ReadFile("testdata/token_success.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}
		if r.Form.Get("grant_type") != "authorization_code" {
			t.Errorf("grant_type = %q, want authorization_code", r.Form.Get("grant_type"))
		}
		if r.Form.Get("code") != "auth-code-123" {
			t.Errorf("code = %q, want auth-code-123", r.Form.Get("code"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewClient(nil, WithTokenURL(server.URL))

	token, err := client.Auth.ExchangeCode(context.Background(), "cid", "csec", "https://cb", "auth-code-123")
	if err != nil {
		t.Fatalf("ExchangeCode() unexpected error: %v", err)
	}
	if token == nil || token.AccessToken != "mock-kitsu-access-token" {
		t.Errorf("unexpected token: %+v", token)
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	fixture, err := os.ReadFile("testdata/token_success.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse form: %v", err)
		}
		if r.Form.Get("grant_type") != "refresh_token" {
			t.Errorf("grant_type = %q, want refresh_token", r.Form.Get("grant_type"))
		}
		if r.Form.Get("refresh_token") != "old-refresh-token" {
			t.Errorf("refresh_token = %q, want old-refresh-token", r.Form.Get("refresh_token"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewClient(nil, WithTokenURL(server.URL))

	token, err := client.Auth.RefreshToken(context.Background(), "cid", "csec", "old-refresh-token")
	if err != nil {
		t.Fatalf("RefreshToken() unexpected error: %v", err)
	}
	if token == nil || token.AccessToken != "mock-kitsu-access-token" {
		t.Errorf("unexpected token: %+v", token)
	}
}

func TestAuthService_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"Invalid username or password"}`))
	}))
	defer server.Close()

	client := NewClient(nil, WithTokenURL(server.URL))

	token, err := client.Auth.LoginPassword(context.Background(), "", "", "wrong", "creds")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Invalid username or password") {
		t.Errorf("error = %q, want error to contain description", err.Error())
	}
	if token != nil {
		t.Errorf("expected nil token, got %+v", token)
	}
}
