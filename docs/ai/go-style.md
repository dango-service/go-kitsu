# Go Style Guidelines  -  go-kitsu

Technical and code style standards for Go code in the `go-kitsu` library.

---

## 1. Error Handling

- **Always wrap with `%w`**: All propagated errors must be wrapped using
  `fmt.Errorf("kitsu: <op>: %w", err)`. Never discard errors silently with `_`
  in critical execution paths.
- **Typed and Sentinel Errors**:
  - Define sentinel errors for predictable conditions (`ErrNotFound`, `ErrRateLimited`, `ErrUnauthorized`).
  - Provide structured types `JSONAPIError` and `ErrorResponse` exposing HTTP status codes and API error details.
  - Implement `errors.Is` / `errors.As` support so consumers can inspect underlying causes.

---

## 2. Context and Cancellation

- **`context.Context` as First Parameter**: Any method or function performing HTTP, I/O, or asynchronous work must accept `ctx context.Context`:
  ```go
  func (s *AnimeService) GetByID(ctx context.Context, id string) (*Anime, *Response, error)
  ```
- **Respect Cancellation**: Propagate the context to `http.NewRequestWithContext(ctx, ...)` and check `ctx.Err()`.
- If `ctx == nil`, reject the request immediately with an error before initiating network traffic.

---

## 3. State, Global Variables, and `init()` Functions

- **No Mutable Global Variables**: Zero shared mutable global state. Each `Client` instance manages its own configuration and HTTP transport independently.
- **No `init()` Functions**: No hidden registration in `init()`. Configuration must be explicit via `NewClient` functional options.
- Constants for endpoints and URLs must be immutable.

---

## 4. Interface and API Design

- **Small, Consumer-Driven Interfaces**:
  - `go-kitsu` exposes concrete structs for services (`AnimeService`, `LibraryEntryService`, etc.) and returned domain types.
  - Do not force giant interfaces inside the library; interfaces should be defined by the consumer (e.g., in `Dango` via its `Tracker` interface).
- **Typed Functional Options**: Use typed functional options for optional parameters and filters (e.g., `WithSubtype(AnimeSubtypeTV)`, `WithLimit(10)`).
- **Dependency Injection**:
  - The client must accept an injected `*http.Client`:
    ```go
    func NewClient(httpClient *http.Client, opts ...ClientOption) *Client
    ```
  - If `httpClient == nil`, default to an `http.Client` with standard timeouts (30s).

---

## 5. Security & Defensive Boundaries

- **Bounded Stream Reading**: Always wrap response bodies with `io.LimitReader(body, maxBodyBytes)` (10 MiB limit) to prevent memory exhaustion / DoS attacks.
- **Credential Redaction**: Types holding credentials (like `Token`) must implement `fmt.Stringer` and `fmt.GoStringer` returning `[REDACTED]` to prevent secret leakage in application logs.
- **SSRF Validation**: URL overrides via options must validate the URI scheme (`http` / `https`).

---

## 6. Testing Strategy

- **Table-Driven Tests**: Use test tables with named subtests via `t.Run(tc.name, func(t *testing.T) { ... })`.
- **Zero Real Network Calls**: Unit tests must never depend on live Kitsu servers. Rely exclusively on `net/http/httptest.Server` and mock JSON fixtures.
- **Race Detector**: All tests must pass cleanly under `go test -v -race ./...`.

---

## 7. Build and Compilation

- **`CGO_ENABLED=0`**: All code must build cleanly without CGO for static portability.
- **External Dependencies**:
  - Zero external dependencies. Always use the Go standard library (`net/http`, `encoding/json`, `context`, `time`, `net/url`).
