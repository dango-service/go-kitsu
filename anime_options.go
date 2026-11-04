package kitsu

type animeSearchParams struct {
	subtype    AnimeSubtype
	seasonYear int
	limit      int
}

// AnimeSearchOption configures parameters for AnimeService.Search.
type AnimeSearchOption func(*animeSearchParams)

// WithSubtype filters search results by media release subtype (TV, Movie, OVA, etc.).
func WithSubtype(subtype AnimeSubtype) AnimeSearchOption {
	return func(p *animeSearchParams) {
		p.subtype = subtype
	}
}

// WithSeasonYear filters search results by release year.
func WithSeasonYear(year int) AnimeSearchOption {
	return func(p *animeSearchParams) {
		p.seasonYear = year
	}
}

// WithLimit sets the maximum number of items to return in a search page (max 20).
func WithLimit(limit int) AnimeSearchOption {
	return func(p *animeSearchParams) {
		p.limit = limit
	}
}
