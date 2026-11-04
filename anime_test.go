package kitsu

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestAnimeService_GetByID(t *testing.T) {
	fixture, err := os.ReadFile("testdata/anime_get_success.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/anime/1" {
			t.Errorf("Path = %q, want /anime/1", r.URL.Path)
		}
		w.Header().Set("Content-Type", MediaTypeJSONAPI)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	anime, resp, err := client.Anime.GetByID(context.Background(), "1")
	if err != nil {
		t.Fatalf("GetByID() unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil Response")
	}
	if anime == nil {
		t.Fatal("expected non-nil Anime")
	}
	if anime.ID != "1" {
		t.Errorf("anime.ID = %q, want 1", anime.ID)
	}
	if anime.CanonicalTitle != "Cowboy Bebop" {
		t.Errorf("anime.CanonicalTitle = %q, want Cowboy Bebop", anime.CanonicalTitle)
	}
	if anime.EpisodeCount == nil || *anime.EpisodeCount != 26 {
		t.Errorf("anime.EpisodeCount = %v, want 26", anime.EpisodeCount)
	}
	if anime.Subtype != AnimeSubtypeTV {
		t.Errorf("anime.Subtype = %q, want TV", anime.Subtype)
	}
	if anime.Status != AnimeStatusFinished {
		t.Errorf("anime.Status = %q, want finished", anime.Status)
	}
}

func TestAnimeService_GetByID_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", MediaTypeJSONAPI)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"status":"404","title":"Record not found"}]}`))
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	anime, resp, err := client.Anime.GetByID(context.Background(), "99999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("errors.Is(err, ErrNotFound) = false, want true: %v", err)
	}
	if anime != nil {
		t.Errorf("expected nil anime, got %+v", anime)
	}
	if resp == nil {
		t.Fatal("expected non-nil Response")
	}
}

func TestAnimeService_Search(t *testing.T) {
	fixture, err := os.ReadFile("testdata/anime_search_success.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/anime" {
			t.Errorf("Path = %q, want /anime", r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("filter[text]") != "Bebop" {
			t.Errorf("filter[text] = %q, want Bebop", query.Get("filter[text]"))
		}
		if query.Get("filter[subtype]") != "TV" {
			t.Errorf("filter[subtype] = %q, want TV", query.Get("filter[subtype]"))
		}
		if query.Get("page[limit]") != "5" {
			t.Errorf("page[limit] = %q, want 5", query.Get("page[limit]"))
		}

		w.Header().Set("Content-Type", MediaTypeJSONAPI)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	results, resp, err := client.Anime.Search(
		context.Background(),
		"Bebop",
		WithSubtype(AnimeSubtypeTV),
		WithLimit(5),
	)
	if err != nil {
		t.Fatalf("Search() unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil Response")
	}
	if len(results) != 2 {
		t.Fatalf("got %d anime results, want 2", len(results))
	}
	if results[0].CanonicalTitle != "Cowboy Bebop" {
		t.Errorf("results[0].CanonicalTitle = %q, want Cowboy Bebop", results[0].CanonicalTitle)
	}
	if results[1].CanonicalTitle != "Cowboy Bebop: Knockin' on Heaven's Door" {
		t.Errorf("results[1].CanonicalTitle = %q, want Cowboy Bebop: Knockin' on Heaven's Door", results[1].CanonicalTitle)
	}
}
