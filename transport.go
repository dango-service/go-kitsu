package kitsu

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	// MediaTypeJSONAPI is the standard media type required by the JSON:API specification.
	MediaTypeJSONAPI = "application/vnd.api+json"

	// maxResponseBodyBytes limits HTTP response body reading to 10 MiB to prevent memory exhaustion DoS.
	maxResponseBodyBytes = 10 * 1024 * 1024
)

// jsonAPIDocument represents the top-level document structure of JSON:API responses.
type jsonAPIDocument struct {
	Data     json.RawMessage   `json:"data,omitempty"`
	Included []jsonAPIResource `json:"included,omitempty"`
	Meta     map[string]any    `json:"meta,omitempty"`
	Links    *jsonAPILinks     `json:"links,omitempty"`
	Errors   []JSONAPIError    `json:"errors,omitempty"`
}

// jsonAPIResource represents a single primary or included resource in a JSON:API document.
type jsonAPIResource struct {
	ID            string                          `json:"id"`
	Type          string                          `json:"type"`
	Attributes    json.RawMessage                 `json:"attributes,omitempty"`
	Relationships map[string]jsonAPIRelationship  `json:"relationships,omitempty"`
	Links         *jsonAPILinks                   `json:"links,omitempty"`
}

// jsonAPIRelationship represents relationships linked to a resource.
type jsonAPIRelationship struct {
	Data  json.RawMessage `json:"data,omitempty"`
	Links *jsonAPILinks   `json:"links,omitempty"`
}

// jsonAPILinks represents pagination and navigation hypermedia links.
type jsonAPILinks struct {
	Self    string `json:"self,omitempty"`
	Related string `json:"related,omitempty"`
	First   string `json:"first,omitempty"`
	Prev    string `json:"prev,omitempty"`
	Next    string `json:"next,omitempty"`
	Last    string `json:"last,omitempty"`
}

// doRequest performs an HTTP request, parses JSON:API envelope, inspects rate limit headers,
// and returns the parsed document and response metadata.
func (c *Client) doRequest(ctx context.Context, method, endpoint string, body io.Reader) (*jsonAPIDocument, *Response, error) {
	if ctx == nil {
		return nil, nil, errors.New("kitsu: context must not be nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, fmt.Errorf("kitsu: context cancelled: %w", err)
	}

	fullURL := endpoint
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		base := strings.TrimRight(c.baseURL, "/")
		if !strings.HasPrefix(endpoint, "/") {
			fullURL = base + "/" + endpoint
		} else {
			fullURL = base + endpoint
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, nil, fmt.Errorf("kitsu: failed to create request: %w", err)
	}

	req.Header.Set("Accept", MediaTypeJSONAPI)
	if body != nil {
		req.Header.Set("Content-Type", MediaTypeJSONAPI)
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	httpResp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("kitsu: request failed: %w", err)
	}
	defer httpResp.Body.Close()

	resp := newResponse(httpResp)

	// HTTP 204 No Content (typically DELETE mutations)
	if httpResp.StatusCode == http.StatusNoContent {
		return nil, resp, nil
	}

	// Limit response reading to bounded memory buffer
	limitedReader := io.LimitReader(httpResp.Body, maxResponseBodyBytes)

	var doc jsonAPIDocument
	decErr := json.NewDecoder(limitedReader).Decode(&doc)
	if decErr != nil && decErr != io.EOF {
		if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
			return nil, resp, &ErrorResponse{
				Response: httpResp,
				Message:  fmt.Sprintf("http status %d: failed to decode response", httpResp.StatusCode),
			}
		}
		return nil, resp, fmt.Errorf("kitsu: failed to decode json:api response: %w", decErr)
	}

	if len(doc.Errors) > 0 {
		return nil, resp, &ErrorResponse{
			Response: httpResp,
			Errors:   doc.Errors,
		}
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, resp, &ErrorResponse{
			Response: httpResp,
		}
	}

	return &doc, resp, nil
}

// doRequestWithJSON is a convenience wrapper that encodes a JSON payload into the request.
func (c *Client) doRequestWithJSON(ctx context.Context, method, endpoint string, payload any) (*jsonAPIDocument, *Response, error) {
	var body io.Reader
	if payload != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(payload); err != nil {
			return nil, nil, fmt.Errorf("kitsu: failed to encode request body: %w", err)
		}
		body = &buf
	}
	return c.doRequest(ctx, method, endpoint, body)
}
