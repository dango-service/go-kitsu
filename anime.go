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

// AnimeSubtype represents the release format of an anime.
type AnimeSubtype string

const (
	AnimeSubtypeTV      AnimeSubtype = "TV"
	AnimeSubtypeMovie   AnimeSubtype = "movie"
	AnimeSubtypeOVA     AnimeSubtype = "OVA"
	AnimeSubtypeONA     AnimeSubtype = "ONA"
	AnimeSubtypeSpecial AnimeSubtype = "special"
	AnimeSubtypeMusic   AnimeSubtype = "music"
)

// AnimeStatus represents the release status of an anime.
type AnimeStatus string

const (
	AnimeStatusCurrent    AnimeStatus = "current"
	AnimeStatusFinished   AnimeStatus = "finished"
	AnimeStatusTBA        AnimeStatus = "tba"
	AnimeStatusUnreleased AnimeStatus = "unreleased"
	AnimeStatusUpcoming   AnimeStatus = "upcoming"
)

// AnimeTitles contains title representations across languages and scripts.
type AnimeTitles struct {
	Romaji   string  `json:"en_jp,omitempty"`
	English  *string `json:"en,omitempty"`
	Japanese *string `json:"ja_jp,omitempty"`
}

// Anime represents an anime record in Kitsu.
type Anime struct {
	ID             string       `json:"id"`
	Slug           string       `json:"slug"`
	Synopsis       string       `json:"synopsis"`
	CanonicalTitle string       `json:"canonicalTitle"`
	Titles         AnimeTitles  `json:"titles"`
	Subtype        AnimeSubtype `json:"subtype"`
	Status         AnimeStatus  `json:"status"`
	EpisodeCount   *int         `json:"episodeCount,omitempty"`
	EpisodeLength  *int         `json:"episodeLength,omitempty"`
	StartDate      *string      `json:"startDate,omitempty"`
	EndDate        *string      `json:"endDate,omitempty"`
	AgeRating      *string      `json:"ageRating,omitempty"`
}

func unmarshalAnime(res jsonAPIResource) (*Anime, error) {
	var a Anime
	if len(res.Attributes) > 0 {
		if err := json.Unmarshal(res.Attributes, &a); err != nil {
			return nil, fmt.Errorf("kitsu: unmarshal anime attributes: %w", err)
		}
	}
	a.ID = res.ID
	return &a, nil
}

// GetByID retrieves an anime record by its Kitsu ID.
func (s *AnimeService) GetByID(ctx context.Context, id string) (*Anime, *Response, error) {
	if id == "" {
		return nil, nil, errors.New("kitsu: anime id must not be empty")
	}

	doc, resp, err := s.client.doRequest(ctx, http.MethodGet, "/anime/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, resp, fmt.Errorf("kitsu: get anime by id %s: %w", id, err)
	}

	if doc == nil || len(doc.Data) == 0 {
		return nil, resp, fmt.Errorf("kitsu: get anime by id %s: %w", id, ErrNotFound)
	}

	var res jsonAPIResource
	if err := json.Unmarshal(doc.Data, &res); err != nil {
		return nil, resp, fmt.Errorf("kitsu: unmarshal anime resource: %w", err)
	}

	anime, err := unmarshalAnime(res)
	if err != nil {
		return nil, resp, err
	}

	return anime, resp, nil
}

// Search queries Kitsu for anime matching search criteria and optional filters.
func (s *AnimeService) Search(ctx context.Context, query string, opts ...AnimeSearchOption) ([]Anime, *Response, error) {
	params := animeSearchParams{
		limit: 10,
	}
	for _, opt := range opts {
		opt(&params)
	}

	if params.limit > 20 {
		params.limit = 20
	}

	v := url.Values{}
	if query != "" {
		v.Set("filter[text]", query)
	}
	if params.subtype != "" {
		v.Set("filter[subtype]", string(params.subtype))
	}
	if params.seasonYear > 0 {
		v.Set("filter[seasonYear]", strconv.Itoa(params.seasonYear))
	}
	v.Set("page[limit]", strconv.Itoa(params.limit))

	path := "/anime?" + v.Encode()
	doc, resp, err := s.client.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, resp, fmt.Errorf("kitsu: search anime %q: %w", query, err)
	}

	if doc == nil || len(doc.Data) == 0 {
		return []Anime{}, resp, nil
	}

	var resources []jsonAPIResource
	if err := json.Unmarshal(doc.Data, &resources); err != nil {
		return nil, resp, fmt.Errorf("kitsu: unmarshal search anime resources: %w", err)
	}

	animes := make([]Anime, 0, len(resources))
	for _, res := range resources {
		anime, err := unmarshalAnime(res)
		if err != nil {
			return nil, resp, err
		}
		animes = append(animes, *anime)
	}

	return animes, resp, nil
}
