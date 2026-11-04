package kitsu

import (
	"net/http"
	"time"
)

const (
	// DefaultBaseURL is the default JSON:API base endpoint for Kitsu.io.
	DefaultBaseURL = "https://kitsu.io/api/edge"

	// DefaultTokenURL is the default OAuth2 token endpoint for Kitsu.io.
	DefaultTokenURL = "https://kitsu.io/api/oauth/token"

	// DefaultTimeout is the default timeout for the HTTP client.
	DefaultTimeout = 30 * time.Second

	// DefaultUserAgent is the default User-Agent header sent with requests.
	DefaultUserAgent = "go-kitsu/0.1.0"
)

// Client manages communication with the Kitsu.io JSON:API.
type Client struct {
	client    *http.Client
	baseURL   string
	tokenURL  string
	token     string
	userAgent string

	// Services
	Anime        *AnimeService
	LibraryEntry *LibraryEntryService
	User         *UserService
	Mapping      *MappingService
	Auth         *AuthService
}

// AnimeService handles communication with anime catalog endpoints.
type AnimeService struct {
	client *Client
}

// LibraryEntryService handles communication with user library and tracking endpoints.
type LibraryEntryService struct {
	client *Client
}

// UserService handles communication with user profile endpoints.
type UserService struct {
	client *Client
}

// MappingService handles external ID cross-referencing (e.g. MyAnimeList / TVDB to Kitsu).
type MappingService struct {
	client *Client
}

// AuthService handles authentication with the Kitsu OAuth2 API.
type AuthService struct {
	client *Client
}

// NewClient returns a new Kitsu API client. If httpClient is nil, a default
// http.Client with DefaultTimeout will be used.
func NewClient(httpClient *http.Client, opts ...ClientOption) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: DefaultTimeout,
		}
	}

	c := &Client{
		client:    httpClient,
		baseURL:   DefaultBaseURL,
		tokenURL:  DefaultTokenURL,
		userAgent: DefaultUserAgent,
	}

	for _, opt := range opts {
		opt(c)
	}

	c.Anime = &AnimeService{client: c}
	c.LibraryEntry = &LibraryEntryService{client: c}
	c.User = &UserService{client: c}
	c.Mapping = &MappingService{client: c}
	c.Auth = &AuthService{client: c}

	return c
}
