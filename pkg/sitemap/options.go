package sitemap

import (
	"net/http"
	"net/url"
)

// OptionFunc configures a builder
type OptionFunc func(*Builder)

// WithClient sets the client for a builder
func WithClient(client *http.Client) OptionFunc {
	return func(b *Builder) {
		b.client = client
	}
}

// WithFilter configures url filter
func WithFilter(filter FilterFunc) OptionFunc {
	return func(b *Builder) {
		b.filter = filter
	}
}

// FilterFunc filters u against its base,
// it returns url valid for sitemap and true
// or nil and false otherwise.
type FilterFunc func(base, request *url.URL, foundLink string) (*url.URL, bool)

// WithMaxDepth configures max depth
func WithMaxDepth(maxDepth int) OptionFunc {
	return func(b *Builder) {
		b.maxDepth = maxDepth
	}
}

// WithConcurrentRequests configures how many concurrent requests
// can be processed simultaneously
func WithConcurrentRequests(concurrentRequests int) OptionFunc {
	return func(b *Builder) {
		b.concurrentRequests = concurrentRequests
	}
}
