# Architectural Invariants  -  go-kitsu

Structural rules and invariants that must be upheld at all times in `go-kitsu`.

---

## 1. Independence and Decoupling

- **Completely Agnostic of Dango**: `go-kitsu` is an autonomous library.
  Under no circumstances may it import `Dango` packages, reference Plex, SQLite,
  media matching algorithms, or webhook handlers.
- **Zero Global State**: Multiple `Client` instances must be able to run concurrently
  and safely within the same process without interference.

---

## 2. Layer Separation: JSON:API Transport vs Typed Public API

1. **Internal Transport Layer (`doRequest`)**:
   - Handles HTTP communication (`GET`, `POST`, `PATCH`, `DELETE`).
   - Configures `Content-Type: application/vnd.api+json` and `Accept: application/vnd.api+json`.
   - Injects `Authorization: Bearer <token>` when authenticated.
   - Enforces response size bounds via `io.LimitReader(resp.Body, maxBodyBytes)` (10 MiB limit).
   - Deserializes the standard JSON:API top-level envelope (`data`, `included`, `meta`, `links`, `errors`).
   - Normalizes HTTP status codes and JSON:API error objects into Go error types.

2. **Public Service Layer (`AnimeService`, `LibraryEntryService`, `UserService`, `MappingService`, `AuthService`)**:
   - Exposes high-level methods with idiomatic Go domain signatures:
     - `anime.GetByID(ctx, id)`
     - `anime.Search(ctx, query, opts...)`
     - `libraryEntry.GetEntry(ctx, userID, animeID)`
     - `libraryEntry.SaveEntry(ctx, input)`
     - `mapping.FindAnimeByExternalID(ctx, site, id)`
   - Completely encapsulates raw JSON:API payloads and URL query construction.
   - Returns strongly typed Go domain models (`*Anime`, `*LibraryEntry`, `*User`).

---

## 3. Rate Limiting and Quota Management

- **Header Inspection**: The transport layer extracts `X-RateLimit-Limit`,
  `X-RateLimit-Remaining`, and `X-RateLimit-Reset` on every response.
- **Response Metadata**: Public methods return `(*T, *Response, error)` where `Response`
  includes rate limit metadata so consumers (like Dango) can schedule intelligent throttling.
- **HTTP 429 Compliance**: When receiving HTTP 429 Too Many Requests, extract `Retry-After`
  (supporting both integer seconds and RFC 7231 HTTP-Date) and return structured `ErrRateLimited`.

---

## 4. `http.Client` Lifecycle and Injection

- The client must never leave orphaned HTTP transport resources.
- Supports custom `http.Client` injection for:
  - Delegated OAuth2 authentication.
  - Proxies, metrics, or logging roundtrippers.
  - Local testing servers with `httptest.Server`.
