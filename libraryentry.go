package kitsu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// LibraryStatus represents the tracking status of a media in the user's library.
type LibraryStatus string

const (
	LibraryStatusCurrent   LibraryStatus = "current"
	LibraryStatusPlanned   LibraryStatus = "planned"
	LibraryStatusCompleted LibraryStatus = "completed"
	LibraryStatusOnHold    LibraryStatus = "on_hold"
	LibraryStatusDropped   LibraryStatus = "dropped"
)

// LibraryEntry represents a user's tracking entry for an anime or manga in Kitsu.
type LibraryEntry struct {
	ID             string        `json:"id"`
	Status         LibraryStatus `json:"status"`
	Progress       int           `json:"progress"`
	RatingTwenty   *int          `json:"ratingTwenty,omitempty"`
	Reconsuming    bool          `json:"reconsuming"`
	ReconsumeCount int           `json:"reconsumeCount"`
	UpdatedAt      time.Time     `json:"updatedAt"`
	AnimeID        string        `json:"animeId,omitempty"`
	MangaID        string        `json:"mangaId,omitempty"`
	UserID         string        `json:"userId,omitempty"`
	Anime          *Anime        `json:"anime,omitempty"`
	Manga          *Manga        `json:"manga,omitempty"`
}

// SaveLibraryEntryInput specifies parameters to create or update a library entry.
type SaveLibraryEntryInput struct {
	ID           string         `json:"id,omitempty"`
	AnimeID      string         `json:"animeId,omitempty"`
	MangaID      string         `json:"mangaId,omitempty"`
	UserID       string         `json:"userId"`
	Status       *LibraryStatus `json:"status,omitempty"`
	Progress     *int           `json:"progress,omitempty"`
	RatingTwenty *int           `json:"ratingTwenty,omitempty"`
	Reconsuming  *bool          `json:"reconsuming,omitempty"`
}

func unmarshalLibraryEntry(res jsonAPIResource, included []jsonAPIResource) (*LibraryEntry, error) {
	var entry LibraryEntry
	if len(res.Attributes) > 0 {
		if err := json.Unmarshal(res.Attributes, &entry); err != nil {
			return nil, fmt.Errorf("kitsu: unmarshal library entry attributes: %w", err)
		}
	}
	entry.ID = res.ID

	if animeRel, ok := res.Relationships["anime"]; ok && len(animeRel.Data) > 0 {
		var ref struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		}
		if err := json.Unmarshal(animeRel.Data, &ref); err == nil {
			entry.AnimeID = ref.ID
		}
	}

	if mangaRel, ok := res.Relationships["manga"]; ok && len(mangaRel.Data) > 0 {
		var ref struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		}
		if err := json.Unmarshal(mangaRel.Data, &ref); err == nil {
			entry.MangaID = ref.ID
		}
	}

	if userRel, ok := res.Relationships["user"]; ok && len(userRel.Data) > 0 {
		var ref struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		}
		if err := json.Unmarshal(userRel.Data, &ref); err == nil {
			entry.UserID = ref.ID
		}
	}

	if entry.AnimeID != "" && len(included) > 0 {
		for _, inc := range included {
			if inc.Type == "anime" && inc.ID == entry.AnimeID {
				anime, err := unmarshalAnime(inc)
				if err == nil {
					entry.Anime = anime
				}
				break
			}
		}
	}

	if entry.MangaID != "" && len(included) > 0 {
		for _, inc := range included {
			if inc.Type == "manga" && inc.ID == entry.MangaID {
				manga, err := unmarshalManga(inc)
				if err == nil {
					entry.Manga = manga
				}
				break
			}
		}
	}

	return &entry, nil
}

// GetEntry retrieves a user's library entry for a specific anime.
func (s *LibraryEntryService) GetEntry(ctx context.Context, userID, animeID string) (*LibraryEntry, *Response, error) {
	if userID == "" || animeID == "" {
		return nil, nil, errors.New("kitsu: userID and animeID must not be empty")
	}

	v := url.Values{}
	v.Set("filter[userId]", userID)
	v.Set("filter[animeId]", animeID)
	v.Set("include", "anime")

	path := "/library-entries?" + v.Encode()
	doc, resp, err := s.client.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, resp, fmt.Errorf("kitsu: get entry (user %s, anime %s): %w", userID, animeID, err)
	}

	if doc == nil || len(doc.Data) == 0 {
		return nil, resp, fmt.Errorf("kitsu: get entry (user %s, anime %s): %w", userID, animeID, ErrNotFound)
	}

	var resources []jsonAPIResource
	if err := json.Unmarshal(doc.Data, &resources); err != nil {
		return nil, resp, fmt.Errorf("kitsu: unmarshal library entry resources: %w", err)
	}

	if len(resources) == 0 {
		return nil, resp, fmt.Errorf("kitsu: get entry (user %s, anime %s): %w", userID, animeID, ErrNotFound)
	}

	entry, err := unmarshalLibraryEntry(resources[0], doc.Included)
	if err != nil {
		return nil, resp, err
	}

	return entry, resp, nil
}

// GetEntryByManga retrieves a user's library entry for a specific manga or comic.
func (s *LibraryEntryService) GetEntryByManga(ctx context.Context, userID, mangaID string) (*LibraryEntry, *Response, error) {
	if userID == "" || mangaID == "" {
		return nil, nil, errors.New("kitsu: userID and mangaID must not be empty")
	}

	v := url.Values{}
	v.Set("filter[userId]", userID)
	v.Set("filter[mangaId]", mangaID)
	v.Set("include", "manga")

	path := "/library-entries?" + v.Encode()
	doc, resp, err := s.client.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, resp, fmt.Errorf("kitsu: get manga entry (user %s, manga %s): %w", userID, mangaID, err)
	}

	if doc == nil || len(doc.Data) == 0 {
		return nil, resp, fmt.Errorf("kitsu: get manga entry (user %s, manga %s): %w", userID, mangaID, ErrNotFound)
	}

	var resources []jsonAPIResource
	if err := json.Unmarshal(doc.Data, &resources); err != nil {
		return nil, resp, fmt.Errorf("kitsu: unmarshal library entry resources: %w", err)
	}

	if len(resources) == 0 {
		return nil, resp, fmt.Errorf("kitsu: get manga entry (user %s, manga %s): %w", userID, mangaID, ErrNotFound)
	}

	entry, err := unmarshalLibraryEntry(resources[0], doc.Included)
	if err != nil {
		return nil, resp, err
	}

	return entry, resp, nil
}

// GetCollection retrieves a list of library entries for a user, optionally filtered by status.
func (s *LibraryEntryService) GetCollection(ctx context.Context, userID string, opts ...CollectionOption) ([]LibraryEntry, *Response, error) {
	if userID == "" {
		return nil, nil, errors.New("kitsu: userID must not be empty")
	}

	params := collectionParams{}
	for _, opt := range opts {
		opt(&params)
	}

	v := url.Values{}
	v.Set("filter[userId]", userID)
	v.Set("include", "anime")
	if params.status != "" {
		v.Set("filter[status]", string(params.status))
	}

	path := "/library-entries?" + v.Encode()
	doc, resp, err := s.client.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, resp, fmt.Errorf("kitsu: get collection for user %s: %w", userID, err)
	}

	if doc == nil || len(doc.Data) == 0 {
		return []LibraryEntry{}, resp, nil
	}

	var resources []jsonAPIResource
	if err := json.Unmarshal(doc.Data, &resources); err != nil {
		return nil, resp, fmt.Errorf("kitsu: unmarshal collection resources: %w", err)
	}

	entries := make([]LibraryEntry, 0, len(resources))
	for _, res := range resources {
		entry, err := unmarshalLibraryEntry(res, doc.Included)
		if err != nil {
			return nil, resp, err
		}
		entries = append(entries, *entry)
	}

	return entries, resp, nil
}

// SaveEntry creates or updates an anime or manga entry in the user's library.
func (s *LibraryEntryService) SaveEntry(ctx context.Context, input SaveLibraryEntryInput) (*LibraryEntry, *Response, error) {
	targetID := input.ID

	// If entry ID is omitted, check if an existing entry exists to update idempotently
	if targetID == "" {
		if input.MangaID != "" {
			existing, _, err := s.GetEntryByManga(ctx, input.UserID, input.MangaID)
			if err == nil && existing != nil {
				targetID = existing.ID
			}
		} else if input.AnimeID != "" {
			existing, _, err := s.GetEntry(ctx, input.UserID, input.AnimeID)
			if err == nil && existing != nil {
				targetID = existing.ID
			}
		}
	}

	attributes := make(map[string]any)
	if input.Status != nil {
		attributes["status"] = *input.Status
	}
	if input.Progress != nil {
		attributes["progress"] = *input.Progress
	}
	if input.RatingTwenty != nil {
		attributes["ratingTwenty"] = *input.RatingTwenty
	}
	if input.Reconsuming != nil {
		attributes["reconsuming"] = *input.Reconsuming
	}

	var method, path string
	var payload any

	mediaRelKey := "anime"
	mediaType := "anime"
	mediaID := input.AnimeID
	if input.MangaID != "" {
		mediaRelKey = "manga"
		mediaType = "manga"
		mediaID = input.MangaID
	}

	if targetID != "" {
		// PATCH existing entry
		method = http.MethodPatch
		path = "/library-entries/" + url.PathEscape(targetID)
		payload = map[string]any{
			"data": map[string]any{
				"id":         targetID,
				"type":       "libraryEntries",
				"attributes": attributes,
			},
		}
	} else {
		// POST create new entry
		method = http.MethodPost
		path = "/library-entries"
		payload = map[string]any{
			"data": map[string]any{
				"type":       "libraryEntries",
				"attributes": attributes,
				"relationships": map[string]any{
					mediaRelKey: map[string]any{
						"data": map[string]string{
							"type": mediaType,
							"id":   mediaID,
						},
					},
					"user": map[string]any{
						"data": map[string]string{
							"type": "users",
							"id":   input.UserID,
						},
					},
				},
			},
		}
	}

	doc, resp, err := s.client.doRequestWithJSON(ctx, method, path, payload)
	if err != nil {
		return nil, resp, fmt.Errorf("kitsu: save entry for %s %s: %w", mediaType, mediaID, err)
	}

	if doc == nil || len(doc.Data) == 0 {
		return nil, resp, fmt.Errorf("kitsu: save entry for %s %s: unexpected empty result", mediaType, mediaID)
	}

	var res jsonAPIResource
	if err := json.Unmarshal(doc.Data, &res); err != nil {
		return nil, resp, fmt.Errorf("kitsu: unmarshal saved library entry: %w", err)
	}

	entry, err := unmarshalLibraryEntry(res, doc.Included)
	if err != nil {
		return nil, resp, err
	}

	return entry, resp, nil
}

// DeleteEntry removes an entry from the library by its entry ID.
func (s *LibraryEntryService) DeleteEntry(ctx context.Context, id string) (*Response, error) {
	if id == "" {
		return nil, errors.New("kitsu: entry id must not be empty")
	}

	_, resp, err := s.client.doRequest(ctx, http.MethodDelete, "/library-entries/"+url.PathEscape(id), nil)
	if err != nil {
		return resp, fmt.Errorf("kitsu: delete entry %s: %w", id, err)
	}

	return resp, nil
}
