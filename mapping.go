package kitsu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// Mapping represents external site cross-referencing information for media.
type Mapping struct {
	ID           string `json:"id"`
	ExternalSite string `json:"externalSite"`
	ExternalID   string `json:"externalId"`
	AnimeID      string `json:"animeId,omitempty"`
}

// FindAnimeByExternalID resolves an external database reference (such as "myanimelist/anime",
// "thetvdb/series", or "anidb") into the corresponding Kitsu anime ID.
func (s *MappingService) FindAnimeByExternalID(ctx context.Context, externalSite, externalID string) (string, *Response, error) {
	if externalSite == "" || externalID == "" {
		return "", nil, errors.New("kitsu: externalSite and externalID must not be empty")
	}

	v := url.Values{}
	v.Set("filter[externalSite]", externalSite)
	v.Set("filter[externalId]", externalID)
	v.Set("include", "item")

	path := "/mappings?" + v.Encode()
	doc, resp, err := s.client.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return "", resp, fmt.Errorf("kitsu: find mapping (%s, %s): %w", externalSite, externalID, err)
	}

	if doc == nil || len(doc.Data) == 0 {
		return "", resp, fmt.Errorf("kitsu: find mapping (%s, %s): %w", externalSite, externalID, ErrNotFound)
	}

	var resources []jsonAPIResource
	if err := json.Unmarshal(doc.Data, &resources); err != nil {
		return "", resp, fmt.Errorf("kitsu: unmarshal mapping resources: %w", err)
	}

	if len(resources) == 0 {
		return "", resp, fmt.Errorf("kitsu: find mapping (%s, %s): %w", externalSite, externalID, ErrNotFound)
	}

	itemRel, ok := resources[0].Relationships["item"]
	if !ok || len(itemRel.Data) == 0 {
		return "", resp, fmt.Errorf("kitsu: mapping (%s, %s) has no item relationship: %w", externalSite, externalID, ErrNotFound)
	}

	var itemIdentifier struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}
	if err := json.Unmarshal(itemRel.Data, &itemIdentifier); err != nil {
		return "", resp, fmt.Errorf("kitsu: unmarshal mapping relationship: %w", err)
	}

	if itemIdentifier.ID == "" {
		return "", resp, fmt.Errorf("kitsu: mapping item has empty id: %w", ErrNotFound)
	}

	return itemIdentifier.ID, resp, nil
}
