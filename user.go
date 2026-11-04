package kitsu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// User represents a user account profile in Kitsu.
type User struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Avatar string `json:"avatar,omitempty"`
}

func unmarshalUser(res jsonAPIResource) (*User, error) {
	var user User
	if len(res.Attributes) > 0 {
		var attrs struct {
			Name   string `json:"name"`
			Slug   string `json:"slug"`
			Avatar *struct {
				Original string `json:"original,omitempty"`
				Medium   string `json:"medium,omitempty"`
			} `json:"avatar,omitempty"`
		}
		if err := json.Unmarshal(res.Attributes, &attrs); err != nil {
			return nil, fmt.Errorf("kitsu: unmarshal user attributes: %w", err)
		}
		user.Name = attrs.Name
		user.Slug = attrs.Slug
		if attrs.Avatar != nil {
			if attrs.Avatar.Original != "" {
				user.Avatar = attrs.Avatar.Original
			} else {
				user.Avatar = attrs.Avatar.Medium
			}
		}
	}
	user.ID = res.ID
	return &user, nil
}

// CurrentUser retrieves the authenticated user's profile using the self filter.
func (s *UserService) CurrentUser(ctx context.Context) (*User, *Response, error) {
	doc, resp, err := s.client.doRequest(ctx, http.MethodGet, "/users?filter[self]=true", nil)
	if err != nil {
		return nil, resp, fmt.Errorf("kitsu: get current user: %w", err)
	}

	if doc == nil || len(doc.Data) == 0 {
		return nil, resp, fmt.Errorf("kitsu: get current user: %w", ErrNotFound)
	}

	var resources []jsonAPIResource
	if err := json.Unmarshal(doc.Data, &resources); err != nil {
		return nil, resp, fmt.Errorf("kitsu: unmarshal current user resources: %w", err)
	}

	if len(resources) == 0 {
		return nil, resp, fmt.Errorf("kitsu: get current user: %w", ErrNotFound)
	}

	user, err := unmarshalUser(resources[0])
	if err != nil {
		return nil, resp, err
	}

	return user, resp, nil
}

// GetByID retrieves a user profile by numeric/string Kitsu ID.
func (s *UserService) GetByID(ctx context.Context, id string) (*User, *Response, error) {
	if id == "" {
		return nil, nil, errors.New("kitsu: user id must not be empty")
	}

	doc, resp, err := s.client.doRequest(ctx, http.MethodGet, "/users/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, resp, fmt.Errorf("kitsu: get user by id %s: %w", id, err)
	}

	if doc == nil || len(doc.Data) == 0 {
		return nil, resp, fmt.Errorf("kitsu: get user by id %s: %w", id, ErrNotFound)
	}

	var res jsonAPIResource
	if err := json.Unmarshal(doc.Data, &res); err != nil {
		return nil, resp, fmt.Errorf("kitsu: unmarshal user resource: %w", err)
	}

	user, err := unmarshalUser(res)
	if err != nil {
		return nil, resp, err
	}

	return user, resp, nil
}
