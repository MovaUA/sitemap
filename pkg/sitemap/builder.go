// Package sitemap builds sitemap for the site.
// Built sitemap is suitable for XML encoding in Sitemaps XML format.
// See https://www.sitemaps.org/protocol.html
package sitemap

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strings"

	"github.com/movaua/link/pkg/link"
)

// Build walks all the links on the site provided by url
// and returns sitemap with links belonging to the same domain
// as of the provided url, or an error if any.
// Build uses default Builder.
func Build(rootURL string) (*URLSet, error) {
	defaultBuilder := NewBuilder()
	return defaultBuilder.Build(rootURL)
}

// Builder buils sitemap
type Builder struct {
	client             *http.Client
	filter             FilterFunc
	maxDepth           int
	concurrentRequests int
}

// NewBuilder creates a builder
func NewBuilder(opts ...OptionFunc) *Builder {
	b := &Builder{
		client:             http.DefaultClient,
		filter:             defaultFilter,
		maxDepth:           3,
		concurrentRequests: runtime.NumCPU(),
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// Build walks all the links on the site provided by url
// and returns sitemap with links belonging to the same domain
// as of the provided url, or an error if any
func (b *Builder) Build(rootURL string) (*URLSet, error) {
	root, err := url.Parse(rootURL)
	if err != nil {
		return nil, err
	}
	if root.Scheme != "http" && root.Scheme != "https" {
		return nil, fmt.Errorf("rootURL schema is not supported: %s", root.Scheme)
	}

	seen := make(map[string]struct{})

	var queue []string
	next := []string{root.String()}

	for depth := 0; depth < b.maxDepth; depth++ {
		queue, next = next, []string{}
		for _, rawurl := range queue {
			base, foundURLs, err := b.findURLs(rawurl)
			if err != nil {
				log.Printf("could not get URls from %q: %v\n", rawurl, err)
				continue
			}
			for _, found := range foundURLs {
				filtered, ok := b.filter(base, found)
				if !ok {
					continue
				}

				rawurl := filtered.String()
				if _, ok := seen[rawurl]; ok {
					continue
				}
				seen[rawurl] = struct{}{}
				next = append(next, rawurl)
			}
		}
		if len(next) == 0 {
			break
		}
	}

	urlset := URLSet{Entries: make([]URLEntry, 0, len(seen))}
	for rawurl := range seen {
		urlset.Entries = append(urlset.Entries, URLEntry{URL: rawurl})
	}
	return &urlset, nil
}

func (b *Builder) findURLs(rawurl string) (*url.URL, []*url.URL, error) {
	res, err := b.client.Get(rawurl)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	links, err := link.Find(res.Body)
	if err != nil {
		return res.Request.URL, nil, err
	}

	found := make([]*url.URL, 0, len(links))

	for _, l := range links {
		u, err := url.Parse(l.Href)
		if err != nil {
			log.Printf("could not parse URL %q: %v\n", l.Href, err)
			continue
		}
		found = append(found, u)
	}

	return res.Request.URL, found, nil
}

func defaultFilter(base, u *url.URL) (*url.URL, bool) {
	if strings.HasPrefix(u.String(), "#") {
		return nil, false
	}
	result := base.ResolveReference(u)
	if ok := (result.Scheme == "http" || result.Scheme == "https") && result.Host == base.Host; !ok {
		return nil, false
	}
	return result, true
}
