package kitsu

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestLibraryEntryService_GetEntry(t *testing.T) {
	fixture, err := os.ReadFile("testdata/library_entry_get_success.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/library-entries" {
			t.Errorf("Path = %q, want /library-entries", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("filter[userId]") != "42" {
			t.Errorf("filter[userId] = %q, want 42", q.Get("filter[userId]"))
		}
		if q.Get("filter[animeId]") != "1" {
			t.Errorf("filter[animeId] = %q, want 1", q.Get("filter[animeId]"))
		}

		w.Header().Set("Content-Type", MediaTypeJSONAPI)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	entry, resp, err := client.LibraryEntry.GetEntry(context.Background(), "42", "1")
	if err != nil {
		t.Fatalf("GetEntry() unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil Response")
	}
	if entry == nil {
		t.Fatal("expected non-nil LibraryEntry")
	}
	if entry.ID != "12345" {
		t.Errorf("entry.ID = %q, want 12345", entry.ID)
	}
	if entry.Status != LibraryStatusCurrent {
		t.Errorf("entry.Status = %q, want current", entry.Status)
	}
	if entry.Progress != 12 {
		t.Errorf("entry.Progress = %d, want 12", entry.Progress)
	}
	if entry.AnimeID != "1" {
		t.Errorf("entry.AnimeID = %q, want 1", entry.AnimeID)
	}
	if entry.Anime == nil || entry.Anime.CanonicalTitle != "Cowboy Bebop" {
		t.Errorf("entry.Anime = %+v, want Cowboy Bebop", entry.Anime)
	}
}

func TestLibraryEntryService_GetEntry_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", MediaTypeJSONAPI)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[],"meta":{"count":0}}`))
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	entry, resp, err := client.LibraryEntry.GetEntry(context.Background(), "42", "99999")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("errors.Is(err, ErrNotFound) = false, want true: %v", err)
	}
	if entry != nil {
		t.Errorf("expected nil entry, got %+v", entry)
	}
	if resp == nil {
		t.Fatal("expected non-nil Response")
	}
}

func TestLibraryEntryService_GetCollection(t *testing.T) {
	fixture, err := os.ReadFile("testdata/library_entry_collection_success.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/library-entries" {
			t.Errorf("Path = %q, want /library-entries", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("filter[userId]") != "42" {
			t.Errorf("filter[userId] = %q, want 42", q.Get("filter[userId]"))
		}
		if q.Get("filter[status]") != "current" {
			t.Errorf("filter[status] = %q, want current", q.Get("filter[status]"))
		}

		w.Header().Set("Content-Type", MediaTypeJSONAPI)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	entries, resp, err := client.LibraryEntry.GetCollection(context.Background(), "42", WithStatus(LibraryStatusCurrent))
	if err != nil {
		t.Fatalf("GetCollection() unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil Response")
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
	if entries[0].ID != "12345" || entries[1].ID != "12346" {
		t.Errorf("entry IDs mismatch: %+v", entries)
	}
}

func TestLibraryEntryService_SaveEntry_Patch(t *testing.T) {
	fixture, err := os.ReadFile("testdata/library_entry_save_success.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("Method = %q, want PATCH", r.Method)
		}
		if r.URL.Path != "/library-entries/12345" {
			t.Errorf("Path = %q, want /library-entries/12345", r.URL.Path)
		}

		var payload struct {
			Data struct {
				ID         string `json:"id"`
				Type       string `json:"type"`
				Attributes struct {
					Progress int `json:"progress"`
				} `json:"attributes"`
			} `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if payload.Data.ID != "12345" {
			t.Errorf("payload id = %q, want 12345", payload.Data.ID)
		}
		if payload.Data.Attributes.Progress != 13 {
			t.Errorf("payload progress = %d, want 13", payload.Data.Attributes.Progress)
		}

		w.Header().Set("Content-Type", MediaTypeJSONAPI)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	prog := 13
	entry, resp, err := client.LibraryEntry.SaveEntry(context.Background(), SaveLibraryEntryInput{
		ID:       "12345",
		AnimeID:  "1",
		UserID:   "42",
		Progress: &prog,
	})
	if err != nil {
		t.Fatalf("SaveEntry() unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil Response")
	}
	if entry.Progress != 13 {
		t.Errorf("entry.Progress = %d, want 13", entry.Progress)
	}
}

func TestLibraryEntryService_SaveEntry_Post(t *testing.T) {
	fixture, err := os.ReadFile("testdata/library_entry_save_success.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// First lookup during upsert -> return not found
			w.Header().Set("Content-Type", MediaTypeJSONAPI)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"data":[],"meta":{"count":0}}`))
			return
		}

		if r.Method == http.MethodPost {
			if r.URL.Path != "/library-entries" {
				t.Errorf("Path = %q, want /library-entries", r.URL.Path)
			}
			w.Header().Set("Content-Type", MediaTypeJSONAPI)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write(fixture)
			return
		}

		t.Errorf("unexpected method: %s", r.Method)
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	status := LibraryStatusCurrent
	prog := 13
	entry, resp, err := client.LibraryEntry.SaveEntry(context.Background(), SaveLibraryEntryInput{
		AnimeID:  "1",
		UserID:   "42",
		Status:   &status,
		Progress: &prog,
	})
	if err != nil {
		t.Fatalf("SaveEntry() unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("expected non-nil Response")
	}
	if entry.ID != "12345" {
		t.Errorf("entry.ID = %q, want 12345", entry.ID)
	}
}

func TestLibraryEntryService_DeleteEntry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Method = %q, want DELETE", r.Method)
		}
		if r.URL.Path != "/library-entries/12345" {
			t.Errorf("Path = %q, want /library-entries/12345", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(nil, WithBaseURL(server.URL))

	resp, err := client.LibraryEntry.DeleteEntry(context.Background(), "12345")
	if err != nil {
		t.Fatalf("DeleteEntry() unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status code = %d, want 204", resp.StatusCode)
	}
}
