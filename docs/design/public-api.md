# Go Public API Specification  -  `go-kitsu`

This document defines the exact Go public API signatures, models, options, and contracts
for `go-kitsu`. The design mirrors the clean architecture of `go-anilist` and allows `Dango`
to wrap it seamlessly behind its `Tracker` domain interface.

---

## 1. Package and Core Client

```go
package kitsu

import (
    "context"
    "net/http"
    "time"
)

// Client manages communication with the Kitsu.io JSON:API.
type Client struct {
    client    *http.Client
    baseURL   string
    tokenURL  string
    token     string
    userAgent string

    Anime        *AnimeService
    LibraryEntry *LibraryEntryService
    User         *UserService
    Mapping      *MappingService
    Auth         *AuthService
}

// NewClient returns a new Kitsu API client.
// If httpClient is nil, a default http.Client with DefaultTimeout (30s) is used.
func NewClient(httpClient *http.Client, opts ...ClientOption) *Client

// ClientOption configures a Client using functional options.
type ClientOption func(*Client)

func WithBaseURL(url string) ClientOption
func WithTokenURL(url string) ClientOption
func WithToken(token string) ClientOption
func WithUserAgent(ua string) ClientOption
```

---

## 2. Domain Types and Models

```go
type AnimeSubtype string

const (
    AnimeSubtypeTV      AnimeSubtype = "TV"
    AnimeSubtypeMovie   AnimeSubtype = "movie"
    AnimeSubtypeOVA     AnimeSubtype = "OVA"
    AnimeSubtypeONA     AnimeSubtype = "ONA"
    AnimeSubtypeSpecial AnimeSubtype = "special"
    AnimeSubtypeMusic   AnimeSubtype = "music"
)

type AnimeStatus string

const (
    AnimeStatusCurrent    AnimeStatus = "current"
    AnimeStatusFinished   AnimeStatus = "finished"
    AnimeStatusTBA        AnimeStatus = "tba"
    AnimeStatusUnreleased AnimeStatus = "unreleased"
    AnimeStatusUpcoming   AnimeStatus = "upcoming"
)

type AnimeTitles struct {
    Romaji   string  `json:"en_jp,omitempty"`
    English  *string `json:"en,omitempty"`
    Japanese *string `json:"ja_jp,omitempty"`
}

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

type LibraryStatus string

const (
    LibraryStatusCurrent   LibraryStatus = "current"
    LibraryStatusPlanned   LibraryStatus = "planned"
    LibraryStatusCompleted LibraryStatus = "completed"
    LibraryStatusOnHold    LibraryStatus = "on_hold"
    LibraryStatusDropped   LibraryStatus = "dropped"
)

type LibraryEntry struct {
    ID             string        `json:"id"`
    Status         LibraryStatus `json:"status"`
    Progress       int           `json:"progress"`
    RatingTwenty   *int          `json:"ratingTwenty,omitempty"` // Rating scale: 2..20
    Reconsuming    bool          `json:"reconsuming"`
    ReconsumeCount int           `json:"reconsumeCount"`
    UpdatedAt      time.Time     `json:"updatedAt"`
    AnimeID        string        `json:"animeId,omitempty"`
    Anime          *Anime        `json:"anime,omitempty"`
}

type SaveLibraryEntryInput struct {
    ID           string         `json:"id,omitempty"` // If set, performs PATCH; if empty, POST
    AnimeID      string         `json:"animeId"`
    UserID       string         `json:"userId"`
    Status       *LibraryStatus `json:"status,omitempty"`
    Progress     *int           `json:"progress,omitempty"`
    RatingTwenty *int           `json:"ratingTwenty,omitempty"`
}

type User struct {
    ID     string `json:"id"`
    Name   string `json:"name"`
    Slug   string `json:"slug"`
    Avatar string `json:"avatar,omitempty"`
}

type Mapping struct {
    ID           string `json:"id"`
    ExternalSite string `json:"externalSite"`
    ExternalID   string `json:"externalId"`
}
```

---

## 3. Services and Method Signatures

### 3.1. AnimeService
```go
type AnimeService struct {
    client *Client
}

func (s *AnimeService) GetByID(ctx context.Context, id string) (*Anime, *Response, error)
func (s *AnimeService) Search(ctx context.Context, query string, opts ...SearchOption) ([]Anime, *Response, error)

type SearchOption func(*searchParams)
func WithSubtype(subtype AnimeSubtype) SearchOption
func WithSeasonYear(year int) SearchOption
func WithLimit(limit int) SearchOption
```

### 3.2. LibraryEntryService
```go
type LibraryEntryService struct {
    client *Client
}

func (s *LibraryEntryService) GetEntry(ctx context.Context, userID, animeID string) (*LibraryEntry, *Response, error)
func (s *LibraryEntryService) GetCollection(ctx context.Context, userID string, opts ...CollectionOption) ([]LibraryEntry, *Response, error)
func (s *LibraryEntryService) SaveEntry(ctx context.Context, input SaveLibraryEntryInput) (*LibraryEntry, *Response, error)
func (s *LibraryEntryService) DeleteEntry(ctx context.Context, id string) (*Response, error)

type CollectionOption func(*collectionParams)
func WithStatus(status LibraryStatus) CollectionOption
```

### 3.3. MappingService
```go
type MappingService struct {
    client *Client
}

func (s *MappingService) FindAnimeByExternalID(ctx context.Context, externalSite, externalID string) (string, *Response, error)
```

### 3.4. UserService
```go
type UserService struct {
    client *Client
}

func (s *UserService) CurrentUser(ctx context.Context) (*User, *Response, error)
func (s *UserService) GetByID(ctx context.Context, id string) (*User, *Response, error)
```

### 3.5. AuthService
```go
type AuthService struct {
    client *Client
}

type Token struct {
    AccessToken  string    `json:"access_token"`
    TokenType    string    `json:"token_type"`
    ExpiresIn    int       `json:"expires_in"`
    RefreshToken string    `json:"refresh_token"`
    Expiry       time.Time `json:"expiry"`
}

// Redact credentials in string representations to prevent log leakage.
func (t Token) String() string
func (t Token) GoString() string

func (s *AuthService) LoginPassword(ctx context.Context, username, password string) (*Token, error)
func (s *AuthService) ExchangeCode(ctx context.Context, clientID, clientSecret, redirectURI, code string) (*Token, error)
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*Token, error)
```

---

## 4. Structured Error Handling

```go
var (
    ErrNotFound     = errors.New("kitsu: resource not found")
    ErrUnauthorized = errors.New("kitsu: unauthorized or invalid credentials")
    ErrRateLimited  = errors.New("kitsu: rate limit exceeded (429)")
)

type JSONAPIError struct {
    ID     string `json:"id,omitempty"`
    Status string `json:"status,omitempty"`
    Code   string `json:"code,omitempty"`
    Title  string `json:"title,omitempty"`
    Detail string `json:"detail,omitempty"`
}

type ErrorResponse struct {
    Response *http.Response
    Errors   []JSONAPIError
    Message  string
}
```
