package kitsu

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMangaService_GetByID(t *testing.T) {
	fixture := []byte(`{
		"data": {
			"id": "1",
			"type": "manga",
			"attributes": {
				"slug": "berserk",
				"synopsis": "Guts, a former mercenary...",
				"canonicalTitle": "Berserk",
				"subtype": "manga",
				"status": "current",
				"chapterCount": 380,
				"volumeCount": 42
			}
		}
	}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/manga/1" {
			t.Errorf("Path = %q, want /manga/1", r.URL.Path)
		}
		w.Header().Set("Content-Type", MediaTypeJSONAPI)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	manga, resp, err := client.Manga.GetByID(context.Background(), "1")
	if err != nil {
		t.Fatalf("GetByID() unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil Response")
	}
	if manga == nil {
		t.Fatal("expected non-nil Manga")
	}
	if manga.ID != "1" {
		t.Errorf("manga.ID = %q, want 1", manga.ID)
	}
	if manga.CanonicalTitle != "Berserk" {
		t.Errorf("manga.CanonicalTitle = %q, want Berserk", manga.CanonicalTitle)
	}
	if manga.ChapterCount == nil || *manga.ChapterCount != 380 {
		t.Errorf("manga.ChapterCount = %v, want 380", manga.ChapterCount)
	}
	if manga.VolumeCount == nil || *manga.VolumeCount != 42 {
		t.Errorf("manga.VolumeCount = %v, want 42", manga.VolumeCount)
	}
	if manga.Subtype != MangaSubtypeManga {
		t.Errorf("manga.Subtype = %q, want manga", manga.Subtype)
	}
	if manga.Status != MangaStatusCurrent {
		t.Errorf("manga.Status = %q, want current", manga.Status)
	}
}

func TestMangaService_GetByID_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", MediaTypeJSONAPI)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"status":"404","title":"Record not found"}]}`))
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	manga, resp, err := client.Manga.GetByID(context.Background(), "99999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("errors.Is(err, ErrNotFound) = false, want true: %v", err)
	}
	if manga != nil {
		t.Errorf("expected nil manga, got %+v", manga)
	}
	if resp == nil {
		t.Fatal("expected non-nil Response")
	}
}

func TestMangaService_Search(t *testing.T) {
	fixture := []byte(`{
		"data": [
			{
				"id": "1",
				"type": "manga",
				"attributes": {
					"canonicalTitle": "Berserk",
					"subtype": "manga",
					"status": "current"
				}
			}
		]
	}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/manga" {
			t.Errorf("Path = %q, want /manga", r.URL.Path)
		}
		if q := r.URL.Query().Get("filter[text]"); q != "Berserk" {
			t.Errorf("filter[text] = %q, want Berserk", q)
		}
		if l := r.URL.Query().Get("page[limit]"); l != "5" {
			t.Errorf("page[limit] = %q, want 5", l)
		}
		if s := r.URL.Query().Get("filter[subtype]"); s != "manga" {
			t.Errorf("filter[subtype] = %q, want manga", s)
		}
		w.Header().Set("Content-Type", MediaTypeJSONAPI)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	results, resp, err := client.Manga.Search(context.Background(), "Berserk",
		WithMangaLimit(5),
		WithMangaSubtype(MangaSubtypeManga),
	)
	if err != nil {
		t.Fatalf("Search() unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil Response")
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].CanonicalTitle != "Berserk" {
		t.Errorf("results[0].CanonicalTitle = %q, want Berserk", results[0].CanonicalTitle)
	}
}
