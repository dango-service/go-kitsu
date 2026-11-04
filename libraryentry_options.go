package kitsu

type collectionParams struct {
	status LibraryStatus
}

// CollectionOption configures parameters for LibraryEntryService.GetCollection.
type CollectionOption func(*collectionParams)

// WithStatus filters the collection by a specific library status (current, completed, etc.).
func WithStatus(status LibraryStatus) CollectionOption {
	return func(p *collectionParams) {
		p.status = status
	}
}
