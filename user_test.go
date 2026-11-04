package kitsu

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestUserService_CurrentUser(t *testing.T) {
	fixture, err := os.ReadFile("testdata/user_self_success.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users" {
			t.Errorf("Path = %q, want /users", r.URL.Path)
		}
		if r.URL.Query().Get("filter[self]") != "true" {
			t.Errorf("filter[self] = %q, want true", r.URL.Query().Get("filter[self]"))
		}

		w.Header().Set("Content-Type", MediaTypeJSONAPI)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL), WithToken("test-token"))

	user, resp, err := client.User.CurrentUser(context.Background())
	if err != nil {
		t.Fatalf("CurrentUser() unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil Response")
	}
	if user == nil {
		t.Fatal("expected non-nil User")
	}
	if user.ID != "42" {
		t.Errorf("user.ID = %q, want 42", user.ID)
	}
	if user.Name != "SpikeSpiegel" {
		t.Errorf("user.Name = %q, want SpikeSpiegel", user.Name)
	}
	if user.Slug != "spikespiegel" {
		t.Errorf("user.Slug = %q, want spikespiegel", user.Slug)
	}
}

func TestUserService_CurrentUser_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", MediaTypeJSONAPI)
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"errors":[{"status":"401","title":"Unauthorized"}]}`))
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	user, resp, err := client.User.CurrentUser(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("errors.Is(err, ErrUnauthorized) = false, want true: %v", err)
	}
	if user != nil {
		t.Errorf("expected nil user, got %+v", user)
	}
	if resp == nil {
		t.Fatal("expected non-nil Response")
	}
}

func TestUserService_GetByID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/42" {
			t.Errorf("Path = %q, want /users/42", r.URL.Path)
		}

		w.Header().Set("Content-Type", MediaTypeJSONAPI)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"data": {
				"id": "42",
				"type": "users",
				"attributes": {
					"name": "SpikeSpiegel",
					"slug": "spikespiegel"
				}
			}
		}`))
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	user, resp, err := client.User.GetByID(context.Background(), "42")
	if err != nil {
		t.Fatalf("GetByID() unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil Response")
	}
	if user == nil || user.ID != "42" || user.Name != "SpikeSpiegel" {
		t.Errorf("unexpected user: %+v", user)
	}
}
