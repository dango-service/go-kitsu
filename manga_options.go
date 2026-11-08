package kitsu

type mangaSearchParams struct {
	limit   int
	offset  int
	subtype MangaSubtype
}

// MangaSearchOption configures parameters for MangaService.Search.
type MangaSearchOption func(*mangaSearchParams)

// WithMangaLimit sets the maximum number of results to return (max 20 per Kitsu API).
func WithMangaLimit(limit int) MangaSearchOption {
	return func(p *mangaSearchParams) {
		p.limit = limit
	}
}

// WithMangaOffset sets the pagination offset for search results.
func WithMangaOffset(offset int) MangaSearchOption {
	return func(p *mangaSearchParams) {
		p.offset = offset
	}
}

// WithMangaSubtype filters manga search results by format (manga, novel, manhwa, manhua, oneshot, doujin).
func WithMangaSubtype(subtype MangaSubtype) MangaSearchOption {
	return func(p *mangaSearchParams) {
		p.subtype = subtype
	}
}
