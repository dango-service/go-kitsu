# Testing Strategy  -  `go-kitsu`

This document defines testing standards for `go-kitsu`, ensuring deterministic,
fast, isolated, and reproducible test suites without external network access.

---

## 1. Core Testing Principles

1. **Zero Real Network Traffic**: No unit test may connect to `https://kitsu.io`. All network traffic is mocked locally.
2. **Determinism and Speed**: The full suite must complete in less than 2 seconds.
3. **Concurrency Safety**: All tests must pass cleanly under the Go race detector (`go test -v -race ./...`).
4. **Bidirectional Verification**: Tests must verify both what the client *sends* (method, URL query parameters, JSON:API body, HTTP headers) and what it *parses* (typed domain models, rate limit metadata, errors).

---

## 2. Local Test Server (`httptest.Server`)

Kitsu's RESTful JSON:API routes operations across standard HTTP methods (`GET`, `POST`, `PATCH`, `DELETE`) with standard paths (`/api/edge/anime`, `/api/edge/library-entries`).

A local standard `http.ServeMux` or `httptest.Server` handles requests deterministically:

```go
func setupTestServer(t *testing.T) (*Client, *http.ServeMux, func()) {
    t.Helper()
    mux := http.NewServeMux()
    server := httptest.NewServer(mux)

    client := NewClient(nil, WithBaseURL(server.URL+"/api/edge"), WithTokenURL(server.URL+"/api/oauth/token"))
    return client, mux, server.Close
}
```

---

## 3. Fixtures and Test Data (`testdata/`)

Mock JSON responses are kept in `testdata/`:

```
testdata/
  anime_get_success.json            <- Single anime resource
  anime_search_success.json         <- Anime collection with filters
  mapping_get_success.json          <- External site mapping
  library_entry_get_success.json    <- Single library entry
  library_entry_save_success.json   <- Updated library entry
  token_success.json                <- OAuth2 token response
```

Loaded in unit tests via `os.ReadFile("testdata/...")`.

---

## 4. Context Cancellation Testing

Each service must verify that a cancelled `context.Context` (`context.Canceled` or
`context.DeadlineExceeded`) immediately aborts the HTTP request without goroutine leaks.
