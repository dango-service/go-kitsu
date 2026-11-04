package kitsu

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTransport_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Method = %q, want GET", r.Method)
		}
		if r.Header.Get("Accept") != MediaTypeJSONAPI {
			t.Errorf("Accept = %q, want %q", r.Header.Get("Accept"), MediaTypeJSONAPI)
		}
		if r.Header.Get("Authorization") != "Bearer test-kitsu-token" {
			t.Errorf("Authorization = %q, want Bearer test-kitsu-token", r.Header.Get("Authorization"))
		}
		if r.Header.Get("User-Agent") != "test-agent" {
			t.Errorf("User-Agent = %q, want test-agent", r.Header.Get("User-Agent"))
		}

		w.Header().Set("Content-Type", MediaTypeJSONAPI)
		w.Header().Set("X-RateLimit-Limit", "60")
		w.Header().Set("X-RateLimit-Remaining", "59")
		w.Header().Set("X-RateLimit-Reset", "1700000000")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"id":"1","type":"anime","attributes":{"canonicalTitle":"Cowboy Bebop"}}}`))
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL), WithToken("test-kitsu-token"), WithUserAgent("test-agent"))

	doc, resp, err := client.doRequest(context.Background(), http.MethodGet, "/anime/1", nil)
	if err != nil {
		t.Fatalf("doRequest() unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil Response")
	}
	if resp.RateLimit.Limit != 60 {
		t.Errorf("RateLimit.Limit = %d, want 60", resp.RateLimit.Limit)
	}
	if resp.RateLimit.Remaining != 59 {
		t.Errorf("RateLimit.Remaining = %d, want 59", resp.RateLimit.Remaining)
	}
	if resp.RateLimit.Reset.Unix() != 1700000000 {
		t.Errorf("RateLimit.Reset = %v, want unix 1700000000", resp.RateLimit.Reset)
	}
	if doc == nil || len(doc.Data) == 0 {
		t.Fatal("expected non-empty doc.Data")
	}
}

func TestTransport_JSONAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", MediaTypeJSONAPI)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errors":[{"status":"400","title":"Invalid query parameter","detail":"filter[invalid] not allowed"}]}`))
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	_, _, err := client.doRequest(context.Background(), http.MethodGet, "/anime", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var errResp *ErrorResponse
	if !errors.As(err, &errResp) {
		t.Fatalf("expected *ErrorResponse, got %T: %v", err, err)
	}
	if len(errResp.Errors) != 1 {
		t.Fatalf("got %d errors, want 1", len(errResp.Errors))
	}
	if errResp.Errors[0].Title != "Invalid query parameter" {
		t.Errorf("error title = %q, want Invalid query parameter", errResp.Errors[0].Title)
	}
}

func TestTransport_RateLimitError_Seconds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", MediaTypeJSONAPI)
		w.Header().Set("Retry-After", "45")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"errors":[{"status":"429","title":"Rate limit exceeded"}]}`))
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	_, resp, err := client.doRequest(context.Background(), http.MethodGet, "/anime", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("errors.Is(err, ErrRateLimited) = false, want true")
	}
	if resp == nil || resp.RateLimit.RetryAfter != 45*time.Second {
		t.Errorf("RetryAfter = %v, want 45s", resp.RateLimit.RetryAfter)
	}
}

func TestTransport_RateLimitError_HTTPDate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		future := time.Now().Add(60 * time.Second).UTC().Format(http.TimeFormat)
		w.Header().Set("Content-Type", MediaTypeJSONAPI)
		w.Header().Set("Retry-After", future)
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"errors":[{"status":"429","title":"Rate limit exceeded"}]}`))
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	_, resp, err := client.doRequest(context.Background(), http.MethodGet, "/anime", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("errors.Is(err, ErrRateLimited) = false, want true")
	}
	if resp == nil || resp.RateLimit.RetryAfter <= 0 {
		t.Errorf("RetryAfter = %v, want > 0", resp.RateLimit.RetryAfter)
	}
}

func TestTransport_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", MediaTypeJSONAPI)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"status":"404","title":"Record not found","detail":"The record does not exist"}]}`))
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	_, _, err := client.doRequest(context.Background(), http.MethodGet, "/anime/99999999", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrNotFound) {
		t.Errorf("errors.Is(err, ErrNotFound) = false, want true")
	}
}

func TestTransport_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", MediaTypeJSONAPI)
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errors":[{"status":"401","title":"Unauthorized"}]}`))
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	_, _, err := client.doRequest(context.Background(), http.MethodGet, "/users", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("errors.Is(err, ErrUnauthorized) = false, want true")
	}
}

func TestTransport_NoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	doc, resp, err := client.doRequest(context.Background(), http.MethodDelete, "/library-entries/123", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc != nil {
		t.Errorf("expected nil doc on 204 No Content, got %+v", doc)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status code = %d, want 204", resp.StatusCode)
	}
}

func TestTransport_ContextCancellation(t *testing.T) {
	client := NewClient(nil)

	// Test nil context
	_, _, err := client.doRequest(nil, http.MethodGet, "/anime", nil) //nolint:staticcheck
	if err == nil {
		t.Error("expected error for nil context, got nil")
	}

	// Test pre-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err = client.doRequest(ctx, http.MethodGet, "/anime", nil)
	if err == nil {
		t.Error("expected error for cancelled context, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}
