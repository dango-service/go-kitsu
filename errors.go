package kitsu

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	// ErrNotFound is returned when a requested resource does not exist (HTTP 404).
	ErrNotFound = errors.New("kitsu: resource not found")

	// ErrUnauthorized is returned when the request lacks valid authentication credentials (HTTP 401).
	ErrUnauthorized = errors.New("kitsu: unauthorized or invalid credentials")

	// ErrRateLimited is returned when the rate limit quota is exceeded (HTTP 429).
	ErrRateLimited = errors.New("kitsu: rate limit exceeded (429)")
)

// JSONAPISource points to the origin of an error in the request document.
type JSONAPISource struct {
	Pointer   string `json:"pointer,omitempty"`
	Parameter string `json:"parameter,omitempty"`
}

// JSONAPIError represents an individual error object in a JSON:API document.
type JSONAPIError struct {
	ID     string         `json:"id,omitempty"`
	Status string         `json:"status,omitempty"`
	Code   string         `json:"code,omitempty"`
	Title  string         `json:"title,omitempty"`
	Detail string         `json:"detail,omitempty"`
	Source *JSONAPISource `json:"source,omitempty"`
}

func (e JSONAPIError) Error() string {
	parts := make([]string, 0, 3)
	if e.Status != "" {
		parts = append(parts, fmt.Sprintf("(%s)", e.Status))
	}
	if e.Title != "" {
		parts = append(parts, e.Title)
	}
	if e.Detail != "" && e.Detail != e.Title {
		parts = append(parts, e.Detail)
	}
	if len(parts) == 0 {
		return "unspecified api error"
	}
	return strings.Join(parts, ": ")
}

// ErrorResponse wraps HTTP and JSON:API errors returned by the Kitsu API.
type ErrorResponse struct {
	Response *http.Response `json:"-"`
	Errors   []JSONAPIError `json:"errors,omitempty"`
	Message  string         `json:"message,omitempty"`
}

func (e *ErrorResponse) Error() string {
	code := 0
	if e.Response != nil {
		code = e.Response.StatusCode
	}

	if len(e.Errors) > 0 {
		msgs := make([]string, len(e.Errors))
		for i, err := range e.Errors {
			msgs[i] = err.Error()
		}
		if code > 0 {
			return fmt.Sprintf("kitsu: api error (%d): %s", code, strings.Join(msgs, "; "))
		}
		return fmt.Sprintf("kitsu: api error: %s", strings.Join(msgs, "; "))
	}

	if e.Message != "" {
		if code > 0 {
			return fmt.Sprintf("kitsu: api error (%d): %s", code, e.Message)
		}
		return fmt.Sprintf("kitsu: api error: %s", e.Message)
	}

	if code > 0 {
		return fmt.Sprintf("kitsu: api error (%d)", code)
	}
	return "kitsu: unknown api error"
}

// Is reports whether ErrorResponse matches sentinel errors.
func (e *ErrorResponse) Is(target error) bool {
	if e.Response != nil {
		if target == ErrRateLimited && e.Response.StatusCode == http.StatusTooManyRequests {
			return true
		}
		if target == ErrUnauthorized && (e.Response.StatusCode == http.StatusUnauthorized || e.Response.StatusCode == http.StatusForbidden) {
			return true
		}
		if target == ErrNotFound && e.Response.StatusCode == http.StatusNotFound {
			return true
		}
	}

	for _, je := range e.Errors {
		if target == ErrNotFound && (je.Status == "404" || je.Code == "record_not_found") {
			return true
		}
		if target == ErrRateLimited && (je.Status == "429" || je.Code == "rate_limit_exceeded") {
			return true
		}
		if target == ErrUnauthorized && (je.Status == "401" || je.Status == "403") {
			return true
		}
	}

	return false
}
