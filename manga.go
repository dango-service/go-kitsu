package kitsu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// MangaSubtype represents the format of a manga or comic work.
type MangaSubtype string

const (
	MangaSubtypeManga   MangaSubtype = "manga"
	MangaSubtypeNovel   MangaSubtype = "novel"
	MangaSubtypeManhwa  MangaSubtype = "manhwa"
	MangaSubtypeManhua  MangaSubtype = "manhua"
	MangaSubtypeOneshot MangaSubtype = "oneshot"
	MangaSubtypeDoujin  MangaSubtype = "doujin"
)

// MangaStatus represents the publication status of a manga.
type MangaStatus string

const (
	MangaStatusCurrent    MangaStatus = "current"
	MangaStatusFinished   MangaStatus = "finished"
	MangaStatusTBA        MangaStatus = "tba"
	MangaStatusUnreleased MangaStatus = "unreleased"
	MangaStatusUpcoming   MangaStatus = "upcoming"
)

// MangaTitles contains title representations across languages and scripts.
type MangaTitles struct {
	Romaji   string  `json:"en_jp,omitempty"`
	English  *string `json:"en,omitempty"`
	Japanese *string `json:"ja_jp,omitempty"`
}

// Manga represents a manga, manhwa, manhua, or light novel record in Kitsu.
type Manga struct {
	ID             string       `json:"id"`
	Slug           string       `json:"slug"`
	Synopsis       string       `json:"synopsis"`
	CanonicalTitle string       `json:"canonicalTitle"`
	Titles         MangaTitles  `json:"titles"`
	Subtype        MangaSubtype `json:"subtype"`
	Status         MangaStatus  `json:"status"`
	ChapterCount   *int         `json:"chapterCount,omitempty"`
	VolumeCount    *int         `json:"volumeCount,omitempty"`
	StartDate      *string      `json:"startDate,omitempty"`
	EndDate        *string      `json:"endDate,omitempty"`
	AgeRating      *string      `json:"ageRating,omitempty"`
}

func unmarshalManga(res jsonAPIResource) (*Manga, error) {
	var m Manga
	if len(res.Attributes) > 0 {
		if err := json.Unmarshal(res.Attributes, &m); err != nil {
			return nil, fmt.Errorf("kitsu: unmarshal manga attributes: %w", err)
		}
	}
	m.ID = res.ID
	return &m, nil
}

// MangaService handles communication with manga catalog endpoints.
type MangaService struct {
	client *Client
}

// GetByID retrieves a manga record by its Kitsu ID.
func (s *MangaService) GetByID(ctx context.Context, id string) (*Manga, *Response, error) {
	if id == "" {
		return nil, nil, errors.New("kitsu: manga id must not be empty")
	}

	doc, resp, err := s.client.doRequest(ctx, http.MethodGet, "/manga/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, resp, fmt.Errorf("kitsu: get manga by id %s: %w", id, err)
	}

	if doc == nil || len(doc.Data) == 0 {
		return nil, resp, fmt.Errorf("kitsu: get manga by id %s: %w", id, ErrNotFound)
	}

	var res jsonAPIResource
	if err := json.Unmarshal(doc.Data, &res); err != nil {
		return nil, resp, fmt.Errorf("kitsu: unmarshal manga resource: %w", err)
	}

	manga, err := unmarshalManga(res)
	if err != nil {
		return nil, resp, err
	}

	return manga, resp, nil
}

// Search queries Kitsu for manga matching search criteria and optional filters.
func (s *MangaService) Search(ctx context.Context, query string, opts ...MangaSearchOption) ([]Manga, *Response, error) {
	params := mangaSearchParams{
		limit: 10,
	}
	for _, opt := range opts {
		opt(&params)
	}

	values := url.Values{}
	if query != "" {
		values.Set("filter[text]", query)
	}
	if params.limit > 0 {
		values.Set("page[limit]", strconv.Itoa(params.limit))
	}
	if params.offset > 0 {
		values.Set("page[offset]", strconv.Itoa(params.offset))
	}
	if params.subtype != "" {
		values.Set("filter[subtype]", string(params.subtype))
	}

	path := "/manga"
	if len(values) > 0 {
		path += "?" + values.Encode()
	}

	doc, resp, err := s.client.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, resp, fmt.Errorf("kitsu: search manga %q: %w", query, err)
	}

	if doc == nil || len(doc.Data) == 0 {
		return []Manga{}, resp, nil
	}

	var resources []jsonAPIResource
	if err := json.Unmarshal(doc.Data, &resources); err != nil {
		return nil, resp, fmt.Errorf("kitsu: unmarshal manga list: %w", err)
	}

	mangaList := make([]Manga, 0, len(resources))
	for _, res := range resources {
		m, err := unmarshalManga(res)
		if err != nil {
			return nil, resp, err
		}
		mangaList = append(mangaList, *m)
	}

	return mangaList, resp, nil
}
