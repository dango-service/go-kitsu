package kitsu

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMappingService_FindAnimeByExternalID(t *testing.T) {
	fixture, err := os.ReadFile("testdata/mapping_get_success.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mappings" {
			t.Errorf("Path = %q, want /mappings", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("filter[externalSite]") != "myanimelist/anime" {
			t.Errorf("filter[externalSite] = %q, want myanimelist/anime", q.Get("filter[externalSite]"))
		}
		if q.Get("filter[externalId]") != "1" {
			t.Errorf("filter[externalId] = %q, want 1", q.Get("filter[externalId]"))
		}
		if q.Get("include") != "item" {
			t.Errorf("include = %q, want item", q.Get("include"))
		}

		w.Header().Set("Content-Type", MediaTypeJSONAPI)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	kitsuID, resp, err := client.Mapping.FindAnimeByExternalID(context.Background(), "myanimelist/anime", "1")
	if err != nil {
		t.Fatalf("FindAnimeByExternalID() unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil Response")
	}
	if kitsuID != "1" {
		t.Errorf("kitsuID = %q, want 1", kitsuID)
	}
}

func TestMappingService_FindAnimeByExternalID_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", MediaTypeJSONAPI)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[],"meta":{"count":0}}`))
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	kitsuID, resp, err := client.Mapping.FindAnimeByExternalID(context.Background(), "myanimelist/anime", "9999999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("errors.Is(err, ErrNotFound) = false, want true: %v", err)
	}
	if kitsuID != "" {
		t.Errorf("expected empty kitsuID, got %q", kitsuID)
	}
	if resp == nil {
		t.Fatal("expected non-nil Response")
	}
}
