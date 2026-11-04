# go-kitsu

[![Go Reference](https://pkg.go.dev/badge/github.com/dango-service/go-kitsu.svg)](https://pkg.go.dev/github.com/dango-service/go-kitsu)
[![CI](https://github.com/dango-service/go-kitsu/actions/workflows/ci.yml/badge.svg)](https://github.com/dango-service/go-kitsu/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A lightweight, standalone Go client library for the [Kitsu.io](https://kitsu.io) JSON:API (v3 / edge).
Designed with **zero external dependencies** (standard library only), direct streaming deserialization bounded by memory limits, and strict efficiency for background sync daemons like [Dango](https://github.com/dango-service/dango).

---

## Features

- **Zero External Dependencies**: Built entirely upon the Go standard library (`net/http`, `encoding/json`, `context`, `time`, `net/url`).
- **Streaming JSON:API**: Direct streaming decoding with `json.NewDecoder` bounded by `io.LimitReader` (10 MiB limit) to prevent memory exhaustion DoS.
- **Rate Limit Aware**: Automatic inspection of quota headers (`X-RateLimit-*`) and native handling of HTTP `429 Too Many Requests` (`Retry-After` in seconds or RFC 7231 HTTP-Date).
- **Credential Protection**: Redacted string formatting (`[REDACTED]`) on `Token` to prevent secret leakage in logs.
- **Modular Services**:
  - `Anime`: Search and lookup anime records by Kitsu ID.
  - `Mapping`: Resolve external IDs (MyAnimeList, AniList, TVDB) to Kitsu ID.
  - `LibraryEntry`: Watch status tracking, user collections, and progress mutations (`SaveEntry`).
  - `User`: Profile and user identification (`CurrentUser`).
  - `Auth`: Full OAuth2 authentication (Password grant, Code grant, Refresh token).
- **100% Deterministic Testing**: 29 unit tests passing in milliseconds using `net/http/httptest` with zero real-network requests.

---

## Installation

```bash
go get github.com/dango-service/go-kitsu
```

---

## Usage Examples

### 1. Client Initialization

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/dango-service/go-kitsu"
)

func main() {
    // Initialize client with a static OAuth2 Bearer token
    client := kitsu.NewClient(nil, kitsu.WithToken("your_access_token_here"))

    // Fetch authenticated user profile
    user, resp, err := client.User.CurrentUser(context.Background())
    if err != nil {
        log.Fatalf("Error fetching user: %v", err)
    }

    fmt.Printf("User: %s (ID: %s, Slug: %s)\n", user.Name, user.ID, user.Slug)
    fmt.Printf("Remaining API quota: %d requests\n", resp.RateLimit.Remaining)
}
```

### 2. Anime Search and External Mapping

```go
// Search anime by title
results, _, err := client.Anime.Search(
    ctx,
    "Sousou no Frieren",
    kitsu.WithSubtype(kitsu.AnimeSubtypeTV),
    kitsu.WithLimit(5),
)
if err != nil {
    log.Fatal(err)
}

for _, anime := range results {
    fmt.Printf("[%s] %s (%d episodes)\n", anime.ID, anime.CanonicalTitle, *anime.EpisodeCount)
}

// Resolve MyAnimeList ID to Kitsu ID
kitsuID, _, err := client.Mapping.FindAnimeByExternalID(ctx, "myanimelist/anime", "52991")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Resolved MAL 52991 -> Kitsu %s\n", kitsuID)
```

### 3. Update Watch Progress

```go
progress := 16
status := kitsu.LibraryStatusCurrent
rating := 18 // 2..20 scale (e.g. 9.0/10)

entry, _, err := client.LibraryEntry.SaveEntry(ctx, kitsu.SaveLibraryEntryInput{
    UserID:       "42",
    AnimeID:      "1",
    Progress:     &progress,
    Status:       &status,
    RatingTwenty: &rating,
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Updated entry %s: Episode %d (%s)\n", entry.ID, entry.Progress, entry.Status)
```

### 4. Structured Error Handling

```go
if errors.Is(err, kitsu.ErrNotFound) {
    fmt.Println("Requested resource not found on Kitsu")
} else if errors.Is(err, kitsu.ErrRateLimited) {
    fmt.Println("Rate limit exceeded (HTTP 429)")
} else if errors.Is(err, kitsu.ErrUnauthorized) {
    fmt.Println("Invalid or expired OAuth2 credentials")
}
```

---

## License

Distributed under the [MIT License](LICENSE).
