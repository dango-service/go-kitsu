# JSON:API v1.0 Integration Strategy  -  `go-kitsu`

This document outlines how `go-kitsu` interacts with Kitsu's JSON:API specification (`application/vnd.api+json`).

---

## 1. Protocol Specifications

Kitsu uses the official **JSON:API** v1.0 standard:
- **Base Endpoint**: `https://kitsu.io/api/edge`
- **Media Type**: `application/vnd.api+json`
- Both `Accept` and `Content-Type` headers must be set to `application/vnd.api+json`.

---

## 2. Document Structure

Responses from Kitsu conform to the JSON:API top-level envelope:

```json
{
  "data": {
    "id": "12",
    "type": "anime",
    "attributes": {
      "canonicalTitle": "Cowboy Bebop",
      "episodeCount": 26
    },
    "relationships": {
      "genres": {
        "links": {
          "self": "...",
          "related": "..."
        }
      }
    }
  },
  "included": [],
  "meta": {
    "count": 1
  },
  "links": {
    "first": "...",
    "next": "...",
    "last": "..."
  }
}
```

---

## 3. Streaming and Parsing Strategy

To fulfill the zero-dependency and minimal-memory constraints:

1. **No heavy reflection or third-party JSON:API packages**:
   - Generic standard envelope struct `jsonAPIDocument` handles `data`, `included`, and `errors`.
   - Specific model types parse directly from the inner `attributes` and `relationships`.

2. **Bounded Response Consumption**:
   - Every response body is limited using `io.LimitReader(resp.Body, maxBodyBytes)` (default 10 MiB) before passing to `json.NewDecoder`.
   - This prevents denial-of-service from unbounded streaming responses or malicious proxies.

3. **Error Handling**:
   - JSON:API errors return HTTP 4xx/5xx with an `errors` array containing `status`, `title`, and `detail`.
   - The transport layer maps these into Go's `ErrorResponse` and matches sentinel errors (`ErrNotFound`, `ErrUnauthorized`, `ErrRateLimited`).
